CREATE TABLE sessions (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_id       UUID        NOT NULL REFERENCES readings(id) ON DELETE CASCADE,
    start_page       INTEGER     NOT NULL CHECK (start_page >= 1),
    end_page         INTEGER     NOT NULL CHECK (end_page >= start_page AND end_page <= 100000),
    duration_seconds INTEGER,
    session_date     DATE        NOT NULL DEFAULT CURRENT_DATE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_reading_date_idx ON sessions (reading_id, session_date DESC);
