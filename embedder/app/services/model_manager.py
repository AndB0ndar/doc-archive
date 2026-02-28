import logging

from transformers import pipeline
from sentence_transformers import SentenceTransformer, CrossEncoder


logger = logging.getLogger(__name__)


_embed_model: SentenceTransformer = None
_rerank_model: CrossEncoder = None
_reader_pipeline = None


def get_embed_model():
    return _embed_model


def get_rerank_model():
    return _rerank_model


def get_reader_pipeline():
    return _reader_pipeline


def load_models(settings):
    global _embed_model, _rerank_model, _reader_pipeline
    logger.info("Loading models...")

    logger.info(f"Embeder model: {settings.embed_model_name}")
    _embed_model = SentenceTransformer(settings.embed_model_name)
    logger.info("Embeder loaded successfully.")

    if settings.rerank_model_name:
        logger.info(f"Reranker model: {settings.rerank_model_name}")
        _rerank_model = CrossEncoder(settings.rerank_model_name)
        logger.info("Reranker loaded successfully.")

    if settings.reader_model_name:
        logger.info(f"Reader model: {settings.reader_model_name}")
        _reader_pipeline = pipeline(
            "question-answering",
            model=settings.reader_model_name,
            tokenizer=settings.reader_model_name
        )
        logger.info("Reader loaded successfully.")


def unload_models():
    global _embed_model, _rerank_model, _reader_pipeline
    logger.info("Shutting down models...")
    _embed_model = None
    _rerank_model = None
    _reader_pipeline = None

