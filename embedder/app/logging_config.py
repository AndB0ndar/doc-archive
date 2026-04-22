import sys
import uuid
import logging

from typing import Optional
from fastapi import FastAPI, Request
from starlette.middleware.base import BaseHTTPMiddleware

from pythonjsonlogger import jsonlogger


logger = logging.getLogger(__name__)


def setup_logging(level: str = "INFO") -> None:
    log_level = getattr(logging, level.upper(), logging.INFO)

    formatter = jsonlogger.JsonFormatter(
        fmt='%(asctime)s %(levelname)s %(name)s %(message)s %(request_id)s',
        rename_fields={
            'levelname': 'level',
            'asctime': 'timestamp'
        }
    )

    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(formatter)

    root_logger = logging.getLogger()
    for h in root_logger.handlers[:]:
        root_logger.removeHandler(h)
    root_logger.addHandler(handler)
    root_logger.setLevel(log_level)

    for name in ["uvicorn", "uvicorn.error", "uvicorn.access"]:
        uvicorn_logger = logging.getLogger(name)
        uvicorn_logger.handlers.clear()
        uvicorn_logger.addHandler(handler)
        uvicorn_logger.setLevel(log_level)

    logger.info("Logging configured with JSON format", extra={"level": level})


class RequestIDMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        request_id = request.headers.get("X-Request-Id")
        if not request_id:
            request_id = str(uuid.uuid4())

        request.state.request_id = request_id

        extra = {"request_id": request_id}
        request.state.logger = logging.LoggerAdapter(logging.getLogger(), extra)

        request.state.logger.info(
            "Request started",
            extra={
                "method": request.method,
                "path": request.url.path,
                "client": request.client.host if request.client else None,
                **extra
            }
        )

        try:
            response = await call_next(request)
        except Exception as e:
            request.state.logger.exception("Request failed", extra=extra)
            raise

        request.state.logger.info(
            "Request finished",
            extra={
                "method": request.method,
                "path": request.url.path,
                "status_code": response.status_code,
                "request_id": request_id,
            }
        )

        response.headers["X-Request-Id"] = request_id
        return response


def get_logger(request: Optional[Request] = None) -> logging.LoggerAdapter:
    if request and hasattr(request.state, "logger"):
        return request.state.logger
    return logging.getLogger()

