// Package coalesce implements the pending-updates buffer shared by every
// database backend.
//
// Character saves arrive far more often than they need to be durable: a game
// server may push the same character dozens of times a minute. Rather than
// writing each one, UpdateCharacter hands the payload to a Buffer, which keeps
// only the latest state per character and commits the whole set periodically.
// N updates for one character between two flushes become exactly one write.
//
// The buffer lives here rather than in each backend because it was previously
// copy-pasted into the sqlite and postgres drivers and the two copies drifted
// apart — different intervals, different locks, and opposite behaviour when a
// flush failed.
package coalesce

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/msrevive/nexus2/internal/database"

	"github.com/google/uuid"
)

const (
	// DefaultInterval is how long an update may sit unwritten. It doubles as the
	// worst-case data loss window if the process dies, so it is deliberately
	// short enough that a player loses at most a few seconds of progress.
	DefaultInterval = 5 * time.Second

	// DefaultThreshold bounds memory: the buffer holds a full save payload per
	// character, so a busy server with no cap would grow with the player count.
	// Reaching this many pending entries triggers a flush without waiting.
	DefaultThreshold = 256

	// DefaultBatchSize caps how many characters are committed in one
	// transaction. On SQLite every write goes through a single connection, so an
	// unbounded flush would block all other writes for the length of the batch.
	DefaultBatchSize = 256
)

// Update is the state queued for one character.
type Update struct {
	Size       int
	Data       string
	BackupMax  int
	BackupTime time.Duration
}

// Entry pairs a character ID with its queued state.
type Entry struct {
	ID     uuid.UUID
	Update Update
}

// ApplyFunc commits a batch of entries. The backend is expected to run the
// whole batch inside one transaction and roll it back entirely on error; the
// buffer handles recovering the entries that were not at fault.
//
// An entry whose character row no longer exists must fail with
// database.ErrNoDocument so the buffer can tell "gone forever" apart from
// "try again later".
type ApplyFunc func(ctx context.Context, batch []Entry) error

type Config struct {
	// Interval between automatic flushes. Zero means DefaultInterval.
	Interval time.Duration
	// Threshold pending entries triggers an immediate flush. Zero means
	// DefaultThreshold; negative disables the trigger.
	Threshold int
	// BatchSize entries per transaction. Zero means DefaultBatchSize; negative
	// means unlimited.
	BatchSize int
	// Name identifies the backend in log messages.
	Name string
	// Logger may be nil, in which case the buffer stays quiet.
	Logger *slog.Logger
}

type Buffer struct {
	cfg   Config
	apply ApplyFunc

	mu      sync.RWMutex
	pending map[uuid.UUID]Update

	nudge chan struct{}
	done  chan struct{}
	once  sync.Once
	wg    sync.WaitGroup
}

func New(cfg Config, apply ApplyFunc) *Buffer {
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultInterval
	}
	if cfg.Threshold == 0 {
		cfg.Threshold = DefaultThreshold
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = DefaultBatchSize
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}

	return &Buffer{
		cfg:     cfg,
		apply:   apply,
		pending: make(map[uuid.UUID]Update),
		nudge:   make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
}

// Start launches the background flush worker.
func (b *Buffer) Start() {
	b.wg.Add(1)
	go b.worker()
}

// Stop shuts the worker down and performs one final flush, so nothing queued is
// lost on a clean exit.
//
// The final flush runs on the caller's goroutine, after the worker is gone.
// That ordering matters for backends whose write path is itself a goroutine:
// the caller can stop the buffer first, while its writer is still alive to
// serve the flush, and only then tear the writer down.
func (b *Buffer) Stop(ctx context.Context) error {
	var err error
	b.once.Do(func() {
		close(b.done)
		b.wg.Wait()
		err = b.Flush(ctx)
	})
	return err
}

// Queue stores the latest state for a character, replacing anything already
// queued for it, and returns the resulting pending count.
func (b *Buffer) Queue(id uuid.UUID, u Update) int {
	b.mu.Lock()
	b.pending[id] = u
	n := len(b.pending)
	b.mu.Unlock()

	if b.cfg.Threshold > 0 && n >= b.cfg.Threshold {
		// Non-blocking: one queued nudge is as good as ten.
		select {
		case b.nudge <- struct{}{}:
		default:
		}
	}

	return n
}

