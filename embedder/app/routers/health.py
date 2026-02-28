import logging

from fastapi import APIRouter, HTTPException

from app.config import settings
from app.schemas import HealthResponse


router = APIRouter(tags=["health"])

logger = logging.getLogger(__name__)


@router.get(
    "/health",
    response_model=HealthResponse,
    summary="Health check endpoint",
    description="Returns the status of the service and the names of loaded models.",
    tags=["health"],
    responses={
        200: {
            "description": "Successful response",
            "content": {
                "application/json": {
                    "example": {
                        "status": "ok",
                        "embed_model": "multi-qa-MiniLM-L6-cos-v1",
                        "rerank_model": "cross-encoder/ms-marco-MiniLM-L-6-v2",
                        "reader_model": "distilbert-base-cased-distilled-squad"
                    }
                }
            }
        }
    }
)
async def health():
    return HealthResponse(
        status="ok",
        embed_model=settings.embed_model_name,
        rerank_model=settings.rerank_model_name,
        reader_model=settings.reader_model_name,
    )

