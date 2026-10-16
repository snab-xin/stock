-- 登录用户与笔记。笔记正文为 Markdown；trade_date 必填，ts_code 可空且不外键。

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(64)  NOT NULL,
    password_hash VARCHAR(256) NOT NULL,
    email         VARCHAR(128) NOT NULL DEFAULT '',
    phone         VARCHAR(16)  NOT NULL DEFAULT '',
    role          VARCHAR(16)  NOT NULL DEFAULT 'user',
    status        VARCHAR(16)  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_users_username UNIQUE (username)
);

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE users IS '登录用户';
COMMENT ON COLUMN users.password_hash IS 'bcrypt 哈希，不存明文';
COMMENT ON COLUMN users.status IS 'active / locked';

CREATE TABLE IF NOT EXISTS user_sessions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_user_sessions_token UNIQUE (token_hash)
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions (user_id);

CREATE TABLE IF NOT EXISTS user_notes (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    ts_code     VARCHAR(16)  NOT NULL DEFAULT '',
    target_type VARCHAR(16)  NOT NULL DEFAULT '',
    trade_date  DATE         NOT NULL,
    summary     VARCHAR(256) NOT NULL DEFAULT '',
    content     TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_notes_user_date ON user_notes (user_id, trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_user_notes_user_code_date
    ON user_notes (user_id, ts_code, trade_date)
    WHERE ts_code <> '';

DROP TRIGGER IF EXISTS trg_user_notes_updated_at ON user_notes;
CREATE TRIGGER trg_user_notes_updated_at
    BEFORE UPDATE ON user_notes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE user_notes IS '用户笔记，正文为 Markdown';
COMMENT ON COLUMN user_notes.ts_code IS '可选标的代码，不建外键';
COMMENT ON COLUMN user_notes.target_type IS 'stock / cb / index；无代码时为空';
COMMENT ON COLUMN user_notes.trade_date IS '笔记所属交易日，必填';
COMMENT ON COLUMN user_notes.content IS 'Markdown 正文';
