package session

import (
	"context"
	"time"
)

func (s *Session) WaitForSpeechDone(ctx context.Context, stepID string, fallback time.Duration) {
	ch := make(chan struct{}, 1)
	s.pendingSpeechMu.Lock()
	if s.pendingSpeech == nil {
		s.pendingSpeech = make(map[string]chan struct{})
	}
	s.pendingSpeech[stepID] = ch
	s.pendingSpeechMu.Unlock()

	defer func() {
		s.pendingSpeechMu.Lock()
		delete(s.pendingSpeech, stepID)
		s.pendingSpeechMu.Unlock()
	}()

	select {
	case <-ch:
	case <-time.After(fallback):
	case <-ctx.Done():
	}
}

func (s *Session) ResolveSpeechDone(stepID string) bool {
	s.pendingSpeechMu.Lock()
	ch, ok := s.pendingSpeech[stepID]
	if ok {
		delete(s.pendingSpeech, stepID)
	}
	s.pendingSpeechMu.Unlock()

	if ok {
		ch <- struct{}{}
	}
	return ok
}
