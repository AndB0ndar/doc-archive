import logging
import requests

from flask import (
    Blueprint, render_template, request, redirect, url_for, flash, session
)

from app.api_client import call_go_api


bp = Blueprint('auth', __name__, url_prefix='/auth')

logger = logging.getLogger(__name__)


@bp.route('/login', methods=['GET', 'POST'])
def login():
    """
    Handle user login.
    ---
    tags:
      - Authentication
    parameters:
      - name: email
        in: formData
        type: string
        required: true
        description: User's email address
      - name: password
        in: formData
        type: string
        required: true
        description: User's password
    responses:
      200:
        description: Renders login.html (GET request)
      302:
        description: Redirect to index after successful login (POST)
      401:
        description: Invalid email or password
      500:
        description: Connection error with authentication server
    """
    if request.method == 'POST':
        data = {
            'email': request.form['email'],
            'password': request.form['password']
        }
        resp_json, err = call_go_api('/login', method='POST', json=data)
        if err:
            logger.error(f"Login error: {err}")
            flash("Server connection error")
        elif resp_json and 'token' in resp_json:
            session['token'] = resp_json['token']
            session['user'] = resp_json.get('user')
            return redirect(url_for('main.index'))
        else:
            logger.error("Invalid email or password")
            flash("Invalid email or password")
    return render_template('auth/login.html')


@bp.route('/register', methods=['GET', 'POST'])
def register():
    """
    Handle user registration.
    ---
    tags:
      - Authentication
    parameters:
      - name: email
        in: formData
        type: string
        required: true
        description: User's email address
      - name: password
        in: formData
        type: string
        required: true
        description: User's password
    responses:
      200:
        description: Renders register.html (GET request)
      302:
        description: Redirect to login after successful registration (POST)
      400:
        description: Registration error (e.g., email already taken)
      500:
        description: Connection error with registration server
    """
    if request.method == 'POST':
        data = {
            'email': request.form['email'],
            'password': request.form['password']
        }
        resp_json, err = call_go_api('/register', method='POST', json=data)
        if err:
            logger.error(f"Registration error: {err}")
            flash("Server connection error")
        elif resp_json and 'id' in resp_json:  # adjust based on actual response
            logger.info("Registration successful")
            return redirect(url_for('auth.login'))
        else:
            error_msg = (
                resp_json.get('error', 'Registration failed')
                if resp_json else 'Registration failed'
            )
            logger.error(f"Register error: {error_msg}")
            flash(error_msg)
    return render_template('auth/register.html')


@bp.route('/logout')
def logout():
    """
    Log out the current user.
    ---
    tags:
      - Authentication
    responses:
      302:
        description: Redirect to login page after clearing session
    """
    session.clear()
    return redirect(url_for('auth.login'))

