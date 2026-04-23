---
title: Course Project. Intelligent PDF Search
author: Student
year: 2025
tags: [project, search, pdf, go]
---

# Course Project: Intelligent PDF Search

This project is a system for uploading and searching PDF documents using both full‑text and semantic search. The backend is written in Go, the database is PostgreSQL with pg_trgm and pgvector extensions. A Python microservice (Flask) generates embeddings using the all‑MiniLM‑L6‑v2 model. The web UI is built with Flask + htmx, enabling asynchronous search result updates. All components are packaged with Docker Compose for easy deployment.
