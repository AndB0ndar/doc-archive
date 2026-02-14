import os
#import redis
import secrets
import logging
import requests

from flask import Flask
from flask import render_template, request, redirect, url_for, abort, session
from flask import flash, make_response, send_from_directory, after_this_request

#from flask_session import Session

from flasgger import Swagger


# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


app = Flask(__name__)
app.secret_key = os.getenv('SECRET_KEY', secrets.token_urlsafe(16))


# Setting Redis for store session
#app.config['SESSION_TYPE'] = 'redis'
#app.config['SESSION_PERMANENT'] = False
#app.config['SESSION_USE_SIGNER'] = True
#app.config['SESSION_REDIS'] = redis.from_url(os.getenv('REDIS_URL', 'redis://redis:6379/0'))
#Session(app)


# Configuration from environment variables
app.config['GO_API_BASE_URL'] = os.getenv('GO_API_BASE_URL', 'http://api:8080')
app.config['UPLOAD_DIR'] = os.getenv('UPLOAD_DIR', '/app/uploads')


# Swagger configuration
app.config['SWAGGER'] = {
    'title': 'Document Management Frontend API',
    'description': 'Endpoints served by the Flask frontend (HTML views and htmx fragments)',
    'version': '1.0.0',
    'uiversion': 3,
    'specs_route': '/docs/',
    'specs': [
        {
            'endpoint': 'apispec_1',
            'route': '/apispec_1.json',
            'rule_filter': lambda rule: True,  # include all routes
            'model_filter': lambda tag: True,  # include all models
        }
    ],
    'static_url_path': '/flasgger_static',
}
swagger = Swagger(app)


# ----------------------------------------------------------------------
# Utility: call the Go backend API
# ----------------------------------------------------------------------
def call_go_api(endpoint, method='GET', headers={}, **kwargs):
    """
    Send a request to the Go backend API and return (response_json, error_message).
    """
    url = f"{app.config['GO_API_BASE_URL']}/{endpoint.lstrip('/')}"
    try:
        if method == 'GET':
            resp = requests.get(
                url,
                params=kwargs.get('params'),
                headers=headers,
                timeout=10
            )
        elif method == 'POST':
            resp = requests.post(
                url,
                data=kwargs.get('data'),
                files=kwargs.get('files'), headers=headers,
                timeout=3
            )
        elif method == 'DELETE':
            resp = requests.delete(url, headers=headers, timeout=10)
        else:
            return None, f'Unsupported method: {method}'

        resp.raise_for_status()

        if resp.content:
            try:
                return resp.json(), None
            except ValueError:
                logger.warning(
                    f"Non-JSON response (status {resp.status_code})"
                    f" from {url}: {resp.text[:200]}"
                )
                return resp.text, None
        else:
            return None, None

    except requests.exceptions.RequestException as e:
        logger.error(f"Error calling Go API at {url}: {e}")
        return None, str(e)


def call_go_api_auth(endpoint, method='GET', **kwargs):
    """As call_go_api_auth. Adds an authorization token to the request."""
    token = session.get('token')
    headers = {'Authorization': f'Bearer {token}'} if token else {}
    return call_go_api(endpoint, method, headers, **kwargs)


def login_required(f):
    """A decorator for checking the presence of a token in the session."""
    from functools import wraps
    @wraps(f)
    def decorated_function(*args, **kwargs):
        if 'token' not in session:
            flash('Пожалуйста, войдите в систему')
            return redirect(url_for('login'))
        return f(*args, **kwargs)
    return decorated_function


# ----------------------------------------------------------------------
# Routes
# ----------------------------------------------------------------------
@app.route('/health')
def health():
    """
    Health check endpoint for the frontend service.
    ---
    tags:
      - Monitoring
    responses:
      200:
        description: Service is healthy
        content:
          application/json:
            schema:
              type: object
              properties:
                status:
                  type: string
                  example: ok
    """
    return {"status": "ok"}


@app.route('/login', methods=['GET', 'POST'])
def login():
    """
    Handle user login.
    ---
    tags:
      - Authentication
    parameters:
      - name: email
        in: formData
        type: string
        required: true
        description: User's email address
      - name: password
        in: formData
        type: string
        required: true
        description: User's password
    responses:
      200:
        description: Renders login.html (GET request)
      302:
        description: Redirect to index after successful login (POST)
      401:
        description: Invalid email or password
      500:
        description: Connection error with authentication server
    """
    if request.method == 'POST':
        data = {'email': request.form['email'], 'password': request.form['password']}
        try:
            resp = requests.post(f"{app.config['GO_API_BASE_URL']}/login", json=data)
            if resp.status_code == 200:
                session['token'] = resp.json()['token']
                session['user'] = resp.json()['user']
                #session.permanent = True
                return redirect(url_for('index'))
            else:
                logger.error("Invalid email or password")
                flash("Invalid email or password")
        except:
            logger.error("Server connection error")
            flash("Server connection error")
    return render_template('login.html')


@app.route('/register', methods=['GET', 'POST'])
def register():
    """
    Handle user registration.
    ---
    tags:
      - Authentication
    parameters:
      - name: email
        in: formData
        type: string
        required: true
        description: User's email address
      - name: password
        in: formData
        type: string
        required: true
        description: User's password
    responses:
      200:
        description: Renders register.html (GET request)
      302:
        description: Redirect to login after successful registration (POST)
      400:
        description: Registration error (e.g., email already taken)
      500:
        description: Connection error with registration server
    """
    if request.method == 'POST':
        data = {'email': request.form['email'], 'password': request.form['password']}
        try:
            resp = requests.post(f"{app.config['GO_API_BASE_URL']}/register", json=data)
            if resp.status_code == 200:
                logger.info("Registration is successful, log in")
                flash("Registration is successful, log in")
                return redirect(url_for('login'))
            else:
                logger.error(f"Register error: {resp.text}")
                flash(f"Register error: {resp.text}")
        except:
            logger.error("Server connection error")
            flash("Server connection error")
    return render_template('register.html')


