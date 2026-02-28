# Sentence Embedding & QA Service (FastAPI)

A microservice that provides **text embeddings**, **cross-encoder reranking**,
and **extractive question answering** using state-of-the-art transformer models.  
Built with FastAPI for automatic OpenAPI documentation,
input validation, and high performance.

---

## Features

- **Embeddings** – Generate dense vector representations of texts (384‑dim by default).  
- **Reranking** – Re‑score a list of candidate texts against a query using a cross‑encoder (optional).  
- **Question Answering** – Extract exact answer spans from a context passage (optional).  
- **Modular design** – Each model can be enabled/disabled via environment variables.  
- **Interactive Swagger UI** at `/docs` and ReDoc at `/redoc`.  
- **Fully containerized** with Docker – easy deployment.

---

## Requirements

- Python 3.9+ (if running locally)  
- Docker (optional, for containerized deployment)  
- At least 2 GB RAM for the default models (more if using larger ones)

---

## Installation (Local)

1. Clone the repository:
   ```bash
   git clone <your-repo-url>
   cd <repo-folder>/embedder_fastapi
   ```

2. Create and activate a virtual environment (recommended):
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

3. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

---

## Running Locally

Start the server with Uvicorn:

```bash
uvicorn app.main:app --host 0.0.0.0 --port 5001 --reload
```

The service will be available at `http://localhost:5001`.  
API documentation: [http://localhost:5001/docs](http://localhost:5001/docs)

---

## Docker Usage

### Build the Docker Image

From the `embedder_fastapi` directory:

```bash
docker build -t embedding-service .
```

### Run the Container

```bash
docker run -p 5001:5001 \
  -e EMBED_MODEL_NAME="multi-qa-MiniLM-L6-cos-v1" \
  -e RERANK_MODEL_NAME="cross-encoder/ms-marco-MiniLM-L-6-v2" \
  -e READER_MODEL_NAME="distilbert-base-cased-distilled-squad" \
  -e MAX_TEXT_LENGTH=5000 \
  embedding-service
```

- `-p 5001:5001` maps the container's port 5001 to your host.  
- Environment variables (see [Configuration](#configuration)) can be passed with `-e`.  
- Omit `RERANK_MODEL_NAME` or `READER_MODEL_NAME` to disable those features.

---

## Configuration

The service is configured through environment variables:

| Variable              | Default                         | Description                                         |
|-----------------------|---------------------------------|-----------------------------------------------------|
| `EMBED_MODEL_NAME`    | `multi-qa-MiniLM-L6-cos-v1`     | SentenceTransformer model for embeddings.           |
| `RERANK_MODEL_NAME`   | (not set)                       | Cross‑encoder model for reranking (e.g., `cross-encoder/ms-marco-MiniLM-L-6-v2`). If not set, `/rerank` will be unavailable. |
| `READER_MODEL_NAME`   | (not set)                       | Hugging Face extractive QA model (e.g., `distilbert-base-cased-distilled-squad`). If not set, `/extract_answer` will be unavailable. |
| `MAX_TEXT_LENGTH`     | `5000`                          | Maximum number of characters per input text (truncated to this limit). |
| `PORT`                | `5001`                          | Port the server listens on.                         |

---

## API Endpoints

All endpoints accept and return JSON.  
Interactive documentation is available at `/docs` and `/redoc`.

### Health Check

**`GET /health`**

Returns the service status and the names of loaded models.

**Response (example):**
```json
{
  "status": "ok",
  "embed_model": "multi-qa-MiniLM-L6-cos-v1",
  "rerank_model": "cross-encoder/ms-marco-MiniLM-L-6-v2",
  "reader_model": "distilbert-base-cased-distilled-squad"
}
```

---

### Embeddings

**`POST /embed`**

Generate dense vector embeddings for a list of texts.

**Request body:**
```json
{
  "texts": ["What is Go?", "Explain concurrency in Go."]
}
```

**Response:**
```json
{
  "embeddings": [
    [0.123, -0.456, ...],
    [0.789, -0.012, ...]
  ]
}
```

**Example with `curl`:**
```bash
curl -X POST http://localhost:5001/embed \
  -H "Content-Type: application/json" \
  -d '{"texts": ["What is Go?", "Explain concurrency in Go."]}'
```

---

### Reranking

**`POST /rerank`**  
*(only available if `RERANK_MODEL_NAME` is set)*

Re‑scores a list of candidate texts against a query.

**Request body:**
```json
{
  "query": "What is Go?",
  "texts": [
    "Go is a programming language created at Google.",
    "Concurrency is handled with goroutines.",
    "Python is also popular."
  ]
}
```

**Response:**
```json
{
  "scores": [0.95, 0.82, 0.31]
}
```

Higher scores indicate greater relevance to the query.

---

### Extractive Question Answering

**`POST /extract_answer`**  
*(only available if `READER_MODEL_NAME` is set)*

Extracts an exact answer span from a context passage.

**Request body:**
```json
{
  "question": "What is Go?",
  "context": "Go (often referred to as Golang) is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. It is known for its simplicity and efficient concurrency mechanisms."
}
```

**Response:**
```json
{
  "answer": "a statically typed, compiled programming language designed at Google",
  "confidence": 0.87,
  "start": 18,
  "end": 72
}
```

- `start` and `end` are character indices within the context (0‑based, inclusive start, exclusive end).

---

## Project Structure

```
./
├── app/
│   ├── __init__.py
│   ├── main.py              # FastAPI app creation, lifespan, router registration
│   ├── config.py            # Environment variable settings
│   ├── models.py            # Pydantic request/response models
│   ├── dependencies.py      # Global model instances and accessors
│   └── routers/
│       ├── __init__.py
│       ├── embed.py         # /embed endpoint
│       ├── rerank.py        # /rerank endpoint
│       ├── answer.py        # /extract_answer endpoint
│       └── health.py        # /health endpoint
├── requirements.txt
├── README.md
└── Dockerfile
```

---

## Error Handling

- **400 / 422:** Invalid input – FastAPI automatically returns detailed validation errors.  
- **503:** Requested model is not loaded (feature disabled).  
- **500:** Internal server error – logged, client receives a generic message.

---

## License

[MIT License](LICENSE)

