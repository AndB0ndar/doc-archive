import os
import logging

from flask import Flask
from flasgger import Swagger

from app.config import config_by_name
from app.error_handlers import register_error_handlers


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def create_app(config_name='default'):
    app = Flask(__name__)
    # Load configuration
    app.config.from_object(config_by_name[config_name])

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

