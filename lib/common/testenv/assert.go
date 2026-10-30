package testenv

import (
	"testing"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const (
	// WaitTimeout bounds asynchronous expectations: relay, broker and consumers.
	WaitTimeout  = 30 * time.Second
	pollInterval = 50 * time.Millisecond
)

// Eventually waits for condition using the shared timeout.
func Eventually(t *testing.T, condition func() bool, msgAndArgs ...any) {
	t.Helper()
	require.Eventually(t, condition, WaitTimeout, pollInterval, msgAndArgs...)
}

// Events returns envelopes stored in the module outbox in insertion order.
func (a *App) Events(t *testing.T) []event.Envelope {
	t.Helper()
	var rows []struct {
		Payload []byte `db:"payload"`
	}
	err := a.DB.SelectContext(t.Context(), "testenv.outbox", &rows, "SELECT payload FROM event_outbox ORDER BY id")
	require.NoError(t, err)
	envelopes := make([]event.Envelope, 0, len(rows))
	for _, row := range rows {
		envelope, err := event.Unmarshal(row.Payload)
		require.NoError(t, err)
		envelopes = append(envelopes, envelope)
	}
	return envelopes
}

// RequireEvent waits until the module outbox has an event of type PT matching predicate
// and the relay published it.
func RequireEvent[T any, PT interface {
	*T
	proto.Message
}](t *testing.T, app *App, match func(PT) bool) (found event.Envelope, payload PT) {
	t.Helper()
	eventType := string(PT(new(T)).ProtoReflect().Descriptor().Name())
	Eventually(t, func() bool {
		for _, envelope := range app.Events(t) {
			if envelope.Type() != eventType {
				continue
			}
			candidate := PT(new(T))
			require.NoError(t, envelope.UnmarshalPayload(candidate))
			if match(candidate) {
				found, payload = envelope, candidate
				return app.published(t, envelope.ID())
			}
		}
		return false
	}, "event %s was not published", eventType)
	return found, payload
}

// RequireHandled waits until the module consumer processed the event.
func (a *App) RequireHandled(t *testing.T, eventID string) {
	t.Helper()
	a.RequireDelivered(t, eventID, 1)
}

// RequireDelivered waits until the consumer acknowledged the event the given number of times
// and checks that the inbox recorded it exactly once.
func (a *App) RequireDelivered(t *testing.T, eventID string, times int) {
	t.Helper()
	Eventually(t, func() bool {
		return a.Logs.Count("event handling completed", map[string]string{"event_id": eventID, "result": "processed"}) >= times
	}, "event %s was not handled %d times", eventID, times)
	var markers int
	err := a.DB.GetContext(t.Context(), "testenv.inbox", &markers, "SELECT COUNT(*) FROM event_inbox WHERE event_id = ?", eventID)
	require.NoError(t, err)
	require.Equal(t, 1, markers)
}

func (a *App) published(t *testing.T, eventID string) bool {
	var published int
	err := a.DB.GetContext(t.Context(), "testenv.outbox-published", &published,
		"SELECT COUNT(*) FROM event_outbox WHERE event_id = ? AND published_at IS NOT NULL", eventID)
	require.NoError(t, err)
	return published == 1
}
