-- ============================================================
-- Migration: 001_init
-- Description: Initial schema — users, auth, sessions, files,
--              RAG (knowledge base, documents, chunks, conversations)
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- ─── Users ───────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS users (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                   VARCHAR(255) NOT NULL UNIQUE,
    fullname                VARCHAR(100) NOT NULL,
    password_hash           VARCHAR(255),
    avatar                  VARCHAR(500),
    status                  VARCHAR(20)  NOT NULL DEFAULT 'INACTIVE',
    role                    VARCHAR(20)  NOT NULL DEFAULT 'USER',
    verified_at             TIMESTAMPTZ,
    last_login              TIMESTAMPTZ,
    last_change_password_at TIMESTAMPTZ,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

-- ─── User Verifications ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS user_verifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) NOT NULL UNIQUE,
    type       VARCHAR(50)  NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_verifications_user_id ON user_verifications (user_id);

-- ─── Provider Accounts ────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS provider_accounts (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider            VARCHAR(50)  NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    access_token        TEXT         NOT NULL,
    refresh_token       TEXT,
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_provider_accounts_user_id ON provider_accounts (user_id);

-- ─── User Activity Logs ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS user_activity_logs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action     VARCHAR(100) NOT NULL,
    metadata   JSONB,
    ip_address VARCHAR(50),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_activity_logs_user_id ON user_activity_logs (user_id);

-- ─── Sessions ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sessions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,
    user_agent    VARCHAR(500),
    ip_address    VARCHAR(50),
    expires_at    TIMESTAMPTZ  NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);

-- ─── File Storage ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS file_storages (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id  UUID        NOT NULL,
    module     VARCHAR(50)  NOT NULL,
    url        VARCHAR(500) NOT NULL,
    path       VARCHAR(500) NOT NULL,
    metadata   JSONB,
    is_used    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_file_storages_target_id ON file_storages (target_id);
CREATE INDEX IF NOT EXISTS idx_file_storages_module    ON file_storages (module);

-- ─── Knowledge Bases ──────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS knowledge_bases (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    embed_model   VARCHAR(100) NOT NULL DEFAULT 'text-embedding-3-small',
    chunk_size    INT          NOT NULL DEFAULT 512,
    chunk_overlap INT          NOT NULL DEFAULT 64,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─── Documents ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS documents (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID        NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    filename          VARCHAR(255) NOT NULL,
    file_type         VARCHAR(50)  NOT NULL,
    file_size         BIGINT       NOT NULL,
    status            VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    error_message     TEXT,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_documents_knowledge_base_id ON documents (knowledge_base_id);

-- ─── Document Chunks ──────────────────────────────────────────────────────────
-- embedding uses pgvector; text-embedding-3-small = 1536 dimensions
-- IVFFlat index for approximate nearest-neighbour search (cosine similarity)

CREATE TABLE IF NOT EXISTS document_chunks (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index INT         NOT NULL,
    content     TEXT        NOT NULL,
    embedding   vector(1536),
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_document_chunks_document_id ON document_chunks (document_id);

-- Build the IVFFlat index only when the table has data; adjust lists to ~sqrt(row_count)
CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding
    ON document_chunks USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- ─── Conversations ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS conversations (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    knowledge_base_id UUID        NOT NULL REFERENCES knowledge_bases(id),
    title             VARCHAR(255) NOT NULL DEFAULT 'New Conversation',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_user_id           ON conversations (user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_knowledge_base_id ON conversations (knowledge_base_id);

-- ─── Messages ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS messages (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID        NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role            VARCHAR(20)  NOT NULL,
    content         TEXT         NOT NULL,
    tokens_used     INT          NOT NULL DEFAULT 0,
    latency_ms      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages (conversation_id);

-- ─── Message Sources ──────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS message_sources (
    id               UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id       UUID    NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    chunk_id         UUID    NOT NULL REFERENCES document_chunks(id),
    similarity_score REAL    NOT NULL,
    rank             INT     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_message_sources_message_id ON message_sources (message_id);
CREATE INDEX IF NOT EXISTS idx_message_sources_chunk_id   ON message_sources (chunk_id);
