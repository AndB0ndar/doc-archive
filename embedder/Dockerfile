FROM python:3.13-slim

# for healthcheck
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ./app ./app

ENV PORT=5000
ENV EMBED_MODEL_NAME="all-MiniLM-L6-v2"
ENV RERANK_MODEL_NAME="cross-encoder/ms-marco-MiniLM-L-6-v2"
ENV READER_MODEL_NAME="distilbert-base-cased-distilled-squad"
ENV MAX_TEXT_LENGTH=5000

EXPOSE $PORT

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000"]