// Peek returns the queued state for a character, if any. Reads use this to
// report the latest accepted write rather than the last committed one.
func (b *Buffer) Peek(id uuid.UUID) (Update, bool) {
	b.mu.RLock()
	u, ok := b.pending[id]
	b.mu.RUnlock()
	return u, ok
}

// Drop discards queued state without applying it. Callers use this when an
// operation invalidates whatever was pending — the character is being deleted,
// or rolled back to an older version that the queued update would undo.
func (b *Buffer) Drop(ids ...uuid.UUID) {
	if len(ids) == 0 {
		return
	}

	b.mu.Lock()
	for _, id := range ids {
		delete(b.pending, id)
	}
	b.mu.Unlock()
}

// Len reports how many characters are currently queued.
func (b *Buffer) Len() int {
	b.mu.RLock()
	n := len(b.pending)
	b.mu.RUnlock()
	return n
}

// Flush commits everything currently queued.
//
// The pending map is swapped out first so callers can keep writing while the
// flush runs. Entries that fail transiently are merged back in; entries whose
// character no longer exists are dropped.
func (b *Buffer) Flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.pending) == 0 {
		b.mu.Unlock()
		return nil
	}
	snapshot := b.pending
	b.pending = make(map[uuid.UUID]Update, len(snapshot))
	b.mu.Unlock()

	entries := make([]Entry, 0, len(snapshot))
	for id, u := range snapshot {
		entries = append(entries, Entry{ID: id, Update: u})
	}

	// Sort so row locks are always taken in the same order. Postgres needs this
	// to avoid deadlocking two overlapping flushes against each other; on SQLite
	// it just makes batches reproducible.
	sort.Slice(entries, func(i, j int) bool {
		return bytes.Compare(entries[i].ID[:], entries[j].ID[:]) < 0
	})

	size := b.cfg.BatchSize
	if size <= 0 || size > len(entries) {
		size = len(entries)
	}

	var firstErr error
	for start := 0; start < len(entries); start += size {
		batch := entries[start:min(start+size, len(entries))]

		if err := b.apply(ctx, batch); err == nil {
			continue
		}

		// The batch rolled back as a whole, so we no longer know which entry was
		// at fault. Retry them one at a time: a single un-appliable character
		// must not cost every other player the save they had queued.
		if err := b.applyIndividually(ctx, batch); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// applyIndividually retries a failed batch entry by entry, separating the
// permanently broken from the merely unlucky.
func (b *Buffer) applyIndividually(ctx context.Context, batch []Entry) error {
	var (
		requeue  []Entry
		firstErr error
	)

	for i := range batch {
		err := b.apply(ctx, batch[i:i+1])

		switch {
		case err == nil:
			// Committed.

		case errors.Is(err, database.ErrNoDocument):
			// The character row is gone — hard-deleted or garbage collected
			// while this update sat in the buffer. Retrying can never succeed,
			// and keeping it would fail every flush from here on.
			b.cfg.Logger.Warn("coalesce: discarding update for missing character",
				"backend", b.cfg.Name, "uuid", batch[i].ID.String(), "error", err)

		default:
			requeue = append(requeue, batch[i])
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if len(requeue) > 0 {
		b.requeue(requeue)
	}

	return firstErr
}

// requeue merges entries back into the pending map after a transient failure.
// An update written since the snapshot was taken is newer, so it wins.
func (b *Buffer) requeue(entries []Entry) {
	b.mu.Lock()
	for _, e := range entries {
		if _, exists := b.pending[e.ID]; !exists {
			b.pending[e.ID] = e.Update
		}
	}
	b.mu.Unlock()
}

func (b *Buffer) worker() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flushLogged()

		case <-b.nudge:
			b.flushLogged()

		case <-b.done:
			// The final flush belongs to Stop, which owns the context and knows
			// the caller's write path is still up.
			return
		}
	}
}

// flushLogged runs a background flush. There is no request behind it, so the
// error has nowhere to go but the log.
func (b *Buffer) flushLogged() {
	if err := b.Flush(context.Background()); err != nil {
		b.cfg.Logger.Error("coalesce: flush error", "backend", b.cfg.Name, "error", err)
	}
}
