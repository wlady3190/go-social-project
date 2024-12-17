CREATE TABLE
    IF NOT EXISTS followers (
        user_id bigint NOT NULL,
        follower_id BIGINT NOT NULL,
        created_at TIMESTAMP(0)
        with
            TIME ZONE DEFAULT NOW (),
            --composite key para evitar duplicidades
            PRIMARY KEY (user_id, follower_id),
            Foreign Key (user_id) REFERENCES users (id) ON DELETE CASCADE,
            Foreign Key (follower_id) REFERENCES users (id) ON DELETE CASCADE
    )