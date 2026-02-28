import logging

from fastapi import FastAPI

from app.config import settings
from app.lifespan import lifespan
from app.logging_config import setup_logging
from app.routers import embed, rerank, qa, health


setup_logging(level=settings.log_level)


app = FastAPI(
    title="PDF Search Embedder Service",
    description="Microservice for generating text embeddings, reranking search results, and extracting answers using transformer models.",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",          # Swagger UI (default)
    redoc_url="/redoc",        # ReDoc documentation (default)
    openapi_url="/openapi.json" # OpenAPI schema (default)
)


app.include_router(embed)
app.include_router(rerank)
app.include_router(qa)
app.include_router(health)


@app.get("/", include_in_schema=False)
async def root():
    return {"message": "PDF Search Embedder Service", "docs": "/docs"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=False,
        log_level="info"
    )

