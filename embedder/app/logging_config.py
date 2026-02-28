import logging
import sys
from typing import Optional


def setup_logging(level: str = "INFO", log_format: Optional[str] = None):
    """
    Configuring logging for the application.
    """
    if log_format is None:
        log_format = "%(asctime)s | %(name)s | [%(levelname)s] %(message)s"

    logging.basicConfig(
        level=getattr(logging, level.upper(), logging.INFO),
        format=log_format,
        handlers=[logging.StreamHandler(sys.stdout)]
    )

