// libs/progress.go

package libs

import (
	"strings"
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
	if len(h.buffer) == 0 {
		h.mu.Unlock()
		return
	}

	// Объединяем сразу под блокировкой
	latest := h.buffer[len(h.buffer)-1]
	fullText := strings.Join(uniqueLines(h.buffer), "\n")
	h.buffer = h.buffer[:0]
	h.mu.Unlock()

	if h.callback != nil {
		h.callback(fullText)
		h.lastSent = latest
	}
}

func uniqueLines(lines []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}

	return result
}
