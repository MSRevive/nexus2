CREATE TABLE IF NOT EXISTS `characters` (
	'id' TEXT PRIMARY KEY,
	'steamid' TEXT KEY NOT NULL,
	'slot' INTEGER KEY NOT NULL,
	'created_at' TEXT NOT NULL,
	'deleted_at' TEXT,
	'data' BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS `character_versions` (
	'id' TEXT KEY NOT NULL UNIQUE,
	'versions' BLOB,
	FOREIGN KEY ('id')
		REFERENCES `characters` ('id')
			ON UPDATE NO ACTION
			ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `deleted_characters` (
	'id' TEXT KEY NOT NULL UNIQUE,
	'expired' REAL NOT NULL,
	FOREIGN KEY ('id')
		REFERENCES `characters` ('id')
			ON UPDATE NO ACTION
			ON DELETE CASCADE
);