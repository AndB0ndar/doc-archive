import logging

from fastapi import APIRouter, HTTPException, Request

from app.logging_config import get_logger
from app.services.model_manager import get_rerank_model
from app.schemas import RerankRequest, RerankResponse


router = APIRouter(tags=["reranking"])


@router.post(
    "/rerank",
    response_model=RerankResponse,
    summary="Rerank a list of texts by relevance to a query",
    description="Uses a cross‑encoder model to compute relevance scores between the query and each text.",
    tags=["reranking"],
    responses={
        200: {
            "description": "Successful response",
            "content": {
                "application/json": {
                    "example": {"scores": [0.95, 0.82, 0.31]}
                }
            }
        },
        503: {"description": "Reranker not configured"},
        500: {"description": "Internal error"}
    }
)
async def rerank(request: Request, payload: RerankRequest):
    logger = get_logger(request)
    model = get_rerank_model()
    if model is None:
        raise HTTPException(status_code=503, detail="Reranker not configured")
    try:
        pairs = [[payload.query, text] for text in payload.texts]
        scores = model.predict(pairs).tolist()
        return RerankResponse(scores=scores)
    except Exception as e:
        logger.error(f"Rerank error: {e}")
        raise HTTPException(status_code=500, detail="Internal error")

