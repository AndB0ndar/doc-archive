import os
import logging
import requests

from flask import Flask
from flask import render_template, request, redirect, url_for, abort
from flask import send_from_directory, after_this_request

from flasgger import Swagger


# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

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
def call_go_api(endpoint, method='GET', **kwargs):
    """
    Send a request to the Go backend API and return (response_json, error_message).
    """
    url = f"{app.config['GO_API_BASE_URL']}/{endpoint.lstrip('/')}"
    try:
        if method == 'GET':
            resp = requests.get(url, params=kwargs.get('params'), timeout=10)
        elif method == 'POST':
            # Files and form data are expected for uploads
            resp = requests.post(url, data=kwargs.get('data'), files=kwargs.get('files'), timeout=30)
        else:
            return None, f'Unsupported method: {method}'

        resp.raise_for_status()
        return resp.json(), None

    except requests.exceptions.RequestException as e:
        logger.error(f"Error calling Go API at {url}: {e}")
        return None, str(e)


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


@app.route('/')
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
    docs, err = call_go_api('/documents')
    if err:
        docs = []
    return render_template('index.html', documents=docs)


@app.route('/upload', methods=['GET', 'POST'])
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

    result, err = call_go_api('/upload', method='POST', data=data, files=files)
    if err:
        logger.error(f"Upload failed: {err}")
        return f"Upload failed: {err}", 500

    # On success, redirect to the document view page
    return redirect(url_for('document', doc_id=result['id']))


@app.route('/documents/<int:doc_id>')
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
    doc, err = call_go_api(f'/documents/{doc_id}')
    if err or doc is None:
        abort(404)
    return render_template('document.html', doc=doc)


@app.route('/uploads/<path:filename>')
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
    @after_this_request
    def add_cors(response):
        response.headers['Access-Control-Allow-Origin'] = '*'
        return response
    return send_from_directory(app.config['UPLOAD_DIR'], filename)


@app.route('/search')
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
    query = request.args.get('q', '')
    search_type = request.args.get('type', 'text')
    if not query:
        return '', 400

    results, err = call_go_api('/search', params={'q': query, 'type': search_type})
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
