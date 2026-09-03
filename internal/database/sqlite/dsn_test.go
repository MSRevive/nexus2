package sqlite

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/msrevive/nexus2/internal/database"

	"github.com/stretchr/testify/require"
)

// newFileDB opens a file backed database, which is the only way to see the
// journal mode: an in-memory database is always "memory" and silently ignores
// a request for WAL, so :memory: cannot cover this.
func newFileDB(t *testing.T) *sqliteDB {
	t.Helper()

	db := New()
	cfg := database.Config{}
	cfg.SQLite.Path = filepath.Join(t.TempDir(), "test.db")
	require.NoError(t, db.Connect(cfg, database.Options{}))
	t.Cleanup(func() { _ = db.Disconnect() })

	return db
}

// The DSN pragmas have to actually reach the driver. modernc.org/sqlite only
// understands _pragma=name(value) and silently drops anything else, so a DSN
// written for another driver leaves every one of these at its default.
func TestConnectAppliesDSNPragmas(t *testing.T) {
	db := newFileDB(t)

	t.Run("journal_mode is WAL", func(t *testing.T) {
		var mode string
		require.NoError(t, db.db.QueryRow("PRAGMA journal_mode").Scan(&mode))
		require.Equal(t, "wal", strings.ToLower(mode))
	})

	t.Run("busy_timeout is set", func(t *testing.T) {
		var timeout int
		require.NoError(t, db.db.QueryRow("PRAGMA busy_timeout").Scan(&timeout))
		require.Equal(t, 5000, timeout)
	})

	t.Run("synchronous is NORMAL", func(t *testing.T) {
		var sync int
		require.NoError(t, db.db.QueryRow("PRAGMA synchronous").Scan(&sync))
		require.Equal(t, 1, sync) // 0=OFF 1=NORMAL 2=FULL 3=EXTRA
	})
}

// SyncToDisk folds the WAL back into the main database file, so it needs a real
// WAL to be doing anything at all.
func TestSyncToDiskCheckpointsRealWAL(t *testing.T) {
	db := newFileDB(t)

	id := seedCharacter(t, db, "steam1", 0, 10, "data")
	require.NoError(t, db.UpdateCharacter(t.Context(), id, 20, "updated", 0, 0))
	require.NoError(t, db.SyncToDisk(t.Context()))

	c, err := db.GetCharacter(t.Context(), id)
	require.NoError(t, err)
	require.Equal(t, "updated", c.Data.Data)
}
