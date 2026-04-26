# WebUI – Flask Frontend

This is the web interface for the **Doc Archive** intelligent PDF search system. It provides a user-friendly UI for document upload, search (full‑text and semantic), and PDF preview. The frontend communicates with the Go backend via REST API and uses **htmx** for asynchronous interactions.

## Features

- User authentication (register/login) with JWT tokens
- Document upload with metadata (title, authors, year, category)
- Full‑text and semantic search with real‑time results (debounced input)
- Embedded PDF viewer using PDF.js
- Document list and individual document pages
- Server‑side sessions stored in **Redis** (via Flask‑Session)

## Tech Stack

| Component       | Technology                                      |
|----------------|-------------------------------------------------|
| Framework      | Flask                                           |
| Templates      | Jinja2                                          |
| Frontend logic | htmx (asynchronous requests)                   |
| PDF viewer     | PDF.js (included as static)                    |
| Session store  | Redis (using `Flask-Session`)                  |
| HTTP client    | `requests` (calls Go API)                      |
| Deployment     | Docker / Docker Compose                         |

## Project Structure

```
webui/
├── app/
│   ├── __init__.py          # Flask app factory
│   ├── api_client.py        # HTTP helpers to call Go API
│   ├── blueprints/          # Route modules
│   │   ├── auth.py          # login, register
│   │   ├── health.py        # health checks
│   │   └── main.py          # index, upload, documents, search
│   ├── config.py            # Configuration classes
│   ├── decorators.py        # e.g. @login_required
│   ├── error_handlers.py    # custom error pages
│   ├── static/              # CSS, JS (style.css)
│   └── templates/           # Jinja2 templates
│       ├── base.html
│       ├── auth/
│       ├── errors/
│       ├── pages/           # index, upload, document
│       └── schema/          # search_results.html (htmx partial)
├── app.py                   # Development entry point
├── Dockerfile               # Container build
├── requirements.txt
├── wsgi.py                  # Production entry point
└── README.md
```

## Configuration

All settings are loaded from environment variables. See `app/config.py`.

| Variable              | Description                            | Default                     |
|-----------------------|----------------------------------------|-----------------------------|
| `SECRET_KEY`          | Flask secret key (required)            | (none – must be set)        |
| `GO_API_BASE_URL`     | URL of the Go backend API              | `http://api:8080`           |
| `REDIS_URL`           | Redis connection URL (for sessions)    | `redis://redis:6379/0`      |
| `FLASK_ENV`           | Environment (`development`/`production`)| `development`               |
| `DEBUG`               | Enable debug mode                      | `False`                     |

## Quick Start

### Using Docker (recommended)

The webui is part of the `docker-compose.app.yml` stack. From the project root:

```bash
docker-compose -f docker-compose.infra.yml up -d
docker-compose -f docker-compose.app.yml up -d webui
```

Access the interface at [http://localhost:5005](http://localhost:5005)

### Local development (without Docker)

1. Make sure the Go backend and Redis are running (use `docker-compose.infra.yml` to start dependencies).
2. Install Python dependencies:
   ```bash
   cd webui
   pip install -r requirements.txt
   ```
3. Set environment variables (create `.env` or export manually):
   ```bash
   export SECRET_KEY="your-secret-key"
   export GO_API_BASE_URL="http://localhost:8080"
   export REDIS_URL="redis://localhost:6379/0"
   ```
4. Run the Flask development server:
   ```bash
   flask run --port=5005
   ```

## API Integration

The frontend does **not** talk directly to the database. All data operations are done through the Go API using the helper `call_go_api()` and `call_go_api_auth()` (the latter adds the JWT token from the session).

Typical flow:
- User logs in -> JWT token stored in server‑side Redis session.
- For each user request, the token is automatically attached to outgoing API calls.
- On logout, the session is cleared.

## Session Storage

Sessions are **not** stored in cookies (only session ID).  
Instead, **Flask-Session** stores session data in Redis. This ensures that sessions become invalid when Redis is reset (e.g., after `docker-compose down -v`), and allows multiple webui replicas to share sessions.

## Environment‑specific notes

- **Development**: Debug mode enabled, auto‑reload, detailed error pages.
- **Production**: Use `wsgi.py` with a production server (e.g., Gunicorn). The Dockerfile runs Gunicorn by default.

## Customisation

- Static assets (CSS, images) go into `app/static/`.
- PDF.js is not included in this repo; you can copy it into `app/static/pdfjs/` or link a CDN. The template expects PDF.js to be available at `/static/pdfjs/web/viewer.html`.

