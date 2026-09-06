package coalesce

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/msrevive/nexus2/internal/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// recorder is a fake ApplyFunc that records every batch it is handed and can be
// told to fail for specific character IDs.
type recorder struct {
	mu sync.Mutex

	batches  [][]Entry
	applied  map[uuid.UUID]Update
	failWith map[uuid.UUID]error
}

func newRecorder() *recorder {
	return &recorder{
		applied:  make(map[uuid.UUID]Update),
		failWith: make(map[uuid.UUID]error),
	}
}

// apply mimics a real backend: the whole batch rolls back if any entry fails.
func (r *recorder) apply(_ context.Context, batch []Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.batches = append(r.batches, append([]Entry(nil), batch...))

	for _, e := range batch {
		if err, bad := r.failWith[e.ID]; bad {
			return err
		}
	}
	for _, e := range batch {
		r.applied[e.ID] = e.Update
	}
	return nil
}

func (r *recorder) fail(id uuid.UUID, err error) {
	r.mu.Lock()
	r.failWith[id] = err
	r.mu.Unlock()
}

func (r *recorder) heal(id uuid.UUID) {
	r.mu.Lock()
	delete(r.failWith, id)
	r.mu.Unlock()
}

func (r *recorder) get(id uuid.UUID) (Update, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.applied[id]
	return u, ok
}

func (r *recorder) batchCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.batches)
}

// newBuffer builds a buffer with an interval long enough that nothing flushes
// on its own — tests drive Flush explicitly unless they are testing a trigger.
func newBuffer(t *testing.T, rec *recorder, cfg Config) *Buffer {
	t.Helper()
	if cfg.Interval == 0 {
		cfg.Interval = time.Hour
	}
	return New(cfg, rec.apply)
}

func upd(data string) Update {
	return Update{Size: len(data), Data: data}
}

// ─── Coalescing ──────────────────────────────────────────────────────────────

func TestQueue_CoalescesRepeatedUpdates(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{})

	id := uuid.New()
	for i := range 50 {
		b.Queue(id, upd(string(rune('a'+i%26))))
	}
	require.Equal(t, 1, b.Len(), "repeated updates for one character must collapse")

	require.NoError(t, b.Flush(context.Background()))
	require.Equal(t, 1, rec.batchCount(), "50 updates should produce exactly 1 write")

	got, ok := rec.get(id)
	require.True(t, ok)
	require.Equal(t, upd(string(rune('a'+49%26))), got, "the last update wins")
}

func TestFlush_EmptyIsNoop(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{})

	require.NoError(t, b.Flush(context.Background()))
	require.Zero(t, rec.batchCount())
}

func TestFlush_ClearsPending(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{})

	b.Queue(uuid.New(), upd("x"))
	require.NoError(t, b.Flush(context.Background()))
	require.Zero(t, b.Len())
}

// ─── Peek / Drop ─────────────────────────────────────────────────────────────

func TestPeek_ReturnsQueuedState(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{})

	id := uuid.New()
	_, ok := b.Peek(id)
	require.False(t, ok)

	b.Queue(id, upd("fresh"))
	got, ok := b.Peek(id)
	require.True(t, ok)
	require.Equal(t, "fresh", got.Data)

	require.NoError(t, b.Flush(context.Background()))
	_, ok = b.Peek(id)
	require.False(t, ok, "a flushed entry is no longer pending")
}

func TestDrop_DiscardsWithoutApplying(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{})

	id := uuid.New()
	b.Queue(id, upd("doomed"))
	b.Drop(id)

	require.Zero(t, b.Len())
	require.NoError(t, b.Flush(context.Background()))
	require.Zero(t, rec.batchCount(), "dropped entries must never reach the backend")
}

// ─── Batching ────────────────────────────────────────────────────────────────

func TestFlush_SplitsIntoBatches(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{BatchSize: 4, Threshold: -1})

	for range 10 {
		b.Queue(uuid.New(), upd("x"))
	}
	require.NoError(t, b.Flush(context.Background()))

	// 10 entries at 4 per transaction: 4 + 4 + 2.
	require.Equal(t, 3, rec.batchCount())
	require.Len(t, rec.applied, 10)
}

func TestFlush_BatchesAreSortedByID(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Threshold: -1})

	for range 32 {
		b.Queue(uuid.New(), upd("x"))
	}
	require.NoError(t, b.Flush(context.Background()))

	require.Equal(t, 1, rec.batchCount())
	batch := rec.batches[0]
	for i := 1; i < len(batch); i++ {
		require.Negative(t, compareIDs(batch[i-1].ID, batch[i].ID),
			"entries must be ordered so row locks are taken consistently")
	}
}

func compareIDs(a, b uuid.UUID) int {
	for i := range a {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}

// ─── Failure handling ────────────────────────────────────────────────────────

// The behaviour this whole package exists for: one character that can never be
// written must not cost every other character the save they had queued.
func TestFlush_PoisonEntryDoesNotLoseTheBatch(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Threshold: -1})

	poison := uuid.New()
	good := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	rec.fail(poison, database.ErrNoDocument)

	b.Queue(poison, upd("gone"))
	for _, id := range good {
		b.Queue(id, upd("keep"))
	}

	// The flush reports success: nothing recoverable failed.
	require.NoError(t, b.Flush(context.Background()))

	for _, id := range good {
		got, ok := rec.get(id)
		require.True(t, ok, "healthy character %s lost its update", id)
		require.Equal(t, "keep", got.Data)
	}

	_, ok := rec.get(poison)
	require.False(t, ok)
	require.Zero(t, b.Len(), "an entry whose row is gone must be discarded, not retried forever")
}

