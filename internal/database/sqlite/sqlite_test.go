package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/database"
	//"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestDB creates a fresh in-memory SQLite database for each test.
// Using a unique URI per test prevents cross-test contamination while
// still exercising the real schema migration and write worker.
func newTestDB(t *testing.T) *sqliteDB {
	t.Helper()
	db := New()
	cfg := database.Config{}
	cfg.SQLite.Path = ":memory:"
	// Disable both automatic flush triggers so tests decide exactly when the
	// coalescing buffer is drained; a background tick landing mid-assertion
	// would make them flaky. Tests that exercise the triggers build their own.
	cfg.FlushInterval = time.Hour
	cfg.FlushThreshold = -1
	require.NoError(t, db.Connect(cfg, database.Options{}))
	t.Cleanup(func() { _ = db.Disconnect() })
	return db
}

// seedUser inserts a user row directly via NewCharacter (which upserts the user)
// or via SetUserFlags after a character has been created. For tests that only
// need a user without characters we create a throwaway character then delete it.
func seedUser(t *testing.T, db *sqliteDB, steamid string) {
	t.Helper()
	_, err := db.NewCharacter(context.Background(), steamid, 0, 1, "seed")
	require.NoError(t, err)
}

// seedCharacter creates a character and returns its ID.
func seedCharacter(t *testing.T, db *sqliteDB, steamid string, slot, size int, data string) uuid.UUID {
	t.Helper()
	id, err := db.NewCharacter(context.Background(), steamid, slot, size, data)
	require.NoError(t, err)
	return id
}

// flush commits any pending UpdateCharacter calls immediately, rather than
// waiting out the coalescing interval.
func flush(t *testing.T, db *sqliteDB) {
	t.Helper()
	require.NoError(t, db.buf.Flush(context.Background()))
}

// ─── Connect / Disconnect ────────────────────────────────────────────────────

func TestConnect_CreatesSchema(t *testing.T) {
	// If migrate fails the Connect call itself returns an error.
	db := newTestDB(t)
	assert.NotNil(t, db)
}

// ─── User tests ──────────────────────────────────────────────────────────────

