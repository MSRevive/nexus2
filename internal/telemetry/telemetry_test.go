package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/msrevive/nexus2/internal/config"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/oida"
)

func enabledConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Telemetry.Enabled = true

	return cfg
}

// A completed span must never measure as zero. oida's timeline renders
// Duration == 0 as "open", so a zero here shows finished work as still running.
func TestSpansNeverMeasureZero(t *testing.T) {
	tracer, err := New(enabledConfig(), nil)
	require.NoError(t, err)
	require.NotNil(t, tracer)

	// Tight loop: without the monotonic clock these land inside one Windows tick.
	for i := 0; i < 2000; i++ {
		err := tracer.Observe(context.Background(), "fast work", func(ctx context.Context) error {
			_, span := oida.Start(ctx, "instant", oida.KindDatabase)
			span.End()
			return nil
		})
		require.NoError(t, err)
	}

	traces := tracer.Traces()
	require.NotEmpty(t, traces)

	for _, trace := range traces {
		require.NotZero(t, trace.Duration, "trace %q measured zero, renders as still running", trace.Name)
		for _, span := range trace.Spans {
			require.NotZero(t, span.Duration, "span %q measured zero, renders as \"open\"", span.Name)
		}
	}
}

// Every trace must leave the live set once the work returns.
func TestObserveLeavesNothingInFlight(t *testing.T) {
	tracer, err := New(enabledConfig(), nil)
	require.NoError(t, err)

	require.NoError(t, tracer.Observe(context.Background(), "work", func(ctx context.Context) error {
		return nil
	}))

	require.Empty(t, tracer.Live())
}

func TestMonotonicNowAlwaysAdvances(t *testing.T) {
	prev := monotonicNow()
	for i := 0; i < 10000; i++ {
		now := monotonicNow()
		require.True(t, now.After(prev), "clock went backwards or stalled at iteration %d", i)
		prev = now
	}
}

// Telemetry stays off unless it's switched on, so upgrading changes nothing.
func TestDisabledReturnsNilTracer(t *testing.T) {
	tracer, err := New(&config.Config{}, nil)
	require.NoError(t, err)
	require.Nil(t, tracer)

	// A nil tracer has to tolerate the calls the rest of the codebase makes.
	require.Empty(t, Path(tracer))
	require.NoError(t, tracer.Observe(context.Background(), "noop", func(ctx context.Context) error {
		_, span := oida.Start(ctx, "noop", oida.KindDatabase)
		defer span.End()
		return nil
	}))
}

func TestConfigOverridesDefaults(t *testing.T) {
	cfg := enabledConfig()
	cfg.Telemetry.Path = "/telemetry"
	cfg.Telemetry.RingBufferSize = 7
	cfg.Telemetry.SampleRate = 50
	cfg.Telemetry.IgnorePaths = []string{"/skip"}

	tracer, err := New(cfg, nil)
	require.NoError(t, err)

	opts := tracer.Options()
	require.Equal(t, "/telemetry", opts.Path)
	require.Equal(t, "/telemetry", Path(tracer))
	require.Equal(t, 7, opts.RingBufferSize)
	require.Equal(t, float64(50), opts.SampleRate)
	require.Contains(t, opts.IgnorePaths, "/skip")
}

func TestClockShimUnderConcurrency(t *testing.T) {
	const workers = 8
	const each = 500

	seen := make(chan time.Time, workers*each)
	done := make(chan struct{})

	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				seen <- monotonicNow()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < workers; i++ {
		<-done
	}
	close(seen)

	unique := make(map[int64]struct{}, workers*each)
	for ts := range seen {
		key := ts.UnixNano()
		_, dup := unique[key]
		require.False(t, dup, "clock handed out the same instant twice")
		unique[key] = struct{}{}
	}
}
