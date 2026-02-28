FROM python:3.10-slim

# for healthcheck
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ./app ./app

ENV PORT=5001
ENV EMBED_MODEL_NAME=multi-qa-MiniLM-L6-cos-v1
ENV MAX_TEXT_LENGTH=5000

EXPOSE $PORT

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5001"]
