# Managed RAG service

The service runs the pinned Persian-first `rag-v3-fa-bge-m3` pipeline and has two deployment modes from the same binary:

- `RAG_MODE=api`: management, retrieval, and published-chatbot runtime APIs.
- `RAG_MODE=worker`: durable PostgreSQL-leased ingestion with ClamAV, structured Tika extraction, Persian-capable OCR fallback, exact tokenizer chunking, embeddings, and pgvector indexing.

The API accepts management, retrieval, and chatbot runtime requests only through the Control Plane with `RAG_INTERNAL_TOKEN`. Runtime private bots additionally use separate one-hour maximum session JWTs. Public bots are limited independently by bot, IP, and session through Redis. Document text, chunks, questions, and answers are not written to application logs.

Required infrastructure is the main metadata PostgreSQL database, a dedicated PostgreSQL database with pgvector and pg_trgm, Redis, a private S3-compatible bucket, ClamAV, Apache Tika, and the existing Godar gateway for metered embedding, OCR, rerank, and generation calls. Retained conversations are encrypted with `RAG_CONVERSATION_SECRET`; zero-retention conversations exist only in Redis for 30 minutes.

`RAG_GATEWAY_RERANK_MODEL` must be the exact enabled rerank model identifier exposed by the deployment (normally a provider/account/model FQN). The profile keeps the algorithm revision while this environment mapping prevents a tenant-specific account alias from being baked into migrations.

The profile stores source-faithful display text separately from normalized Persian index text, combines dense and trigram candidates with reciprocal-rank fusion, and always reranks globally. Published chatbots pin `default_locale` and `answer_language`; the default locale is `fa-IR` and query-language matching supports Persian and English.