func TestGetAllUsers_Empty(t *testing.T) {
	db := newTestDB(t)
	users, err := db.GetAllUsers(context.Background())
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestGetAllUsers_ReturnsAllUsers(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")
	seedUser(t, db, "steam2")

	users, err := db.GetAllUsers(context.Background())
	require.NoError(t, err)
	assert.Len(t, users, 2)

	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	assert.ElementsMatch(t, []string{"steam1", "steam2"}, ids)
}

func TestGetUser_Found(t *testing.T) {
	db := newTestDB(t)
	charID := seedCharacter(t, db, "steam1", 0, 100, "data")

	u, err := db.GetUser(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Equal(t, "steam1", u.ID)
	assert.Equal(t, charID, u.Characters[0])
}

func TestGetUser_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetUser(context.Background(), "nobody")
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetUser_LoadsDeletedCharacters(t *testing.T) {
	db := newTestDB(t)
	charID := seedCharacter(t, db, "steam1", 0, 100, "data")
	require.NoError(t, db.SoftDeleteCharacter(context.Background(), charID, 24*time.Hour))

	u, err := db.GetUser(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Equal(t, charID, u.DeletedCharacters[0])
	assert.Empty(t, u.Characters) // no longer in the active map
}

// ─── User flag tests ─────────────────────────────────────────────────────────

func TestSetAndGetUserFlags(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	flags := bitmask.Bitmask(0b1010)
	require.NoError(t, db.SetUserFlags(context.Background(), "steam1", flags))

	got, err := db.GetUserFlags(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Equal(t, flags, got)
}

func TestSetUserFlags_UserNotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.SetUserFlags(context.Background(), "ghost", bitmask.Bitmask(1))
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetUserFlags_UserNotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetUserFlags(context.Background(), "ghost")
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetUserFlags_DefaultZero(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	flags, err := db.GetUserFlags(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Equal(t, bitmask.Bitmask(0), flags)
}

func TestSetUserFlags_Overwrite(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	require.NoError(t, db.SetUserFlags(context.Background(), "steam1", bitmask.Bitmask(0xFF)))
	require.NoError(t, db.SetUserFlags(context.Background(), "steam1", bitmask.Bitmask(0x01)))

	flags, err := db.GetUserFlags(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Equal(t, bitmask.Bitmask(0x01), flags)
}

// ─── NewCharacter ─────────────────────────────────────────────────────────────

func TestNewCharacter_ReturnsUniqueIDs(t *testing.T) {
	db := newTestDB(t)
	id1 := seedCharacter(t, db, "steam1", 0, 100, "a")
	id2 := seedCharacter(t, db, "steam1", 1, 100, "b")
	assert.NotEqual(t, id1, id2)
}

func TestNewCharacter_CreatesUserIfMissing(t *testing.T) {
	db := newTestDB(t)
	seedCharacter(t, db, "newuser", 0, 10, "x")

	u, err := db.GetUser(context.Background(), "newuser")
	require.NoError(t, err)
	assert.Equal(t, "newuser", u.ID)
}

func TestNewCharacter_Idempotent_UserUpsert(t *testing.T) {
	db := newTestDB(t)
	// Two characters for the same user should not violate a UNIQUE constraint
	// on the users table.
	seedCharacter(t, db, "steam1", 0, 10, "a")
	seedCharacter(t, db, "steam1", 1, 20, "b")

	u, err := db.GetUser(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Len(t, u.Characters, 2)
}

// ─── GetCharacter ─────────────────────────────────────────────────────────────

func TestGetCharacter_Found(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 42, "mydata")

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, c.ID)
	assert.Equal(t, "steam1", c.SteamID)
	assert.Equal(t, 0, c.Slot)
	assert.Equal(t, 42, c.Data.Size)
	assert.Equal(t, "mydata", c.Data.Data)
	assert.Nil(t, c.DeletedAt)
}

func TestGetCharacter_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetCharacter(context.Background(), uuid.New())
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetCharacter_HasNoVersionsInitially(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Empty(t, c.Versions)
}

// ─── GetCharacters ────────────────────────────────────────────────────────────

func TestGetCharacters_ReturnsActiveOnly(t *testing.T) {
	db := newTestDB(t)
	id0 := seedCharacter(t, db, "steam1", 0, 10, "slot0")
	id1 := seedCharacter(t, db, "steam1", 1, 20, "slot1")

	// Soft-delete slot 1 — it should NOT appear in GetCharacters.
	require.NoError(t, db.SoftDeleteCharacter(context.Background(), id1, time.Hour))

	chars, err := db.GetCharacters(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Len(t, chars, 1)
	assert.Equal(t, id0, chars[0].ID)
}

func TestGetCharacters_Empty(t *testing.T) {
	db := newTestDB(t)
	chars, err := db.GetCharacters(context.Background(), "nobody")
	require.NoError(t, err)
	assert.Empty(t, chars)
}

func TestGetCharacters_KeyedBySlot(t *testing.T) {
	db := newTestDB(t)
	seedCharacter(t, db, "steam1", 3, 10, "three")
	seedCharacter(t, db, "steam1", 7, 20, "seven")

	chars, err := db.GetCharacters(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Equal(t, "three", chars[3].Data.Data)
	assert.Equal(t, "seven", chars[7].Data.Data)
}

// ─── LookUpCharacterID ───────────────────────────────────────────────────────

func TestLookUpCharacterID_Found(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 2, 10, "data")

	got, err := db.LookUpCharacterID(context.Background(), "steam1", 2)
	require.NoError(t, err)
	assert.Equal(t, id, got)
}

func TestLookUpCharacterID_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.LookUpCharacterID(context.Background(), "steam1", 99)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestLookUpCharacterID_IgnoresSoftDeleted(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	require.NoError(t, db.SoftDeleteCharacter(context.Background(), id, time.Hour))

	_, err := db.LookUpCharacterID(context.Background(), "steam1", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── UpdateCharacter ─────────────────────────────────────────────────────────

func TestUpdateCharacter_CoalescedFlush(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "original")

	// Two back-to-back updates — only the last should persist.
	require.NoError(t, db.UpdateCharacter(context.Background(), id, 20, "second", 0, 0))
	require.NoError(t, db.UpdateCharacter(context.Background(), id, 30, "third", 0, 0))

	flush(t, db)

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, 30, c.Data.Size)
	assert.Equal(t, "third", c.Data.Data)
}

func TestUpdateCharacter_CreatesFirstVersion(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "v0")

	require.NoError(t, db.UpdateCharacter(context.Background(), id, 20, "v1", 5, 0))
	flush(t, db)

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Len(t, c.Versions, 1)
	assert.Equal(t, "v0", c.Versions[0].Data)
}

func TestUpdateCharacter_RespectsBackupMax(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "init")

	// With backupMax=2 and backupTime=0 (always snapshot), after 3 updates
	// there should be at most 2 versions.
	for i, payload := range []string{"a", "b", "c"} {
		require.NoError(t, db.UpdateCharacter(context.Background(), id, i+1, payload, 2, 0))
		flush(t, db)
	}

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(c.Versions), 2)
}

// ─── SoftDeleteCharacter / RestoreCharacter ───────────────────────────────────

func TestSoftDeleteCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.SoftDeleteCharacter(context.Background(), id, 24*time.Hour))

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.NotNil(t, c.DeletedAt, "deleted_at should be set after soft delete")
}

func TestSoftDeleteCharacter_NotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.SoftDeleteCharacter(context.Background(), uuid.New(), time.Hour)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestSoftDeleteCharacter_AppearsInDeletedCharacters(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	require.NoError(t, db.SoftDeleteCharacter(context.Background(), id, time.Hour))

	u, err := db.GetUser(context.Background(), "steam1")
	require.NoError(t, err)
	assert.Equal(t, id, u.DeletedCharacters[0])
}

func TestRestoreCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	require.NoError(t, db.SoftDeleteCharacter(context.Background(), id, time.Hour))

	require.NoError(t, db.RestoreCharacter(context.Background(), id))

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Nil(t, c.DeletedAt)

	// Should reappear in active characters.
	got, err := db.LookUpCharacterID(context.Background(), "steam1", 0)
	require.NoError(t, err)
	assert.Equal(t, id, got)
}

