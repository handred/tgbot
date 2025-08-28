// libs/progress.go

package libs

import (
	"sync"
	"time"
)

func NewThrottledHandler(callback func(string), interval time.Duration) (send func(string), stop func()) {
	if interval <= 0 {
		interval = 3 * time.Second
	}

	handler := &throttledHandler{
		callback: callback,
		interval: interval,
		done:     make(chan bool),
		lastSent: "",
		buffer:   nil,
		mu:       sync.Mutex{},
	}

	handler.ticker = time.NewTicker(interval)
	go handler.run()

	send = func(msg string) {
		handler.mu.Lock()
		handler.buffer = append(handler.buffer, msg)
		handler.mu.Unlock()
	}

	stop = func() {
		close(handler.done)
		time.Sleep(100 * time.Millisecond) // даём время на финальный flush
	}

	return send, stop
}

type throttledHandler struct {
	callback func(string)
	interval time.Duration
	ticker   *time.Ticker
	done     chan bool
	buffer   []string
	lastSent string
	mu       sync.Mutex
}

func (h *throttledHandler) run() {
	defer h.ticker.Stop()
	for {
		select {
		case <-h.ticker.C:
			h.flush()
		case <-h.done:
			h.flush()
			return
		}
	}
}

func (h *throttledHandler) flush() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.buffer) == 0 {
		return
	}

	latest := h.buffer[len(h.buffer)-1]
	h.buffer = h.buffer[:0] // очищаем

	if h.callback != nil && latest != h.lastSent {
		h.callback(latest)
		h.lastSent = latest
	}
}
