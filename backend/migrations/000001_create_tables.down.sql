-- 000001_create_tables.down.sql
-- Drop all tables in reverse dependency order

DROP TABLE IF EXISTS reply_pending_queue CASCADE;
DROP TABLE IF EXISTS reply_sent_logs    CASCADE;
DROP TABLE IF EXISTS auto_reply_rules   CASCADE;
DROP TABLE IF EXISTS topic_research     CASCADE;
DROP TABLE IF EXISTS dm_pending_queue    CASCADE;
DROP TABLE IF EXISTS dm_sent_logs        CASCADE;
DROP TABLE IF EXISTS auto_dm_rules       CASCADE;
DROP TABLE IF EXISTS user_notifications  CASCADE;
DROP TABLE IF EXISTS notifications       CASCADE;
DROP TABLE IF EXISTS activities          CASCADE;
DROP TABLE IF EXISTS posts               CASCADE;
DROP TABLE IF EXISTS generated_posts     CASCADE;
DROP TABLE IF EXISTS ai_generation_logs  CASCADE;
DROP TABLE IF EXISTS posting_tips        CASCADE;
DROP TABLE IF EXISTS spike_histories     CASCADE;
DROP TABLE IF EXISTS topic_volumes       CASCADE;
DROP TABLE IF EXISTS user_genres         CASCADE;
DROP TABLE IF EXISTS user_topics         CASCADE;
DROP TABLE IF EXISTS topics              CASCADE;
DROP TABLE IF EXISTS genres              CASCADE;
DROP TABLE IF EXISTS notification_settings CASCADE;
DROP TABLE IF EXISTS twitter_connections CASCADE;
DROP TABLE IF EXISTS users               CASCADE;

DROP EXTENSION IF EXISTS pg_trgm;