func TestRestoreCharacter_NotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.RestoreCharacter(context.Background(), uuid.New())
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── DeleteCharacter ─────────────────────────────────────────────────────────

func TestDeleteCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.DeleteCharacter(context.Background(), id))

	_, err := db.GetCharacter(context.Background(), id)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── DeleteCharacterReference ─────────────────────────────────────────────────

func TestDeleteCharacterReference_RemovesActiveSlot(t *testing.T) {
	db := newTestDB(t)
	seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.DeleteCharacterReference(context.Background(), "steam1", 0))

	_, err := db.LookUpCharacterID(context.Background(), "steam1", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestDeleteCharacterReference_NoopWhenMissing(t *testing.T) {
	db := newTestDB(t)
	// Deleting a reference that doesn't exist should not return an error.
	assert.NoError(t, db.DeleteCharacterReference(context.Background(), "nobody", 99))
}

// ─── MoveCharacter ────────────────────────────────────────────────────────────

func TestMoveCharacter(t *testing.T) {
	db := newTestDB(t)
	// Both users must exist; MoveCharacter checks for the target user.
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	seedUser(t, db, "steam2")

	require.NoError(t, db.MoveCharacter(context.Background(), id, "steam2", 3))

	// Character now belongs to steam2 slot 3.
	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "steam2", c.SteamID)
	assert.Equal(t, 3, c.Slot)

	// Old slot on steam1 should be gone.
	_, err = db.LookUpCharacterID(context.Background(), "steam1", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)

	// New slot on steam2 should resolve.
	got, err := db.LookUpCharacterID(context.Background(), "steam2", 3)
	require.NoError(t, err)
	assert.Equal(t, id, got)
}

