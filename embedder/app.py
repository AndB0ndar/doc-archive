import os
import logging
from flask import Flask, request, jsonify
from sentence_transformers import SentenceTransformer


app = Flask(__name__)
logging.basicConfig(level=logging.INFO)


# Load model then start (cache)
MODEL_NAME = os.getenv("MODEL_NAME", "all-MiniLM-L6-v2")
model = SentenceTransformer(MODEL_NAME)


@app.route('/embed', methods=['POST'])
def embed():
    data = request.get_json()
    if not data or 'texts' not in data:
        return jsonify({'error': 'Missing texts field'}), 400
    
    texts = data['texts']
    if not isinstance(texts, list):
        return jsonify({'error': 'texts must be a list'}), 400
    
    # Limit the length of each text (DoS prevention)
    max_length = int(os.getenv("MAX_TEXT_LENGTH", 5000))
    truncated = [t[:max_length] for t in texts]
    
    # Embedding generation
    try:
        embeddings = model.encode(truncated).tolist()
        return jsonify(embeddings)
    except Exception as e:
        app.logger.error(f"Embedding error: {e}")
        return jsonify({'error': 'Internal error'}), 500


@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'ok'})


if __name__ == '__main__':
    port = int(os.getenv("PORT", 5001))
    app.run(host='0.0.0.0', port=port)

