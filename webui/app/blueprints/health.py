from flask import Blueprint, current_app


bp = Blueprint('health', __name__)


@bp.route('/health')
def health():
    """
    Health check endpoint for the frontend service.
    ---
    tags:
      - Monitoring
    responses:
      200:
        description: Service is healthy
        content:
          application/json:
            schema:
              type: object
              properties:
                status:
                  type: string
                  example: ok
    """
    return {"status": "ok"}