func TestMoveCharacter_CharacterNotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.MoveCharacter(context.Background(), uuid.New(), "steam2", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestMoveCharacter_TargetUserNotFound(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	err := db.MoveCharacter(context.Background(), id, "ghost", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── CopyCharacter ────────────────────────────────────────────────────────────

func TestCopyCharacter(t *testing.T) {
	db := newTestDB(t)
	origID := seedCharacter(t, db, "steam1", 0, 42, "original")

	newID, err := db.CopyCharacter(context.Background(), origID, "steam2", 1)
	require.NoError(t, err)
	assert.NotEqual(t, origID, newID)

	// Original unchanged.
	orig, err := db.GetCharacter(context.Background(), origID)
	require.NoError(t, err)
	assert.Equal(t, "steam1", orig.SteamID)

	// Copy has correct owner and payload.
	copy, err := db.GetCharacter(context.Background(), newID)
	require.NoError(t, err)
	assert.Equal(t, "steam2", copy.SteamID)
	assert.Equal(t, 1, copy.Slot)
	assert.Equal(t, "original", copy.Data.Data)
	assert.Equal(t, 42, copy.Data.Size)
}

func TestCopyCharacter_CreatesTargetUserIfMissing(t *testing.T) {
	db := newTestDB(t)
	origID := seedCharacter(t, db, "steam1", 0, 10, "data")

	_, err := db.CopyCharacter(context.Background(), origID, "brandnew", 0)
	require.NoError(t, err)

	u, err := db.GetUser(context.Background(), "brandnew")
	require.NoError(t, err)
	assert.Equal(t, "brandnew", u.ID)
}

func TestCopyCharacter_OriginalNotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.CopyCharacter(context.Background(), uuid.New(), "steam2", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── RollbackCharacter ────────────────────────────────────────────────────────

func TestRollbackCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "v0")

	// Create a version by updating once (backupMax>0, backupTime=0 means always snapshot).
	require.NoError(t, db.UpdateCharacter(context.Background(), id, 2, "v1", 5, 0))
	flush(t, db)

	// Rollback to version index 0 (the "v0" snapshot).
	require.NoError(t, db.RollbackCharacter(context.Background(), id, 0))

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "v0", c.Data.Data)
	assert.Equal(t, 1, c.Data.Size)
}

func TestRollbackCharacter_InvalidIndex(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "data")
	err := db.RollbackCharacter(context.Background(), id, 99)
	assert.Error(t, err)
}

func TestRollbackCharacterToLatest(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "v0")

	require.NoError(t, db.UpdateCharacter(context.Background(), id, 2, "v1", 5, 0))
	flush(t, db)
	require.NoError(t, db.UpdateCharacter(context.Background(), id, 3, "v2", 5, 0))
	flush(t, db)

	// Manually clobber the current data to simulate corruption.
	require.NoError(t, db.UpdateCharacter(context.Background(), id, 0, "corrupt", 0, 0))
	flush(t, db)

	require.NoError(t, db.RollbackCharacterToLatest(context.Background(), id))

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	// Should have rolled back to the latest version (v2, since backupMax was 5).
	assert.NotEqual(t, "corrupt", c.Data.Data)
}

func TestRollbackCharacterToLatest_NoVersions(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "data")
	err := db.RollbackCharacterToLatest(context.Background(), id)
	assert.Error(t, err)
}

// ─── DeleteCharacterVersions ─────────────────────────────────────────────────

func TestDeleteCharacterVersions(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "v0")

	require.NoError(t, db.UpdateCharacter(context.Background(), id, 2, "v1", 5, 0))
	flush(t, db)

	require.NoError(t, db.DeleteCharacterVersions(context.Background(), id))

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Empty(t, c.Versions)
}

func TestDeleteCharacterVersions_NoVersions(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "data")
	// Should succeed even if there are no versions to delete.
	assert.NoError(t, db.DeleteCharacterVersions(context.Background(), id))
}

// ─── SyncToDisk / RunGC ──────────────────────────────────────────────────────

func TestSyncToDisk(t *testing.T) {
	db := newTestDB(t)
	assert.NoError(t, db.SyncToDisk(context.Background()))
}

func TestRunGC_PurgesExpiredCharacters(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	// Use a negative duration so expires_at is already in the past.
	require.NoError(t, db.SoftDeleteCharacter(context.Background(), id, -1*time.Second))

	require.NoError(t, db.RunGC(context.Background()))

	_, err := db.GetCharacter(context.Background(), id)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestRunGC_KeepsNonExpiredCharacters(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.SoftDeleteCharacter(context.Background(), id, 24*time.Hour))
	require.NoError(t, db.RunGC(context.Background()))

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestRunGC_FlushesBeforeGC(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "old")

	require.NoError(t, db.UpdateCharacter(context.Background(), id, 99, "new", 0, 0))
	// RunGC should flush the pending update before running the GC query.
	require.NoError(t, db.RunGC(context.Background()))

	require.Zero(t, db.buf.Len())
	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "new", c.Data.Data)
}

