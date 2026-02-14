# Sentence Embedding Service (FastAPI)

A lightweight microservice that generates sentence embeddings
using [SentenceTransformers](https://www.sbert.net/).
Built with FastAPI for automatic OpenAPI documentation and robust validation.

## Features

- Loads a SentenceTransformer model once at startup.
- Accepts a list of texts and returns their vector embeddings.
- Configurable model name and maximum text length via environment variables.
- Interactive Swagger UI at `/docs`.
- Fully containerized with Docker.

## Requirements

- Python 3.9+
- Docker (optional, for containerized deployment)

## Installation (Local)

1. Clone the repository and navigate into it:
   ```bash
   git clone <your-repo-url>
   cd <repo-folder>
   ```

2. Create a virtual environment (recommended):
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

3. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

## Running Locally

Start the server with:

```bash
python app.py
```

Or using `uvicorn` directly:

```bash
uvicorn app:app --host 0.0.0.0 --port 5001 --reload
```

The service will be available at `http://localhost:5001`.  
API documentation is at `http://localhost:5001/docs`.

## Docker Usage

### Build the Docker Image

```bash
docker build -t embedding-service .
```

### Run the Container

```bash
docker run -p 5001:5001 \
  -e MODEL_NAME="all-MiniLM-L6-v2" \
  -e MAX_TEXT_LENGTH=5000 \
  embedding-service
```

- The `-p` flag maps the container's port 5001 to your host's port 5001.
- Environment variables can be passed with `-e` (see Configuration below).

## Configuration

The service is configured via environment variables:

| Variable           | Default             | Description                              |
|--------------------|---------------------|------------------------------------------|
| `MODEL_NAME`       | all-MiniLM-L6-v2    | SentenceTransformer model name           |
| `MAX_TEXT_LENGTH`  | 5000                | Maximum characters per input text        |
| `PORT`             | 5001                | Port the server listens on               |

## API Endpoints

### `GET /health`

Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

### `POST /embed`

Generate embeddings for a list of texts.

**Request body:**
```json
{
  "texts": ["Hello world", "FastAPI is awesome"]
}
```

**Response:**
```json
{
  "embeddings": [
    [0.1234, -0.5678, ...],
    [0.2345, -0.6789, ...]
  ]
}
```

**Example using `curl`:**
```bash
curl -X POST http://localhost:5001/embed \
  -H "Content-Type: application/json" \
  -d '{"texts": ["Hello world", "FastAPI is awesome"]}'
```

**Example using Python `requests`:**
```python
import requests

response = requests.post(
    "http://localhost:5001/embed",
    json={"texts": ["Hello world", "FastAPI is awesome"]}
)
print(response.json())
```

## Error Handling

- **400/422:** Invalid input (e.g., missing `texts` field, wrong type) – automatically returned by FastAPI with details.
- **500:** Internal server error (model failure) – logged and returns a generic message.

## License

Not defined

