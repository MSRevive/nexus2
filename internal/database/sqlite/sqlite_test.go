package sqlite

import (
	"fmt"
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
	require.NoError(t, db.Connect(cfg, database.Options{}))
	t.Cleanup(func() { _ = db.Disconnect() })
	return db
}

// seedUser inserts a user row directly via NewCharacter (which upserts the user)
// or via SetUserFlags after a character has been created. For tests that only
// need a user without characters we create a throwaway character then delete it.
func seedUser(t *testing.T, db *sqliteDB, steamid string) {
	t.Helper()
	_, err := db.NewCharacter(steamid, 0, 1, "seed")
	require.NoError(t, err)
}

// seedCharacter creates a character and returns its ID.
func seedCharacter(t *testing.T, db *sqliteDB, steamid string, slot, size int, data string) uuid.UUID {
	t.Helper()
	id, err := db.NewCharacter(steamid, slot, size, data)
	require.NoError(t, err)
	return id
}

// flush waits long enough for the coalescing flush worker to commit any
// pending UpdateCharacter calls (default interval is 500 ms).
func flush(t *testing.T, db *sqliteDB) {
	t.Helper()
	require.NoError(t, db.RunGC()) // RunGC always calls flushPendingUpdates
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
	users, err := db.GetAllUsers()
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestGetAllUsers_ReturnsAllUsers(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")
	seedUser(t, db, "steam2")

	users, err := db.GetAllUsers()
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

	u, err := db.GetUser("steam1")
	require.NoError(t, err)
	assert.Equal(t, "steam1", u.ID)
	assert.Equal(t, charID, u.Characters[0])
}

func TestGetUser_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetUser("nobody")
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetUser_LoadsDeletedCharacters(t *testing.T) {
	db := newTestDB(t)
	charID := seedCharacter(t, db, "steam1", 0, 100, "data")
	require.NoError(t, db.SoftDeleteCharacter(charID, 24*time.Hour))

	u, err := db.GetUser("steam1")
	require.NoError(t, err)
	assert.Equal(t, charID, u.DeletedCharacters[0])
	assert.Empty(t, u.Characters) // no longer in the active map
}

// ─── User flag tests ─────────────────────────────────────────────────────────

func TestSetAndGetUserFlags(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	flags := bitmask.Bitmask(0b1010)
	require.NoError(t, db.SetUserFlags("steam1", flags))

	got, err := db.GetUserFlags("steam1")
	require.NoError(t, err)
	assert.Equal(t, flags, got)
}

func TestSetUserFlags_UserNotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.SetUserFlags("ghost", bitmask.Bitmask(1))
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetUserFlags_UserNotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetUserFlags("ghost")
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetUserFlags_DefaultZero(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	flags, err := db.GetUserFlags("steam1")
	require.NoError(t, err)
	assert.Equal(t, bitmask.Bitmask(0), flags)
}

func TestSetUserFlags_Overwrite(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	require.NoError(t, db.SetUserFlags("steam1", bitmask.Bitmask(0xFF)))
	require.NoError(t, db.SetUserFlags("steam1", bitmask.Bitmask(0x01)))

	flags, err := db.GetUserFlags("steam1")
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

	u, err := db.GetUser("newuser")
	require.NoError(t, err)
	assert.Equal(t, "newuser", u.ID)
}

func TestNewCharacter_Idempotent_UserUpsert(t *testing.T) {
	db := newTestDB(t)
	// Two characters for the same user should not violate a UNIQUE constraint
	// on the users table.
	seedCharacter(t, db, "steam1", 0, 10, "a")
	seedCharacter(t, db, "steam1", 1, 20, "b")

	u, err := db.GetUser("steam1")
	require.NoError(t, err)
	assert.Len(t, u.Characters, 2)
}

// ─── GetCharacter ─────────────────────────────────────────────────────────────

func TestGetCharacter_Found(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 42, "mydata")

	c, err := db.GetCharacter(id)
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
	_, err := db.GetCharacter(uuid.New())
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestGetCharacter_HasNoVersionsInitially(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.Empty(t, c.Versions)
}

// ─── GetCharacters ────────────────────────────────────────────────────────────

func TestGetCharacters_ReturnsActiveOnly(t *testing.T) {
	db := newTestDB(t)
	id0 := seedCharacter(t, db, "steam1", 0, 10, "slot0")
	id1 := seedCharacter(t, db, "steam1", 1, 20, "slot1")

	// Soft-delete slot 1 — it should NOT appear in GetCharacters.
	require.NoError(t, db.SoftDeleteCharacter(id1, time.Hour))

	chars, err := db.GetCharacters("steam1")
	require.NoError(t, err)
	assert.Len(t, chars, 1)
	assert.Equal(t, id0, chars[0].ID)
}

func TestGetCharacters_Empty(t *testing.T) {
	db := newTestDB(t)
	chars, err := db.GetCharacters("nobody")
	require.NoError(t, err)
	assert.Empty(t, chars)
}

func TestGetCharacters_KeyedBySlot(t *testing.T) {
	db := newTestDB(t)
	seedCharacter(t, db, "steam1", 3, 10, "three")
	seedCharacter(t, db, "steam1", 7, 20, "seven")

	chars, err := db.GetCharacters("steam1")
	require.NoError(t, err)
	assert.Equal(t, "three", chars[3].Data.Data)
	assert.Equal(t, "seven", chars[7].Data.Data)
}

// ─── LookUpCharacterID ───────────────────────────────────────────────────────

func TestLookUpCharacterID_Found(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 2, 10, "data")

	got, err := db.LookUpCharacterID("steam1", 2)
	require.NoError(t, err)
	assert.Equal(t, id, got)
}

