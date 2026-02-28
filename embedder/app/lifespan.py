import logging

from fastapi import FastAPI
from contextlib import asynccontextmanager

from app.config import settings
from app.services import model_manager


logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    model_manager.load_models(settings)
    yield
    model_manager.unload_models()

