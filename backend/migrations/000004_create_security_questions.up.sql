CREATE TABLE security_questions (
    id         SERIAL PRIMARY KEY,
    text       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_security_answers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id INTEGER     NOT NULL REFERENCES security_questions(id),
    answer_hash TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, question_id)
);

CREATE TABLE password_reset_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO security_questions (text) VALUES
    ('Qual era o nome da sua primeira escola?'),
    ('Qual o nome do seu primeiro animal de estimação?'),
    ('Em qual cidade você nasceu?'),
    ('Qual o nome do meio da sua mãe?'),
    ('Qual era o nome do seu melhor amigo de infância?'),
    ('Qual o nome da rua onde você morava na infância?'),
    ('Qual o título do primeiro livro que leu?'),
    ('Qual o nome do seu irmão ou irmã mais velho(a)?'),
    ('Qual o nome do seu professor(a) favorito(a) da escola?'),
    ('Qual era a marca do seu primeiro carro ou da família?'),
    ('Qual o nome do hospital onde você nasceu?'),
    ('Para qual time você torce?');