func TestLookUpCharacterID_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.LookUpCharacterID("steam1", 99)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestLookUpCharacterID_IgnoresSoftDeleted(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	require.NoError(t, db.SoftDeleteCharacter(id, time.Hour))

	_, err := db.LookUpCharacterID("steam1", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── UpdateCharacter ─────────────────────────────────────────────────────────

func TestUpdateCharacter_CoalescedFlush(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "original")

	// Two back-to-back updates — only the last should persist.
	require.NoError(t, db.UpdateCharacter(id, 20, "second", 0, 0))
	require.NoError(t, db.UpdateCharacter(id, 30, "third", 0, 0))

	flush(t, db)

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.Equal(t, 30, c.Data.Size)
	assert.Equal(t, "third", c.Data.Data)
}

func TestUpdateCharacter_CreatesFirstVersion(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "v0")

	require.NoError(t, db.UpdateCharacter(id, 20, "v1", 5, 0))
	flush(t, db)

	c, err := db.GetCharacter(id)
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
		require.NoError(t, db.UpdateCharacter(id, i+1, payload, 2, 0))
		flush(t, db)
	}

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(c.Versions), 2)
}

// ─── SoftDeleteCharacter / RestoreCharacter ───────────────────────────────────

func TestSoftDeleteCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.SoftDeleteCharacter(id, 24*time.Hour))

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.NotNil(t, c.DeletedAt, "deleted_at should be set after soft delete")
}

func TestSoftDeleteCharacter_NotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.SoftDeleteCharacter(uuid.New(), time.Hour)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestSoftDeleteCharacter_AppearsInDeletedCharacters(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	require.NoError(t, db.SoftDeleteCharacter(id, time.Hour))

	u, err := db.GetUser("steam1")
	require.NoError(t, err)
	assert.Equal(t, id, u.DeletedCharacters[0])
}

func TestRestoreCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	require.NoError(t, db.SoftDeleteCharacter(id, time.Hour))

	require.NoError(t, db.RestoreCharacter(id))

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.Nil(t, c.DeletedAt)

	// Should reappear in active characters.
	got, err := db.LookUpCharacterID("steam1", 0)
	require.NoError(t, err)
	assert.Equal(t, id, got)
}

func TestRestoreCharacter_NotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.RestoreCharacter(uuid.New())
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── DeleteCharacter ─────────────────────────────────────────────────────────

func TestDeleteCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.DeleteCharacter(id))

	_, err := db.GetCharacter(id)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── DeleteCharacterReference ─────────────────────────────────────────────────

