import logging
import requests

from flask import (
    Blueprint,
    Response,
    render_template,
    request,
    redirect,
    url_for,
    abort,
    current_app,
    make_response,
    send_from_directory
)

from app.api_client import call_go_api_auth
from app.decorators import login_required


bp = Blueprint('main', __name__)

logger = logging.getLogger(__name__)


@bp.route('/')
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
    return render_template('pages/index.html', documents=docs)


@bp.route('/upload', methods=['GET', 'POST'])
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
        return render_template('pages/upload.html')

    file = request.files.get('file')
    if not file:
        return 'File required', 400

    data = {
        'title': request.form.get('title', '').strip(),
        'authors': request.form.get('authors', '').strip(),
        'year': request.form.get('year', '').strip(),
        'category': request.form.get('category', '').strip(),
    }
    files = {'file': (file.filename, file.stream, file.mimetype)}

    result, err = call_go_api_auth(
        '/upload', method='POST', data=data, files=files
    )
    if err:
        logger.error(f"Upload failed: {err}")
        return f"Upload failed: {err}", 500

    return redirect(url_for('main.document', doc_id=result['document_id']))


@bp.route('/documents/<uuid:doc_id>')
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
    
    data, err = call_go_api_auth(f'/documents/{doc_id}/download-url')
    file_url = None if err or not data or 'url' not in data else data['url']
    
    return render_template('pages/document.html', doc=doc, file_url=file_url)


@bp.route('/documents/<uuid:doc_id>/file')
@login_required
def document_file(doc_id):
    """
    Proxy PDF file from MinIO to avoid PDF.js origin restrictions.
    ---
    tags:
      - Files
    parameters:
      - name: doc_id
        in: path
        type: string
        format: uuid
        required: true
        description: Unique document identifier
    responses:
      200:
        description: PDF file content
        content:
          application/pdf:
            schema:
              type: string
              format: binary
      401:
        description: Unauthorized (missing or invalid token)
      404:
        description: Document not found or file unavailable
    """
    data, err = call_go_api_auth(f'/documents/{doc_id}/download-url')
    if err or not data or 'url' not in data:
        abort(404)

    presigned_url = data['url']

    try:
        resp = requests.get(presigned_url, stream=True, timeout=30)
        resp.raise_for_status()
    except requests.RequestException:
        abort(404)

    return Response(
        resp.iter_content(chunk_size=8192),
        content_type='application/pdf',
        headers={
            'Content-Disposition': f'inline; filename="{doc_id}.pdf"'
        }
    )


@bp.route('/documents/<uuid:doc_id>/delete', methods=['DELETE'])
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
    result, err = call_go_api_auth(f'/documents/{doc_id}', method='DELETE')
    if err:
        logger.error(f"Failed to delete document {doc_id}: {err}")
        return f"Delete failed: {err}", 500
    response = make_response('', 200)
    response.headers['HX-Redirect'] = url_for('main.index')
    return response


@bp.route('/search')
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
    query = request.args.get('q', '')
    search_type = request.args.get('type', 'text')
    if not query:
        return '', 400

    results, err = call_go_api_auth(
        '/search', params={'q': query, 'type': search_type}
    )
    if err:
        logger.error(f"Search error: {err}")
        return f"Search error: {err}", 500
    answers = results["results"]

    for answer in answers:
        doc_id = answer['document_id']
        doc, err = call_go_api_auth(f"/documents/{doc_id}")
        answer['document'] = None if err or doc is None else doc

    return render_template(
        'schema/search_results.html', results=answers
    )

