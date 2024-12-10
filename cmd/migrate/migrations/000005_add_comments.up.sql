CREATE TABLE IF NOT EXISTS comments(
    id bigserial PRIMARY KEY,
    post_id bigserial not NULL,
    user_id bigserial NOT NULL,
    content TEXT not NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);