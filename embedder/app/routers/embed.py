import logging

from fastapi import APIRouter, HTTPException

from app.services.model_manager import get_embed_model
from app.schemas import EmbedRequest, EmbedResponse


router = APIRouter(tags=["embeddings"])

logger = logging.getLogger(__name__)


@router.post(
    "/embed",
    response_model=EmbedResponse,
    summary="Generate embeddings for a list of texts",
    description="Accepts a list of strings and returns a list of 384‑dimensional embedding vectors.",
    tags=["embeddings"],
    responses={
        200: {
            "description": "Successful response",
            "content": {
                "application/json": {
                    "example": {
                        "embeddings": [[0.123, -0.456], [0.789, -0.012]]
                    }
                }
            }
        },
        503: {"description": "Model not loaded"},
        500: {"description": "Internal error"}
    }
)
async def embed(request: EmbedRequest):
    model = get_embed_model()
    if model is None:
        raise HTTPException(status_code=503, detail="Embed model not loaded")
    try:
        max_length = int(os.getenv("MAX_TEXT_LENGTH", 5000))
        truncated = [text[:max_length] for text in request.texts]
        embeddings = model.encode(truncated).tolist()
        return EmbedResponse(embeddings=embeddings)
    except Exception as e:
        logger.error(f"Embedding error: {e}")
        raise HTTPException(status_code=500, detail="Internal error")

