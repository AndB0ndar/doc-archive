import os
import logging
from typing import List

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from sentence_transformers import SentenceTransformer
from pydantic_settings import BaseSettings

# ------------------------------------------------------------------
# Configuration (environment variables with defaults)
# ------------------------------------------------------------------
class Settings(BaseSettings):
    model_name: str = Field("all-MiniLM-L6-v2", env="MODEL_NAME")
    max_text_length: int = Field(5000, env="MAX_TEXT_LENGTH")
    port: int = Field(5001, env="PORT")

    class Config:
        env_file = ".env"  # optionally load from .env file

settings = Settings()

# ------------------------------------------------------------------
# Logging
# ------------------------------------------------------------------
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ------------------------------------------------------------------
# FastAPI app with metadata for Swagger UI
# ------------------------------------------------------------------
app = FastAPI(
    title="Embedding Service",
    description="Generate sentence embeddings using SentenceTransformers",
    version="1.0.0",
    docs_url="/docs",          # Swagger UI
    redoc_url="/redoc",        # ReDoc alternative
)

# Global model variable (loaded at startup)
model: SentenceTransformer = None


# ------------------------------------------------------------------
# Pydantic models (request/response)
# ------------------------------------------------------------------
class EmbedRequest(BaseModel):
    texts: List[str] = Field(
        ..., description="List of input texts to embed",
        example=["Hello world", "FastAPI is awesome"]
    )

class EmbedResponse(BaseModel):
    embeddings: List[List[float]] = Field(
        ..., description="List of embedding vectors, one per input text"
    )


# ------------------------------------------------------------------
# Startup event: load the model once
# ------------------------------------------------------------------
@app.on_event("startup")
def load_model():
    """Load the SentenceTransformer model when the application starts."""
    global model
    logger.info(f"Loading model: {settings.model_name}")
    model = SentenceTransformer(settings.model_name)
    logger.info("Model loaded successfully.")


# ------------------------------------------------------------------
# Health check endpoint
# ------------------------------------------------------------------
@app.get("/health", summary="Health Check", tags=["Monitoring"])
async def health_check():
    """Return service health status."""
    return {"status": "ok"}


# ------------------------------------------------------------------
# Embedding endpoint
# ------------------------------------------------------------------
@app.post(
    "/embed",
    response_model=EmbedResponse,
    summary="Generate Embeddings",
    tags=["Embeddings"],
    responses={
        200: {"description": "Successful response with embeddings"},
        400: {"description": "Invalid input (handled automatically by Pydantic)"},
        500: {"description": "Internal server error (model failure)"},
    }
)
def create_embeddings(request: EmbedRequest):
    """
    Generate embeddings for a list of texts.

    - Each text is truncated to `MAX_TEXT_LENGTH` characters (configurable).
    - Returns a list of embedding vectors as floats.
    - The model is loaded once at startup and reused for all requests.
    """
    if model is None:
        logger.error("Model not loaded")
        raise HTTPException(status_code=500, detail="Model not available")

    texts = request.texts

    # Truncate texts to prevent excessively long inputs
    truncated_texts = [text[:settings.max_text_length] for text in texts]

    try:
        # Generate embeddings (synchronous call, executed in thread pool)
        embeddings = model.encode(truncated_texts).tolist()
        return EmbedResponse(embeddings=embeddings)
    except Exception as e:
        logger.error(f"Embedding error: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


# ------------------------------------------------------------------
# Run with uvicorn when script executed directly
# ------------------------------------------------------------------
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=settings.port,
        reload=False,           # set to True for development
        log_level="info"
    )
