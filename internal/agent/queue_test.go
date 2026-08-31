package agent

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSubmit_ClosedChannelDoesNotPanic(t *testing.T) {
	manager := NewPlanningQueue(nil)

	userID := uuid.New()
	ch := make(chan planJob)
	manager.queues[userID] = ch
	close(ch)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Submit panicked instead of returning nil: %v", r)
		}
	}()

	result := manager.Submit(context.Background(), userID, nil, nil, "", nil)

	if result != nil {
		t.Fatalf("expected nil result for a closed queue, got %+v", result)
	}
}
