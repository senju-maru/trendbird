-- x_analytics_daily: 日次集計（user_id + date で UPSERT）
CREATE TABLE IF NOT EXISTS x_analytics_daily (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date            DATE        NOT NULL,
    impressions     INT         NOT NULL DEFAULT 0,
    likes           INT         NOT NULL DEFAULT 0,
    engagements     INT         NOT NULL DEFAULT 0,
    bookmarks       INT         NOT NULL DEFAULT 0,
    shares          INT         NOT NULL DEFAULT 0,
    new_follows     INT         NOT NULL DEFAULT 0,
    unfollows       INT         NOT NULL DEFAULT 0,
    replies         INT         NOT NULL DEFAULT 0,
    reposts         INT         NOT NULL DEFAULT 0,
    profile_visits  INT         NOT NULL DEFAULT 0,
    posts_created   INT         NOT NULL DEFAULT 0,
    video_views     INT         NOT NULL DEFAULT 0,
    media_views     INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_x_analytics_daily_user_date UNIQUE (user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_x_analytics_daily_user_date ON x_analytics_daily (user_id, date DESC);

-- x_analytics_posts: 投稿別メトリクス（user_id + post_id で UPSERT）
CREATE TABLE IF NOT EXISTS x_analytics_posts (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id          VARCHAR(64) NOT NULL,
    posted_at        TIMESTAMPTZ NOT NULL,
    post_text        TEXT        NOT NULL DEFAULT '',
    post_url         TEXT        NOT NULL DEFAULT '',
    impressions      INT         NOT NULL DEFAULT 0,
    likes            INT         NOT NULL DEFAULT 0,
    engagements      INT         NOT NULL DEFAULT 0,
    bookmarks        INT         NOT NULL DEFAULT 0,
    shares           INT         NOT NULL DEFAULT 0,
    new_follows      INT         NOT NULL DEFAULT 0,
    replies          INT         NOT NULL DEFAULT 0,
    reposts          INT         NOT NULL DEFAULT 0,
    profile_visits   INT         NOT NULL DEFAULT 0,
    detail_clicks    INT         NOT NULL DEFAULT 0,
    url_clicks       INT         NOT NULL DEFAULT 0,
    hashtag_clicks   INT         NOT NULL DEFAULT 0,
    permalink_clicks INT         NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_x_analytics_posts_user_post UNIQUE (user_id, post_id)
);

CREATE INDEX IF NOT EXISTS idx_x_analytics_posts_user_posted ON x_analytics_posts (user_id, posted_at DESC);
CREATE INDEX IF NOT EXISTS idx_x_analytics_posts_impressions ON x_analytics_posts (user_id, impressions DESC);
