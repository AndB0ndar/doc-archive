import os
import uuid
import logging

from flask import Flask, g, request
from flasgger import Swagger

from pythonjsonlogger import jsonlogger
from prometheus_flask_exporter import PrometheusMetrics

from app.config import config_by_name
from app.error_handlers import register_error_handlers


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def create_app(config_name='default'):
    app = Flask(__name__)
    # Load configuration
    app.config.from_object(config_by_name[config_name])

    setup_logging(app)
    request_id_middleware(app)

    metrics = PrometheusMetrics(app)

    # Initialize Swagger
    swagger = Swagger(app)

    # Register blueprints
    from app.blueprints import health, auth, main
    app.register_blueprint(health.bp)
    app.register_blueprint(auth.bp)
    app.register_blueprint(main.bp)

    # Register error handlers
    register_error_handlers(app)

    return app


def setup_logging(app: Flask):
    log_level = app.config.get('LOG_LEVEL', logging.INFO)
    app.logger.setLevel(log_level)
    
    for handler in app.logger.handlers[:]:
        app.logger.removeHandler(handler)
    
    formatter = jsonlogger.JsonFormatter(
        fmt='%(asctime)s %(levelname)s %(name)s %(message)s %(request_id)s',
        rename_fields={
            'levelname': 'level',
            'asctime': 'timestamp'
        }
    )
    
    handler = logging.StreamHandler()
    handler.setFormatter(formatter)
    app.logger.addHandler(handler)
    
    root_logger = logging.getLogger()
    for h in root_logger.handlers[:]:
        root_logger.removeHandler(h)
    root_logger.addHandler(handler)
    root_logger.setLevel(log_level)
    
    global logger
    logger = app.logger
    
    app.logger.info("Logging configured with JSON format")


def request_id_middleware(app: Flask):
    @app.before_request
    def before_request():
        request_id = request.headers.get('X-Request-Id')
        if not request_id:
            request_id = str(uuid.uuid4())
        g.request_id = request_id
        app.logger = logging.LoggerAdapter(
            app.logger, {'request_id': request_id}
        )
    
    @app.after_request
    def after_request(response):
        app.logger.info(
            "Request finished",
            extra={
                'method': request.method,
                'path': request.path,
                'status': response.status_code,
                'request_id': getattr(g, 'request_id', None)
            }
        )
        if hasattr(app.logger, 'log'):
            app.logger = app.logger.logger
        return response

