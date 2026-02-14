DROP EXTENSION IF EXISTS vector;
DROP EXTENSION IF EXISTS pg_trgm;

DROP TABLE IF EXISTS users;
DROP INDEX IF EXISTS idx_documents_user_id;

DROP TABLE IF EXISTS documents;

DROP TABLE IF EXISTS chunks;
DROP INDEX IF EXISTS idx_chunks_document_id;
DROP INDEX IF EXISTS idx_chunks_content_trgm;
DROP INDEX IF EXISTS idx_chunks_embedding;
