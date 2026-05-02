CREATE TABLE books (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    google_books_id   TEXT        UNIQUE,
    title             TEXT        NOT NULL,
    authors           TEXT[]      NOT NULL DEFAULT '{}',
    genres            TEXT[]      NOT NULL DEFAULT '{}',
    isbn              TEXT,
    synopsis          TEXT,
    cover_url         TEXT,
    total_pages       INTEGER,
    source            TEXT        NOT NULL CHECK (source IN ('google_books', 'manual')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE readings (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id       UUID        NOT NULL REFERENCES books(id),
    status        TEXT        NOT NULL CHECK (status IN ('want_to_read', 'reading', 'read', 'abandoned')),
    current_page  INTEGER     NOT NULL DEFAULT 1,
    added_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, book_id)
);

CREATE INDEX readings_user_status_idx ON readings (user_id, status);
