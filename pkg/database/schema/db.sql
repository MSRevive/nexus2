-- Users table
CREATE TABLE users (
    id          TEXT PRIMARY KEY,  -- SteamID64
    revision    INTEGER NOT NULL DEFAULT 0,
    flags       INTEGER NOT NULL DEFAULT 0,  -- uint32, stored as INTEGER
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Characters table
CREATE TABLE characters (
    id          TEXT PRIMARY KEY,             -- UUID stored as TEXT (SQLite) or UUID type (Postgres)
    steam_id    TEXT NOT NULL REFERENCES users(id),
    slot        INTEGER NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  DATETIME,                     -- NULL means not deleted

    -- Current active CharacterData (denormalized for fast reads)
    data_created_at DATETIME,
    data_size       INTEGER,
    data_payload    TEXT,                     -- the actual character data blob

    UNIQUE (steam_id, slot)                   -- a user can't have two characters in the same slot
);

-- Soft-deleted characters (mirrors characters but separated for your deleted_characters map)
-- You could also handle this with just deleted_at on characters, but keeping the slot
-- association after deletion requires the explicit map you have in Mongo.
CREATE TABLE deleted_characters (
    steam_id    TEXT NOT NULL REFERENCES users(id),
    slot        INTEGER NOT NULL,
    character_id TEXT NOT NULL REFERENCES characters(id),
    deleted_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (steam_id, slot)
);

-- Character version history (the []CharacterData versions array)
CREATE TABLE character_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,  -- surrogate key
    character_id    TEXT NOT NULL REFERENCES characters(id),
    version_number  INTEGER NOT NULL,
    created_at      DATETIME NOT NULL,
    size            INTEGER NOT NULL,
    data_payload    TEXT NOT NULL,
    UNIQUE (character_id, version_number)
);