// GC used to abort whenever its pre-flush failed, so a single character that
// could not be written stopped expired rows from ever being collected.
func TestRunGC_CollectsEvenWhenFlushCannotApply(t *testing.T) {
	db := newTestDB(t)

	expired := seedCharacter(t, db, "steam1", 0, 10, "expired")
	require.NoError(t, db.SoftDeleteCharacter(context.Background(), expired, -1*time.Second))

	// Queue an update against a character that does not exist. The flush can
	// never apply it, which is exactly the condition that used to block GC.
	require.NoError(t, db.UpdateCharacter(context.Background(), uuid.New(), 1, "orphan", 0, 0))

	require.NoError(t, db.RunGC(context.Background()))

	_, err := db.GetCharacter(context.Background(), expired)
	assert.ErrorIs(t, err, database.ErrNoDocument, "the expired character should still have been purged")
}

// A character deleted while an update was queued for it must not take the rest
// of the batch down with it. This used to silently drop every other player's
// save in the same flush.
func TestFlush_DeletedCharacterDoesNotLoseOtherUpdates(t *testing.T) {
	db := newTestDB(t)

	doomed := seedCharacter(t, db, "steam1", 0, 1, "doomed")
	keep1 := seedCharacter(t, db, "steam2", 0, 1, "old1")
	keep2 := seedCharacter(t, db, "steam3", 0, 1, "old2")

	ctx := context.Background()
	require.NoError(t, db.UpdateCharacter(ctx, keep1, 11, "new1", 0, 0))
	require.NoError(t, db.UpdateCharacter(ctx, keep2, 12, "new2", 0, 0))
	require.NoError(t, db.UpdateCharacter(ctx, doomed, 13, "never", 0, 0))

	// Delete the row out from under the buffer without going through
	// DeleteCharacter, which would helpfully drop the pending entry for us.
	require.NoError(t, db.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `DELETE FROM characters WHERE id = ?`, doomed.String())
		return err
	}))

	require.NoError(t, db.buf.Flush(ctx))

	c1, err := db.GetCharacter(ctx, keep1)
	require.NoError(t, err)
	assert.Equal(t, "new1", c1.Data.Data)

	c2, err := db.GetCharacter(ctx, keep2)
	require.NoError(t, err)
	assert.Equal(t, "new2", c2.Data.Data)

	assert.Zero(t, db.buf.Len(), "the unapplicable entry must be discarded, not retried forever")
}

// ─── Buffer / read consistency ───────────────────────────────────────────────

// Reads used to return the last committed row, so a save followed by a read
// gave back stale data for the length of the flush interval.
func TestGetCharacter_ReflectsUnflushedUpdate(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "old")

	require.NoError(t, db.UpdateCharacter(context.Background(), id, 42, "fresh", 0, 0))

	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "fresh", c.Data.Data)
	assert.Equal(t, 42, c.Data.Size)
}

func TestGetCharacters_ReflectsUnflushedUpdate(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "old")

	require.NoError(t, db.UpdateCharacter(context.Background(), id, 42, "fresh", 0, 0))

	chars, err := db.GetCharacters(context.Background(), "steam1")
	require.NoError(t, err)
	require.Contains(t, chars, 0)
	assert.Equal(t, "fresh", chars[0].Data.Data)
}

func TestCopyCharacter_CopiesUnflushedUpdate(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "old")
	seedUser(t, db, "steam2")

	require.NoError(t, db.UpdateCharacter(context.Background(), id, 42, "fresh", 0, 0))

	newID, err := db.CopyCharacter(context.Background(), id, "steam2", 1)
	require.NoError(t, err)

	flush(t, db)

	c, err := db.GetCharacter(context.Background(), newID)
	require.NoError(t, err)
	assert.Equal(t, "fresh", c.Data.Data, "a copy must not capture pre-update data")
}

// A queued update predates the rollback, so flushing it afterwards would
// silently undo what the rollback just restored.
func TestRollbackCharacter_NotClobberedByPendingUpdate(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "v0")

	ctx := context.Background()
	// Create one version ("v0") and make "v1" current.
	require.NoError(t, db.UpdateCharacter(ctx, id, 2, "v1", 5, 0))
	flush(t, db)

	// A save lands while the player is asking for a rollback.
	require.NoError(t, db.UpdateCharacter(ctx, id, 3, "v2", 5, 0))
	require.NoError(t, db.RollbackCharacterToLatest(ctx, id))

	flush(t, db)

	c, err := db.GetCharacter(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "v0", c.Data.Data, "the rollback must survive the next flush")
}

