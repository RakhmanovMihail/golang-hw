-- +goose Up
CREATE TABLE IF NOT EXISTS notifications
(
    id         BIGSERIAL PRIMARY KEY,
    event_id   BIGINT                   NOT NULL,
    title      TEXT                     NOT NULL,
    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    user_id    BIGINT                   NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    INDEX idx_notifications_user_id(user_id),
    INDEX idx_notifications_event_id(event_id)
);

-- +goose Down
DROP TABLE IF EXISTS notifications;
