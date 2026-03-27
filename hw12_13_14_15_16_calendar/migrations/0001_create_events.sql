CREATE TABLE IF NOT EXISTS events
(
    id         BIGSERIAL PRIMARY KEY,
    title      TEXT                     NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time   TIMESTAMP WITH TIME ZONE NOT NULL CHECK (end_time > start_time),
    user_id    BIGINT                   NOT NULL,


    INDEX      idx_events_user_start(user_id, start_time),
    INDEX      idx_events_start_time(start_time)
);

CREATE OR REPLACE FUNCTION check_event_overlap()
    RETURNS TRIGGER AS
$$
BEGIN
    IF EXISTS (SELECT 1
               FROM events
               WHERE user_id = NEW.user_id
                 AND (start_time, end_time) OVERLAPS (NEW.start_time, NEW.end_time)
                 AND id != COALESCE(NEW.id, 0)) THEN
        RAISE EXCEPTION 'date already busy for user %', NEW.user_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_event_overlap
    BEFORE INSERT OR UPDATE
    ON events
    FOR EACH ROW
EXECUTE FUNCTION check_event_overlap();
