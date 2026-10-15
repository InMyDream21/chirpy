-- +goose Up
CREATE TABLE chirps (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    body TEXT NOT NULL,
    user_id uuid NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;