func TestFlush_TransientFailureRequeues(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Threshold: -1})

	id := uuid.New()
	boom := errors.New("connection reset")
	rec.fail(id, boom)

	b.Queue(id, upd("v1"))

	err := b.Flush(context.Background())
	require.ErrorIs(t, err, boom)
	require.Equal(t, 1, b.Len(), "a transient failure must keep the entry for the next tick")

	// Once the backend recovers, the next flush lands it.
	rec.heal(id)
	require.NoError(t, b.Flush(context.Background()))

	got, ok := rec.get(id)
	require.True(t, ok)
	require.Equal(t, "v1", got.Data)
	require.Zero(t, b.Len())
}

func TestFlush_RequeueDoesNotClobberNewerUpdate(t *testing.T) {
	rec := newRecorder()

	id := uuid.New()
	boom := errors.New("connection reset")

	var b *Buffer
	// Queue a newer update from inside apply, i.e. while the flush is in flight,
	// exactly as a request handler would.
	rec.failWith[id] = boom
	b = New(Config{Interval: time.Hour, Threshold: -1}, func(ctx context.Context, batch []Entry) error {
		b.Queue(id, upd("v2"))
		return rec.apply(ctx, batch)
	})

	b.Queue(id, upd("v1"))
	require.ErrorIs(t, b.Flush(context.Background()), boom)

	got, ok := b.Peek(id)
	require.True(t, ok)
	require.Equal(t, "v2", got.Data, "the newer update must survive the requeue of the older one")
}

func TestFlush_MixedFailuresOnlyRetryTheRecoverable(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Threshold: -1})

	gone := uuid.New()
	flaky := uuid.New()
	fine := uuid.New()

	boom := errors.New("deadlock detected")
	rec.fail(gone, database.ErrNoDocument)
	rec.fail(flaky, boom)

	b.Queue(gone, upd("a"))
	b.Queue(flaky, upd("b"))
	b.Queue(fine, upd("c"))

	require.ErrorIs(t, b.Flush(context.Background()), boom)

	_, ok := rec.get(fine)
	require.True(t, ok, "the healthy entry committed on the per-entry retry")

	require.Equal(t, 1, b.Len())
	_, ok = b.Peek(flaky)
	require.True(t, ok, "only the retryable entry is kept")
}

// ─── Triggers ────────────────────────────────────────────────────────────────

func TestThreshold_FlushesWithoutWaitingForTheInterval(t *testing.T) {
	rec := newRecorder()
	// An hour-long interval: if this test passes, the threshold is what fired.
	b := newBuffer(t, rec, Config{Interval: time.Hour, Threshold: 8})
	b.Start()
	t.Cleanup(func() { _ = b.Stop(context.Background()) })

	for range 8 {
		b.Queue(uuid.New(), upd("x"))
	}

	require.Eventually(t, func() bool {
		return len(rec.applied) == 8
	}, 2*time.Second, 5*time.Millisecond, "reaching the threshold should trigger a flush")
}

func TestInterval_FlushesOnTick(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Interval: 20 * time.Millisecond, Threshold: -1})
	b.Start()
	t.Cleanup(func() { _ = b.Stop(context.Background()) })

	id := uuid.New()
	b.Queue(id, upd("ticked"))

	require.Eventually(t, func() bool {
		_, ok := rec.get(id)
		return ok
	}, 2*time.Second, 5*time.Millisecond)
}

func TestStop_PerformsFinalFlush(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Threshold: -1})
	b.Start()

	id := uuid.New()
	b.Queue(id, upd("last words"))

	require.NoError(t, b.Stop(context.Background()))

	got, ok := rec.get(id)
	require.True(t, ok, "queued updates must not be lost on shutdown")
	require.Equal(t, "last words", got.Data)
}

func TestStop_IsIdempotent(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Threshold: -1})
	b.Start()

	require.NoError(t, b.Stop(context.Background()))
	require.NoError(t, b.Stop(context.Background()), "a second Stop must not panic on the closed channel")
}

// ─── Defaults ────────────────────────────────────────────────────────────────

func TestNew_AppliesDefaults(t *testing.T) {
	b := New(Config{}, func(context.Context, []Entry) error { return nil })

	require.Equal(t, DefaultInterval, b.cfg.Interval)
	require.Equal(t, DefaultThreshold, b.cfg.Threshold)
	require.Equal(t, DefaultBatchSize, b.cfg.BatchSize)
	require.NotNil(t, b.cfg.Logger, "a nil logger must be replaced, not dereferenced")
}

func TestNew_NegativeDisablesThresholdAndBatching(t *testing.T) {
	rec := newRecorder()
	b := New(Config{Interval: time.Hour, Threshold: -1, BatchSize: -1}, rec.apply)

	for range 500 {
		b.Queue(uuid.New(), upd("x"))
	}
	require.Equal(t, 500, b.Len(), "a negative threshold must not trigger a flush")

	require.NoError(t, b.Flush(context.Background()))
	require.Equal(t, 1, rec.batchCount(), "a negative batch size means one transaction")
}

// ─── Concurrency ─────────────────────────────────────────────────────────────

func TestBuffer_ConcurrentQueueAndFlush(t *testing.T) {
	rec := newRecorder()
	b := newBuffer(t, rec, Config{Interval: time.Millisecond, Threshold: 16})
	b.Start()

	ids := make([]uuid.UUID, 32)
	for i := range ids {
		ids[i] = uuid.New()
	}

	var wg sync.WaitGroup
	for w := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				id := ids[(w*100+i)%len(ids)]
				b.Queue(id, upd("x"))
				if i%25 == 0 {
					b.Peek(id)
					b.Len()
				}
			}
		}()
	}
	wg.Wait()

	require.NoError(t, b.Stop(context.Background()))
	require.Zero(t, b.Len())
	require.Len(t, rec.applied, len(ids), "every character must end up written exactly once")
}
