import logging
import requests

from flask import current_app, session


logger = logging.getLogger(__name__)


def call_go_api(endpoint, method='GET', headers={}, **kwargs):
    """Send request to Go API and return (response_json, error_message)."""
    base_url = current_app.config['GO_API_BASE_URL']  # Import will be fixed
    url = f"{base_url}/{endpoint.lstrip('/')}"
    try:
        if method == 'GET':
            resp = requests.get(
                url, params=kwargs.get('params'), headers=headers, timeout=10
            )
        elif method == 'POST':
            resp = requests.post(
                url,
                data=kwargs.get('data'),
                json=kwargs.get('json'),
                files=kwargs.get('files'),
                headers=headers,
                timeout=3
            )
        elif method == 'DELETE':
            resp = requests.delete(url, headers=headers, timeout=10)
        else:
            return None, f'Unsupported method: {method}'

        resp.raise_for_status()
        if resp.content:
            try:
                return resp.json(), None
            except ValueError:
                logger.warning(
                    f"Non-JSON response from {url}: {resp.text[:200]}"
                )
                return resp.text, None
        return None, None

    except requests.exceptions.RequestException as e:
        logger.error(f"Error calling Go API at {url}: {e}")
        return None, str(e)


def call_go_api_auth(endpoint, method='GET', **kwargs):
    """Adds authorization token to the request."""
    token = session.get('token')
    headers = {'Authorization': f'Bearer {token}'} if token else {}
    return call_go_api(endpoint, method, headers, **kwargs)

