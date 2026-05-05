package spotify

import (
	"context"
	"sync"
	"time"
)

func NewSlidingWindow(windowSize time.Duration, maxRequests int) *SlidingWindow {
	return &SlidingWindow{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		requests:    []time.Time{},
	}
}

type SlidingWindow struct {
	requests    []time.Time
	windowSize  time.Duration
	maxRequests int
	mu          sync.Mutex
}

// Thread-safe function to wait until a request can be made without exceeding the rate limit.
func (w *SlidingWindow) Wait(ctx context.Context) error {
	w.mu.Lock()
	now := time.Now()
	// Clean old requests
	cutoff := now.Add(-w.windowSize)
	i := 0
	for ; i < len(w.requests); i++ {
		if !w.requests[i].Before(cutoff) {
			break
		}
	}
	w.requests = w.requests[i:]

	// check if there's available slots after cleaning
	if len(w.requests) < w.maxRequests {
		w.requests = append(w.requests, now)
		w.mu.Unlock()
		return nil
	}

	// >= window size requests in the window at this point, shouldn't panic
	allowanceTime := w.requests[len(w.requests)-w.maxRequests].Add(w.windowSize)
	toWait := allowanceTime.Sub(now)
	if toWait <= 0 {
		w.requests = append(w.requests, now)
		w.mu.Unlock()
		return nil
	}

	// calculate waiting time and reserve a slot, so we can unlock before sleeping
	w.requests = append(w.requests, allowanceTime)
	w.mu.Unlock()

	timer := time.NewTimer(toWait)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil

	case <-ctx.Done():
		w.mu.Lock()

		// Remove the reservation we added.
		for i := range w.requests {
			if w.requests[i].Equal(allowanceTime) {
				w.requests = append(w.requests[:i], w.requests[i+1:]...)
				break
			}
		}

		w.mu.Unlock()
		return ctx.Err()
	}
}
