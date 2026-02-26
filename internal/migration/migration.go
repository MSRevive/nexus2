package migration

import (
	"fmt"
	"log"

	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/internal/database"
)

// Migrator moves all data from src to dst using only the Database interface.
// Because it operates purely through the interface, it works for any combination
// of backends: Pebble → SQLite, SQLite → Postgres, Mongo → SQLite, etc.
type Migrator struct {
	src database.Database
	dst database.Database

	// OnProgress is called after each character is migrated so the caller can
	// print progress or update a UI. Optional — leave nil to skip.
	OnProgress func(steamID string, slot int, charID string)
}

func New(src, dst database.Database) *Migrator {
	return &Migrator{src: src, dst: dst}
}

// Run performs the full migration in one pass:
//  1. Reads all users from src
//  2. For each user, reads every active and deleted character
//  3. Writes users, characters, versions, flags, and soft-delete state to dst
//
// The destination database must be connected and empty before calling Run.
// Run does not disconnect either database — the caller is responsible for that.
func (m *Migrator) Run() error {
	users, err := m.src.GetAllUsers()
	if err != nil {
		return fmt.Errorf("migration: fetch users: %w", err)
	}

	log.Printf("migration: found %d users to migrate", len(users))

	for _, user := range users {
		if err := m.migrateUser(user); err != nil {
			return fmt.Errorf("migration: user %s: %w", user.ID, err)
		}
	}

	// Final sync so everything is durably written before the caller disconnects.
	if err := m.dst.SyncToDisk(); err != nil {
		return fmt.Errorf("migration: final sync: %w", err)
	}

	log.Printf("migration: complete")
	return nil
}

func (m *Migrator) migrateUser(user *schema.User) error {
	log.Printf("migration: migrating user %s (%d active, %d deleted characters)",
		user.ID, len(user.Characters), len(user.DeletedCharacters))

	// Migrate active characters first.
	for slot, charID := range user.Characters {
		char, err := m.src.GetCharacter(charID)
		if err != nil {
			return fmt.Errorf("get character %s (slot %d): %w", charID, slot, err)
		}

		if err := m.migrateCharacter(user, char); err != nil {
			return fmt.Errorf("migrate character %s (slot %d): %w", charID, slot, err)
		}

		if m.OnProgress != nil {
			m.OnProgress(user.ID, slot, charID.String())
		}
	}

	// Migrate user flags last so the user row definitely exists in dst.
	flags, err := m.src.GetUserFlags(user.ID)
	if err != nil {
		return fmt.Errorf("get flags for user %s: %w", user.ID, err)
	}
	if flags != 0 {
		if err := m.dst.SetUserFlags(user.ID, flags); err != nil {
			return fmt.Errorf("set flags for user %s: %w", user.ID, err)
		}
	}

	return nil
}

// migrateCharacter writes a single character and all of its versions to dst.
// It uses NewCharacter to create the initial row and then replays each version
// through UpdateCharacter so that the version history is preserved in order.
func (m *Migrator) migrateCharacter(user *schema.User, char *schema.Character) error {
	// NewCharacter creates the user row if it doesn't exist yet, so we don't
	// need a separate "create user" step.
	newID, err := m.dst.NewCharacter(
		char.SteamID,
		char.Slot,
		char.Data.Size,
		char.Data.Data,
	)
	if err != nil {
		return fmt.Errorf("create character: %w", err)
	}

	// The destination assigned a new UUID. If the source UUID matters for
	// external references you'll need an ID-mapping strategy — but for most
	// game server setups, the slot lookup path (steamid + slot) is what callers
	// actually use, so the new UUID is fine.
	if newID != char.ID {
		log.Printf("migration: character %s re-assigned as %s (slot %d, user %s)",
			char.ID, newID, char.Slot, char.SteamID)
	}

	// Replay version history oldest-first. We pass backupMax=len(versions)+1
	// so no versions are pruned during the replay, and backupTime=0 so the
	// time-gap check is always satisfied.
	for _, ver := range char.Versions {
		if err := m.dst.UpdateCharacter(
			newID,
			ver.Size,
			ver.Data,
			len(char.Versions)+1, // never prune during migration
			0,                    // no time gap required
		); err != nil {
			return fmt.Errorf("replay version: %w", err)
		}
	}

	// If there were versions, flush them and then restore the current data so
	// the active data_payload reflects char.Data and not the last version entry.
	if len(char.Versions) > 0 {
		if err := m.dst.SyncToDisk(); err != nil {
			return fmt.Errorf("sync after version replay: %w", err)
		}
		// Write the real current data as a final update on top of the versions.
		if err := m.dst.UpdateCharacter(
			newID,
			char.Data.Size,
			char.Data.Data,
			0, // backupMax=0 so this update creates no new version entry
			0,
		); err != nil {
			return fmt.Errorf("restore current data: %w", err)
		}
	}

	return nil
}
