--https://www.postgresql.org/docs/9.1/pgtrgm.html
CREATE EXTENSION IF NOT EXISTS pg_trgm;

--* gin_trgm_ops se aplica solo para texto
CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin (content gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_post_title ON posts USING gin (title gin_trgm_ops);

--* aqui es solo para arrays, sin gin_trgm_ops
CREATE INDEX IF NOT EXISTS idx_post_tags ON posts USING gin (tags);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

CREATE INDEX IF NOT EXISTS idx_post_user_id ON posts (user_id);

CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (post_id);