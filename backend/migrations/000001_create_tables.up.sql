-- 000001_create_tables.up.sql
-- TrendBird: All tables, indexes, extensions, and seed data

-- =============================================================================
-- Extensions
-- =============================================================================
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- =============================================================================
-- 1. users
-- =============================================================================
CREATE TABLE IF NOT EXISTS users (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    twitter_id          VARCHAR(64)  NOT NULL UNIQUE,
    name                VARCHAR(100) NOT NULL,
    email               VARCHAR(255) NOT NULL DEFAULT '',
    image               TEXT         NOT NULL DEFAULT '',
    twitter_handle      VARCHAR(15)  NOT NULL UNIQUE,
    tutorial_completed  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- 2. twitter_connections
-- =============================================================================
CREATE TABLE IF NOT EXISTS twitter_connections (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    access_token      TEXT         NOT NULL,
    refresh_token     TEXT         NOT NULL,
    token_expires_at  TIMESTAMPTZ  NOT NULL,
    status            INTEGER      NOT NULL DEFAULT 3,
    connected_at      TIMESTAMPTZ,
    last_tested_at    TIMESTAMPTZ,
    error_message     VARCHAR(1000),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- 3. notification_settings
-- =============================================================================
CREATE TABLE IF NOT EXISTS notification_settings (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    spike_enabled  BOOLEAN     NOT NULL DEFAULT TRUE,
    rising_enabled BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- 4. genres
-- =============================================================================
CREATE TABLE IF NOT EXISTS genres (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(50)  NOT NULL UNIQUE,
    label       VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seed: 10 genres
INSERT INTO genres (slug, label, description, sort_order) VALUES
    ('technology',     'テクノロジー',       'AI・プログラミング・ガジェット関連のトレンド',                 1),
    ('business',       'ビジネス・起業',     '経営・スタートアップ・副業関連のトレンド',                     2),
    ('marketing',      'マーケティング',     'SNS運用・広告・PR・ブランディング関連のトレンド',              3),
    ('finance',        '投資・マネー',       '株式・仮想通貨・不動産投資・家計管理関連のトレンド',           4),
    ('creative',       'クリエイティブ',     'デザイン・イラスト・写真・動画制作関連のトレンド',             5),
    ('lifestyle',      'ライフスタイル',     '暮らし・ミニマリスト・生産性・旅行関連のトレンド',             6),
    ('career',         'キャリア・働き方',   '転職・フリーランス・リモートワーク関連のトレンド',             7),
    ('health-beauty',  '健康・美容',         'フィットネス・ダイエット・スキンケア関連のトレンド',           8),
    ('education',      '教育・学び',         '学習法・子育て・資格・語学関連のトレンド',                     9),
    ('entertainment',  'エンタメ',           'ゲーム・アニメ・映画・音楽関連のトレンド',                     10)
ON CONFLICT (slug) DO NOTHING;

-- =============================================================================
-- 5. topics
-- =============================================================================
CREATE TABLE IF NOT EXISTS topics (
    id               UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(100)     NOT NULL,
    keywords         JSONB            NOT NULL DEFAULT '[]',
    genre_id         UUID             NOT NULL REFERENCES genres(id),
    status           INTEGER          NOT NULL DEFAULT 3,
    change_percent   DOUBLE PRECISION NOT NULL DEFAULT 0,
    z_score          DOUBLE PRECISION,
    current_volume   INTEGER          NOT NULL DEFAULT 0,
    baseline_volume  INTEGER          NOT NULL DEFAULT 0,
    context          TEXT,
    context_summary  TEXT,
    spike_started_at TIMESTAMPTZ,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_topics_name_genre_id UNIQUE (name, genre_id)
);

CREATE INDEX IF NOT EXISTS idx_topics_name_trgm ON topics USING GIN (name gin_trgm_ops);

-- =============================================================================
-- 6. user_topics
-- =============================================================================
CREATE TABLE IF NOT EXISTS user_topics (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id             UUID        NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    notification_enabled BOOLEAN     NOT NULL DEFAULT TRUE,
    is_creator           BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_topics_user_topic UNIQUE (user_id, topic_id)
);

CREATE INDEX IF NOT EXISTS idx_user_topics_user_id  ON user_topics (user_id);
CREATE INDEX IF NOT EXISTS idx_user_topics_topic_id ON user_topics (topic_id);

-- =============================================================================
-- 7. user_genres
-- =============================================================================
CREATE TABLE IF NOT EXISTS user_genres (
    id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    genre_id  UUID        NOT NULL REFERENCES genres(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_genres_user_genre_id UNIQUE (user_id, genre_id)
);

CREATE INDEX IF NOT EXISTS idx_user_genres_user_id ON user_genres (user_id);

-- =============================================================================
-- 8. topic_volumes (immutable)
-- =============================================================================
CREATE TABLE IF NOT EXISTS topic_volumes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id   UUID        NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    timestamp  TIMESTAMPTZ NOT NULL,
    value      INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_topic_volumes_topic_timestamp UNIQUE (topic_id, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_topic_volumes_topic_timestamp ON topic_volumes (topic_id, timestamp);

-- =============================================================================
-- 9. spike_histories (immutable, except notified_at)
-- =============================================================================
CREATE TABLE IF NOT EXISTS spike_histories (
    id               UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id         UUID             NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    timestamp        TIMESTAMPTZ      NOT NULL,
    peak_z_score     DOUBLE PRECISION NOT NULL,
    status           INTEGER          NOT NULL,
    summary          VARCHAR(500)     NOT NULL,
    duration_minutes INTEGER          NOT NULL,
    notified_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_spike_histories_topic_id ON spike_histories (topic_id);
CREATE INDEX IF NOT EXISTS idx_spike_histories_notified ON spike_histories (notified_at) WHERE notified_at IS NULL;

-- =============================================================================
-- 10. posting_tips (1:1 with topic)
-- =============================================================================
CREATE TABLE IF NOT EXISTS posting_tips (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id            UUID        NOT NULL UNIQUE REFERENCES topics(id) ON DELETE CASCADE,
    peak_days           JSONB       NOT NULL DEFAULT '[]',
    peak_hours_start    INTEGER     NOT NULL DEFAULT 0,
    peak_hours_end      INTEGER     NOT NULL DEFAULT 23,
    next_suggested_time TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- 11. ai_generation_logs (immutable)
-- =============================================================================
CREATE TABLE IF NOT EXISTS ai_generation_logs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id   UUID        REFERENCES topics(id) ON DELETE SET NULL,
    style      INTEGER     NOT NULL,
    count      INTEGER     NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_generation_logs_user_created ON ai_generation_logs (user_id, created_at);

-- =============================================================================
-- 12. generated_posts (immutable)
-- =============================================================================
CREATE TABLE IF NOT EXISTS generated_posts (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id          UUID        REFERENCES topics(id) ON DELETE SET NULL,
    generation_log_id UUID        REFERENCES ai_generation_logs(id) ON DELETE SET NULL,
    style             INTEGER     NOT NULL,
    content           TEXT        NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generated_posts_user_id           ON generated_posts (user_id);
CREATE INDEX IF NOT EXISTS idx_generated_posts_generation_log_id ON generated_posts (generation_log_id) WHERE generation_log_id IS NOT NULL;

-- =============================================================================
-- 13. posts
-- =============================================================================
CREATE TABLE IF NOT EXISTS posts (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content       TEXT           NOT NULL,
    topic_id      UUID           REFERENCES topics(id) ON DELETE SET NULL,
    topic_name    VARCHAR(100),
    status        INTEGER        NOT NULL DEFAULT 1,
    scheduled_at  TIMESTAMPTZ,
    published_at  TIMESTAMPTZ,
    failed_at     TIMESTAMPTZ,
    error_message VARCHAR(1000),
    tweet_url     TEXT,
    likes         INTEGER        NOT NULL DEFAULT 0,
    retweets      INTEGER        NOT NULL DEFAULT 0,
    replies       INTEGER        NOT NULL DEFAULT 0,
    views         INTEGER        NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id_status        ON posts (user_id, status);
CREATE INDEX IF NOT EXISTS idx_posts_user_id_published_at  ON posts (user_id, published_at) WHERE published_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_posts_scheduled_at          ON posts (scheduled_at)          WHERE status = 2 AND scheduled_at IS NOT NULL;

-- =============================================================================
-- 14. activities (immutable)
-- =============================================================================
CREATE TABLE IF NOT EXISTS activities (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        INTEGER      NOT NULL,
    topic_name  VARCHAR(100) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL,
    timestamp   TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activities_user_timestamp ON activities (user_id, timestamp DESC);

-- =============================================================================
-- 15. notifications
-- =============================================================================
CREATE TABLE IF NOT EXISTS notifications (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    type          INTEGER       NOT NULL,
    title         VARCHAR(200)  NOT NULL,
    message       VARCHAR(1000) NOT NULL,
    topic_id      UUID          REFERENCES topics(id) ON DELETE SET NULL,
    topic_name    VARCHAR(100),
    topic_status  INTEGER,
    action_url    TEXT,
    action_label  VARCHAR(100),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- 16. user_notifications
-- =============================================================================
CREATE TABLE IF NOT EXISTS user_notifications (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_id UUID        NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    is_read         BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_notifications_user_notification UNIQUE (user_id, notification_id)
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_user_read_created ON user_notifications (user_id, is_read, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_notifications_notification_id   ON user_notifications (notification_id);

-- =============================================================================
-- 17. auto_dm_rules
-- =============================================================================
CREATE TABLE IF NOT EXISTS auto_dm_rules (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enabled               BOOLEAN     NOT NULL DEFAULT FALSE,
    trigger_keywords      TEXT[]      NOT NULL DEFAULT '{}',
    template_message      TEXT        NOT NULL DEFAULT '',
    last_checked_reply_id VARCHAR(64),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auto_dm_rules_user_id ON auto_dm_rules (user_id);
CREATE INDEX IF NOT EXISTS idx_auto_dm_rules_enabled ON auto_dm_rules (enabled) WHERE enabled = TRUE;

-- =============================================================================
-- 18. dm_sent_logs
-- =============================================================================
CREATE TABLE IF NOT EXISTS dm_sent_logs (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rule_id              UUID         NOT NULL REFERENCES auto_dm_rules(id) ON DELETE CASCADE,
    recipient_twitter_id VARCHAR(64)  NOT NULL,
    reply_tweet_id       VARCHAR(64)  NOT NULL,
    trigger_keyword      VARCHAR(255) NOT NULL,
    dm_text              TEXT         NOT NULL,
    sent_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_dm_sent_logs_reply UNIQUE (reply_tweet_id, recipient_twitter_id)
);

CREATE INDEX IF NOT EXISTS idx_dm_sent_logs_user_id ON dm_sent_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_dm_sent_logs_sent_at ON dm_sent_logs (user_id, sent_at DESC);

-- =============================================================================
-- 19. dm_pending_queue
-- =============================================================================
CREATE TABLE IF NOT EXISTS dm_pending_queue (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rule_id              UUID         NOT NULL REFERENCES auto_dm_rules(id) ON DELETE CASCADE,
    recipient_twitter_id VARCHAR(64)  NOT NULL,
    reply_tweet_id       VARCHAR(64)  NOT NULL,
    trigger_keyword      VARCHAR(255) NOT NULL,
    status               INT          NOT NULL DEFAULT 1,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_dm_pending_reply UNIQUE (reply_tweet_id, recipient_twitter_id)
);

CREATE INDEX IF NOT EXISTS idx_dm_pending_queue_status  ON dm_pending_queue (status) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_dm_pending_queue_user_id ON dm_pending_queue (user_id);

-- =============================================================================
-- 20. topic_research (immutable)
-- =============================================================================
CREATE TABLE IF NOT EXISTS topic_research (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id     UUID        NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    query        TEXT        NOT NULL,
    summary      TEXT        NOT NULL,
    source_urls  TEXT[]      DEFAULT '{}',
    trigger_type TEXT        NOT NULL,
    searched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_topic_research_topic_id    ON topic_research (topic_id);
CREATE INDEX IF NOT EXISTS idx_topic_research_searched_at ON topic_research (searched_at DESC);
