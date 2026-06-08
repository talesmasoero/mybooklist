CREATE TABLE goals (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    year         INTEGER     NOT NULL CHECK (year BETWEEN 2000 AND 2100),
    target_books INTEGER     NOT NULL CHECK (target_books BETWEEN 1 AND 1000),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, year)
);
