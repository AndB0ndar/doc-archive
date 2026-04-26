import os
import redis
import secrets


class Config:
    """Base configuration."""
    SECRET_KEY = os.getenv('SECRET_KEY', secrets.token_urlsafe(16))
    GO_API_BASE_URL = os.getenv('GO_API_BASE_URL', 'http://api:8080')
    UPLOAD_DIR = os.getenv('UPLOAD_DIR', '/app/uploads')
    SWAGGER = {
        'title': 'Document Management Frontend API',
        'description': 'Endpoints served by the Flask frontend (HTML views and htmx fragments)',
        'version': '1.0.0',
        'uiversion': 3,
        'specs_route': '/docs/',
        'specs': [{
            'endpoint': 'apispec_1',
            'route': '/apispec_1.json',
            'rule_filter': lambda rule: True,
            'model_filter': lambda tag: True,
        }],
        'static_url_path': '/flasgger_static',
    }

    # Redis for session
    SESSION_TYPE = 'redis'
    SESSION_REDIS = redis.from_url(os.getenv('REDIS_URL', 'redis://redis:6379/0'))
    SESSION_PERMANENT = False
    SESSION_USE_SIGNER = True
    SESSION_KEY_PREFIX = 'flask_session:'


class DevelopmentConfig(Config):
    DEBUG = True


class ProductionConfig(Config):
    DEBUG = False
    # Additional production settings


# For simplicity, we use a default configuration; you can select via env var
config_by_name = {
    'development': DevelopmentConfig,
    'production': ProductionConfig,
    'default': DevelopmentConfig
}