func TestDeleteCharacter_DropsPendingUpdate(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "data")

	ctx := context.Background()
	require.NoError(t, db.UpdateCharacter(ctx, id, 2, "queued", 0, 0))
	require.NoError(t, db.DeleteCharacter(ctx, id))

	assert.Zero(t, db.buf.Len())
	require.NoError(t, db.buf.Flush(ctx))
}

// ─── Shutdown ────────────────────────────────────────────────────────────────

// Disconnect used to race its two workers: if the write worker stopped first,
// the final flush parked forever on a reply nobody was left to send, hanging
// the whole shutdown and losing the queued update.
func TestDisconnect_FlushesAndReturns(t *testing.T) {
	dir := t.TempDir()

	db := New()
	cfg := database.Config{}
	cfg.SQLite.Path = filepath.Join(dir, "shutdown.db")
	cfg.FlushInterval = time.Hour
	cfg.FlushThreshold = -1
	require.NoError(t, db.Connect(cfg, database.Options{}))

	ctx := context.Background()
	id, err := db.NewCharacter(ctx, "steam1", 0, 1, "old")
	require.NoError(t, err)
	require.NoError(t, db.UpdateCharacter(ctx, id, 99, "queued", 0, 0))

	done := make(chan error, 1)
	go func() { done <- db.Disconnect() }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(15 * time.Second):
		t.Fatal("Disconnect deadlocked")
	}

	// Reopen and confirm the queued update actually made it to disk.
	reopened := New()
	require.NoError(t, reopened.Connect(cfg, database.Options{}))
	t.Cleanup(func() { _ = reopened.Disconnect() })

	c, err := reopened.GetCharacter(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "queued", c.Data.Data, "the final flush must persist before the DB closes")
}

// ─── Flush triggers ──────────────────────────────────────────────────────────

func TestFlushThreshold_WritesWithoutWaitingForTheInterval(t *testing.T) {
	db := New()
	cfg := database.Config{}
	cfg.SQLite.Path = ":memory:"
	// An hour-long interval: only the threshold can drain this buffer.
	cfg.FlushInterval = time.Hour
	cfg.FlushThreshold = 4
	require.NoError(t, db.Connect(cfg, database.Options{}))
	t.Cleanup(func() { _ = db.Disconnect() })

	ctx := context.Background()
	ids := make([]uuid.UUID, 4)
	for i := range ids {
		ids[i] = seedCharacter(t, db, "steam1", i, 1, "old")
	}
	for i, id := range ids {
		require.NoError(t, db.UpdateCharacter(ctx, id, i+10, fmt.Sprintf("new-%d", i), 0, 0))
	}

	require.Eventually(t, func() bool {
		return db.buf.Len() == 0
	}, 5*time.Second, 10*time.Millisecond, "reaching the threshold should trigger a flush")
}

// ─── schema.User integrity ───────────────────────────────────────────────────

func TestGetUser_CharacterMapsAreInitialized(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	u, err := db.GetUser(context.Background(), "steam1")
	require.NoError(t, err)
	// Maps must not be nil so callers can safely do map[key] lookups.
	assert.NotNil(t, u.Characters)
	assert.NotNil(t, u.DeletedCharacters)
}

func TestGetAllUsers_CharacterMapsAreInitialized(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	users, err := db.GetAllUsers(context.Background())
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.NotNil(t, users[0].Characters)
	assert.NotNil(t, users[0].DeletedCharacters)
}

// ─── Concurrency / coalescing sanity check ───────────────────────────────────

func TestUpdateCharacter_ConcurrentUpdatesCoalesce(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "init")

	const workers = 20
	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		i := i
		go func() {
			_ = db.UpdateCharacter(context.Background(), id, i, fmt.Sprintf("payload-%d", i), 0, 0)
			done <- struct{}{}
		}()
	}
	for i := 0; i < workers; i++ {
		<-done
	}

	flush(t, db)

	// We don't care which payload won — just that the DB is consistent.
	c, err := db.GetCharacter(context.Background(), id)
	require.NoError(t, err)
	assert.NotEmpty(t, c.Data.Data)
}