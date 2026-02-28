import logging

from fastapi import APIRouter, HTTPException

from app.services.model_manager import get_reader_pipeline
from app.schemas import ExtractAnswerRequest, ExtractAnswerResponse


router = APIRouter(tags=["question-answering"])

logger = logging.getLogger(__name__)


@router.post(
    "/extract_answer",
    response_model=ExtractAnswerResponse,
    summary="Extract an answer from a context passage",
    description="Uses an extractive QA model to find the answer span within the context.",
    tags=["question-answering"],
    responses={
        200: {
            "description": "Successful response",
            "content": {
                "application/json": {
                    "example": {
                        "answer": "a statically typed, compiled programming language designed at Google",
                        "confidence": 0.87,
                        "start": 18,
                        "end": 72
                    }
                }
            }
        },
        503: {"description": "Reader not configured"},
        500: {"description": "Internal error"}
    }
)
async def extract_answer(request: ExtractAnswerRequest):
    pipeline = get_reader_pipeline()
    if pipeline is None:
        raise HTTPException(status_code=503, detail="Reader not configured")
    try:
        result = pipeline(question=request.question, context=request.context)
        return ExtractAnswerResponse(
            answer=result["answer"],
            confidence=result["score"],
            start=result["start"],
            end=result["end"]
        )
    except Exception as e:
        logger.error(f"Reader error: {e}")
        raise HTTPException(status_code=500, detail="Internal error")

