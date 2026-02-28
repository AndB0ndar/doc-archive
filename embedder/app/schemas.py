from typing import List, Optional
from pydantic import BaseModel, Field


class EmbedRequest(BaseModel):
    texts: List[str] = Field(
        ...,
        description="List of texts to embed",
        example=["What is Go?", "Explain concurrency in Go."]
    )


class EmbedResponse(BaseModel):
    embeddings: List[List[float]] = Field(
        ...,
        description="List of embedding vectors (each of dimension 384)",
        example=[[0.123, -0.456, 0.789], [0.234, -0.567, 0.890]]
    )
    # Note: FastAPI will truncate long examples; you can use a shorter example.


class RerankRequest(BaseModel):
    query: str = Field(..., description="Search query", example="What is Go?")
    texts: List[str] = Field(
        ...,
        description="List of candidate texts to rerank",
        example=[
            "Go is a programming language created at Google.",
            "Concurrency is handled with goroutines.",
            "Python is also popular."
        ]
    )


class RerankResponse(BaseModel):
    scores: List[float] = Field(
        ...,
        description="Relevance scores for each candidate text",
        example=[0.95, 0.82, 0.31]
    )


class ExtractAnswerRequest(BaseModel):
    question: str = Field(..., description="Question to answer", example="What is Go?")
    context: str = Field(
        ...,
        description="Context passage containing the answer",
        example="Go (often referred to as Golang) is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. It is known for its simplicity and efficient concurrency mechanisms."
    )


class ExtractAnswerResponse(BaseModel):
    answer: str = Field(..., description="Extracted answer text", example="a statically typed, compiled programming language designed at Google")
    confidence: float = Field(..., description="Confidence score (0-1)", example=0.87)
    start: int = Field(..., description="Start character index of answer in context", example=18)
    end: int = Field(..., description="End character index of answer in context", example=72)


class HealthResponse(BaseModel):
    status: str = Field(..., description="Service status", example="ok")
    embed_model: str = Field(..., description="Name of the loaded embedding model", example="multi-qa-MiniLM-L6-cos-v1")
    rerank_model: Optional[str] = Field(None, description="Name of the loaded reranker model, if any", example="cross-encoder/ms-marco-MiniLM-L-6-v2")
    reader_model: Optional[str] = Field(None, description="Name of the loaded reader model, if any", example="distilbert-base-cased-distilled-squad")
