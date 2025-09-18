package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// SignalHandler manages OS signal handling
type SignalHandler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	signals    chan os.Signal
	stderr     io.Writer
	once       sync.Once
	registered bool
	mu         sync.Mutex
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(stderr io.Writer) *SignalHandler {
	if stderr == nil {
		stderr = os.Stderr
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &SignalHandler{
		ctx:     ctx,
		cancel:  cancel,
		signals: make(chan os.Signal, 1),
		stderr:  stderr,
	}
}

// Start begins listening for interrupt signals
func (s *SignalHandler) Start(onInterrupt func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.registered {
		return
	}

	// Register for interrupt signals
	signal.Notify(s.signals, os.Interrupt, syscall.SIGTERM)
	s.registered = true

	// Start signal handler goroutine
	go s.handleSignals(onInterrupt)
}

// handleSignals processes incoming signals
func (s *SignalHandler) handleSignals(onInterrupt func()) {
	for {
		select {
		case sig := <-s.signals:
			s.handleSignal(sig, onInterrupt)
		case <-s.ctx.Done():
			return
		}
	}
}

// handleSignal processes a single signal
func (s *SignalHandler) handleSignal(sig os.Signal, onInterrupt func()) {
	// Ensure we only handle the first interrupt
	s.once.Do(func() {
		fmt.Fprintf(s.stderr, "\nReceived signal: %v\n", sig)
		fmt.Fprintln(s.stderr, "Interrupting, it might take a while...")

		if onInterrupt != nil {
			onInterrupt()
		}

		s.cancel()
	})
}

// Stop stops listening for signals
func (s *SignalHandler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.registered {
		return
	}

	signal.Stop(s.signals)
	close(s.signals)
	s.registered = false
}

// Context returns the context that will be cancelled on interrupt
func (s *SignalHandler) Context() context.Context {
	return s.ctx
}

// Cancel cancels the context manually
func (s *SignalHandler) Cancel() {
	s.cancel()
}

// IsCancelled returns true if the context has been cancelled
func (s *SignalHandler) IsCancelled() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

// Cleanup ensures proper cleanup of the signal handler
func (s *SignalHandler) Cleanup() {
	s.Stop()
	s.cancel()
}