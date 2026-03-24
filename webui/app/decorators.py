from functools import wraps

from flask import session, redirect, url_for


def login_required(f):
    """A decorator for checking the presence of a token in the session."""
    @wraps(f)
    def decorated_function(*args, **kwargs):
        if 'token' not in session:
            return redirect(url_for('auth.login'))  # blueprints name
        return f(*args, **kwargs)
    return decorated_function

