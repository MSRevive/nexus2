CREATE TABLE IF NOT EXISTS users (
    id         TEXT PRIMARY KEY,
    revision   INTEGER NOT NULL DEFAULT 0,
    flags      INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS characters (
    id              TEXT PRIMARY KEY,
    steam_id        TEXT REFERENCES users(id),
    slot            INTEGER,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      DATETIME,
    expires_at      DATETIME,      -- populated on soft-delete for GC
    data_created_at DATETIME,
    data_size       INTEGER NOT NULL DEFAULT 0,
    data_payload    TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS deleted_characters (
    steam_id     TEXT NOT NULL REFERENCES users(id),
    slot         INTEGER NOT NULL,
    character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    deleted_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (steam_id, slot)
    UNIQUE (character_id)
);

-- Stores the version history (Versions []CharacterData on the schema struct).
-- Ordered by autoincrement id to preserve insertion order.
CREATE TABLE IF NOT EXISTS character_versions (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    created_at   DATETIME NOT NULL,
    size         INTEGER NOT NULL,
    data_payload TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chars_steam_id   ON characters(steam_id);
CREATE INDEX IF NOT EXISTS idx_charver_char_id  ON character_versions(character_id);