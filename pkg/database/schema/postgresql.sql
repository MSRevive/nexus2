CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    revision   INTEGER NOT NULL DEFAULT 0,
    flags      INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS characters (
    id              UUID PRIMARY KEY,
    steam_id        TEXT REFERENCES users(id),
    slot            INTEGER,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    data_created_at TIMESTAMPTZ,
    data_size       INTEGER NOT NULL DEFAULT 0,
    data_payload    TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS deleted_characters (
    steam_id     TEXT NOT NULL REFERENCES users(id),
    slot         INTEGER NOT NULL,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    deleted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (steam_id, slot),
    UNIQUE (character_id)
);

CREATE TABLE IF NOT EXISTS character_versions (
    id           SERIAL PRIMARY KEY,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL,
    size         INTEGER NOT NULL,
    data_payload TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chars_steam_id   ON characters(steam_id);
CREATE INDEX IF NOT EXISTS idx_charver_char_id  ON character_versions(character_id);