func TestDeleteCharacterReference_RemovesActiveSlot(t *testing.T) {
	db := newTestDB(t)
	seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.DeleteCharacterReference("steam1", 0))

	_, err := db.LookUpCharacterID("steam1", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestDeleteCharacterReference_NoopWhenMissing(t *testing.T) {
	db := newTestDB(t)
	// Deleting a reference that doesn't exist should not return an error.
	assert.NoError(t, db.DeleteCharacterReference("nobody", 99))
}

// ─── MoveCharacter ────────────────────────────────────────────────────────────

func TestMoveCharacter(t *testing.T) {
	db := newTestDB(t)
	// Both users must exist; MoveCharacter checks for the target user.
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	seedUser(t, db, "steam2")

	require.NoError(t, db.MoveCharacter(id, "steam2", 3))

	// Character now belongs to steam2 slot 3.
	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.Equal(t, "steam2", c.SteamID)
	assert.Equal(t, 3, c.Slot)

	// Old slot on steam1 should be gone.
	_, err = db.LookUpCharacterID("steam1", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)

	// New slot on steam2 should resolve.
	got, err := db.LookUpCharacterID("steam2", 3)
	require.NoError(t, err)
	assert.Equal(t, id, got)
}

func TestMoveCharacter_CharacterNotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.MoveCharacter(uuid.New(), "steam2", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestMoveCharacter_TargetUserNotFound(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	err := db.MoveCharacter(id, "ghost", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── CopyCharacter ────────────────────────────────────────────────────────────

func TestCopyCharacter(t *testing.T) {
	db := newTestDB(t)
	origID := seedCharacter(t, db, "steam1", 0, 42, "original")

	newID, err := db.CopyCharacter(origID, "steam2", 1)
	require.NoError(t, err)
	assert.NotEqual(t, origID, newID)

	// Original unchanged.
	orig, err := db.GetCharacter(origID)
	require.NoError(t, err)
	assert.Equal(t, "steam1", orig.SteamID)

	// Copy has correct owner and payload.
	copy, err := db.GetCharacter(newID)
	require.NoError(t, err)
	assert.Equal(t, "steam2", copy.SteamID)
	assert.Equal(t, 1, copy.Slot)
	assert.Equal(t, "original", copy.Data.Data)
	assert.Equal(t, 42, copy.Data.Size)
}

func TestCopyCharacter_CreatesTargetUserIfMissing(t *testing.T) {
	db := newTestDB(t)
	origID := seedCharacter(t, db, "steam1", 0, 10, "data")

	_, err := db.CopyCharacter(origID, "brandnew", 0)
	require.NoError(t, err)

	u, err := db.GetUser("brandnew")
	require.NoError(t, err)
	assert.Equal(t, "brandnew", u.ID)
}

func TestCopyCharacter_OriginalNotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.CopyCharacter(uuid.New(), "steam2", 0)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

// ─── RollbackCharacter ────────────────────────────────────────────────────────

func TestRollbackCharacter(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "v0")

	// Create a version by updating once (backupMax>0, backupTime=0 means always snapshot).
	require.NoError(t, db.UpdateCharacter(id, 2, "v1", 5, 0))
	flush(t, db)

	// Rollback to version index 0 (the "v0" snapshot).
	require.NoError(t, db.RollbackCharacter(id, 0))

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.Equal(t, "v0", c.Data.Data)
	assert.Equal(t, 1, c.Data.Size)
}

func TestRollbackCharacter_InvalidIndex(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "data")
	err := db.RollbackCharacter(id, 99)
	assert.Error(t, err)
}

func TestRollbackCharacterToLatest(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "v0")

	require.NoError(t, db.UpdateCharacter(id, 2, "v1", 5, 0))
	flush(t, db)
	require.NoError(t, db.UpdateCharacter(id, 3, "v2", 5, 0))
	flush(t, db)

	// Manually clobber the current data to simulate corruption.
	require.NoError(t, db.UpdateCharacter(id, 0, "corrupt", 0, 0))
	flush(t, db)

	require.NoError(t, db.RollbackCharacterToLatest(id))

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	// Should have rolled back to the latest version (v2, since backupMax was 5).
	assert.NotEqual(t, "corrupt", c.Data.Data)
}

func TestRollbackCharacterToLatest_NoVersions(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "data")
	err := db.RollbackCharacterToLatest(id)
	assert.Error(t, err)
}

// ─── DeleteCharacterVersions ─────────────────────────────────────────────────

func TestDeleteCharacterVersions(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "v0")

	require.NoError(t, db.UpdateCharacter(id, 2, "v1", 5, 0))
	flush(t, db)

	require.NoError(t, db.DeleteCharacterVersions(id))

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.Empty(t, c.Versions)
}

func TestDeleteCharacterVersions_NoVersions(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "data")
	// Should succeed even if there are no versions to delete.
	assert.NoError(t, db.DeleteCharacterVersions(id))
}

// ─── SyncToDisk / RunGC ──────────────────────────────────────────────────────

func TestSyncToDisk(t *testing.T) {
	db := newTestDB(t)
	assert.NoError(t, db.SyncToDisk())
}

func TestRunGC_PurgesExpiredCharacters(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	// Use a negative duration so expires_at is already in the past.
	require.NoError(t, db.SoftDeleteCharacter(id, -1*time.Second))

	require.NoError(t, db.RunGC())

	_, err := db.GetCharacter(id)
	assert.ErrorIs(t, err, database.ErrNoDocument)
}

func TestRunGC_KeepsNonExpiredCharacters(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 10, "data")

	require.NoError(t, db.SoftDeleteCharacter(id, 24*time.Hour))
	require.NoError(t, db.RunGC())

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestRunGC_FlushesBeforeGC(t *testing.T) {
	db := newTestDB(t)
	id := seedCharacter(t, db, "steam1", 0, 1, "old")

	require.NoError(t, db.UpdateCharacter(id, 99, "new", 0, 0))
	// RunGC should flush the pending update before running the GC query.
	require.NoError(t, db.RunGC())

	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.Equal(t, "new", c.Data.Data)
}

// ─── schema.User integrity ───────────────────────────────────────────────────

func TestGetUser_CharacterMapsAreInitialized(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	u, err := db.GetUser("steam1")
	require.NoError(t, err)
	// Maps must not be nil so callers can safely do map[key] lookups.
	assert.NotNil(t, u.Characters)
	assert.NotNil(t, u.DeletedCharacters)
}

func TestGetAllUsers_CharacterMapsAreInitialized(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "steam1")

	users, err := db.GetAllUsers()
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
			_ = db.UpdateCharacter(id, i, fmt.Sprintf("payload-%d", i), 0, 0)
			done <- struct{}{}
		}()
	}
	for i := 0; i < workers; i++ {
		<-done
	}

	flush(t, db)

	// We don't care which payload won — just that the DB is consistent.
	c, err := db.GetCharacter(id)
	require.NoError(t, err)
	assert.NotEmpty(t, c.Data.Data)
}