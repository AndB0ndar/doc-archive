import random

from flask import Flask, request, jsonify


app = Flask(__name__)


@app.route('/embed', methods=['POST'])
def embed():
    texts = request.json['texts']
    # TODO
    # so far, the random vector
    vec = [random.uniform(-1, 1) for _ in range(384)]
    return jsonify([vec])


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001)