@app.route('/logout')
def logout():
    """
    Log out the current user.
    ---
    tags:
      - Authentication
    responses:
      302:
        description: Redirect to login page after clearing session
    """
    session.clear()
    return redirect(url_for('login'))


@app.route('/')
@login_required
def index():
    """
    Home page with search form.
    ---
    tags:
      - Views
    responses:
      200:
        description: Renders index.html
    """
    docs, err = call_go_api_auth('/documents')
    if err:
        docs = []
    return render_template('index.html', documents=docs)


@app.route('/upload', methods=['GET', 'POST'])
@login_required
def upload():
    """
    Upload a new PDF document with metadata.
    ---
    tags:
      - Views
      - Upload
    parameters:
      - name: file
        in: formData
        type: file
        required: true
        description: The PDF file to upload
      - name: title
        in: formData
        type: string
        required: false
        description: Document title
      - name: authors
        in: formData
        type: string
        required: false
        description: Author(s) of the document
      - name: year
        in: formData
        type: string
        required: false
        description: Publication year
      - name: category
        in: formData
        type: string
        required: false
        description: Document category
    responses:
      200:
        description: Renders upload form (GET)
      302:
        description: Redirect to the newly created document page (POST)
      400:
        description: Missing file
      500:
        description: Backend API error
    """
    if request.method == 'GET':
        return render_template('upload.html')

    # POST: process file upload
    file = request.files.get('file')
    if not file:
        return 'File required', 400

    # Prepare metadata and file for the Go API
    data = {
        'title': request.form.get('title', '').strip(),
        'authors': request.form.get('authors', '').strip(),
        'year': request.form.get('year', '').strip(),
        'category': request.form.get('category', '').strip(),
    }
    files = {'file': (file.filename, file.stream, file.mimetype)}

    result, err = call_go_api_auth('/upload', method='POST', data=data, files=files)
    if err:
        logger.error(f"Upload failed: {err}")
        return f"Upload failed: {err}", 500

    # On success, redirect to the document view page
    return redirect(url_for('document', doc_id=result['id']))


@app.route('/documents/<int:doc_id>')
@login_required
def document(doc_id):
    """
    Display a single document with its metadata and embedded PDF.
    ---
    tags:
      - Views
    parameters:
      - name: doc_id
        in: path
        type: integer
        required: true
        description: Unique document identifier
    responses:
      200:
        description: Renders document.html
      404:
        description: Document not found
    """
    doc, err = call_go_api_auth(f'/documents/{doc_id}')
    if err or doc is None:
        abort(404)
    return render_template('document.html', doc=doc)


@app.route('/documents/<int:doc_id>/delete', methods=['DELETE'])
@login_required
def delete_document(doc_id):
    """
    Delete a document via Go API.
    ---
    tags:
      - Delete
    parameters:
      - name: doc_id
        in: path
        type: integer
        required: true
        description: Unique document identifier
    responses:
      302:
        description: Redirect to index after successful deletion
      500:
        description: Deletion failed due to API error
    """
    logger.debug(f"Session in DELETE : {dict(session)}")

    result, err = call_go_api_auth(f'/documents/{doc_id}', method='DELETE')
    if err:
        logger.error(f"Failed to delete document {doc_id}: {err}")
        return f"Delete failed: {err}", 500
    response = make_response('', 200)
    response.headers['HX-Redirect'] = url_for('index')
    return response


@app.route('/uploads/<path:filename>')
@login_required
def uploaded_file(filename):
    """
    Serve uploaded PDF files for embedding (e.g., in PDF.js).
    ---
    tags:
      - Files
    parameters:
      - name: filename
        in: path
        type: string
        required: true
        description: Name of the file in the upload folder
    responses:
      200:
        description: The requested file
        content:
          application/pdf:
            schema:
              type: string
              format: binary
      404:
        description: File not found
    """
    return send_from_directory(app.config['UPLOAD_DIR'], filename)


@app.route('/search')
@login_required
def search():
    """
    htmx endpoint: returns an HTML fragment with search results.
    ---
    tags:
      - htmx
    parameters:
      - name: q
        in: query
        type: string
        required: true
        description: Search query string
      - name: type
        in: query
        type: string
        required: false
        default: text
        description: Type of search (e.g., text, title, author)
    responses:
      200:
        description: HTML fragment containing search results
        content:
          text/html:
            schema:
              type: string
      400:
        description: Missing query parameter
      500:
        description: Backend API error
    """
    logger.debug(f"Session in SEARCH: {dict(session)}")

    query = request.args.get('q', '')
    search_type = request.args.get('type', 'text')
    if not query:
        return '', 400

    results, err = call_go_api_auth('/search', params={'q': query, 'type': search_type})
    if err:
        logger.error(f"Search error: {err}")
        return f"Search error: {err}", 500

    return render_template('search_results.html', results=results)


# ----------------------------------------------------------------------
# Error handlers
# ----------------------------------------------------------------------
@app.errorhandler(404)
def page_not_found(e):
    """Custom 404 page."""
    return render_template('404.html'), 404


@app.errorhandler(500)
def internal_server_error(e):
    """Custom 500 page."""
    return render_template('500.html'), 500


if __name__ == '__main__':
    # Run the Flask development server
    app.run(host='0.0.0.0', port=5000, debug=True)
