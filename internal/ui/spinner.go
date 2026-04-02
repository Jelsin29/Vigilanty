package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unicode/utf8"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	message   string
	done      chan struct{}
	mu        sync.Mutex
	stopOnce  sync.Once
	startedAt time.Time
	enabled   bool
	lastWidth int
	wg        sync.WaitGroup
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan struct{}),
		enabled: StdoutIsTTY(),
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	s.startedAt = time.Now()
	s.mu.Unlock()

	if !s.enabled {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		frame := 0
		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.render(fmt.Sprintf("%s Running %s...", spinnerFrames[frame], s.message))
				frame = (frame + 1) % len(spinnerFrames)
			}
		}
	}()
}

func (s *Spinner) Stop(result string) {
	duration := s.duration()
	s.stopOnce.Do(func() {
		close(s.done)
	})
	s.wg.Wait()

	finalLine := fmt.Sprintf("%s %s (%s)", result, s.message, duration.Round(100*time.Millisecond))
	if !s.enabled {
		fmt.Fprintln(os.Stdout, finalLine)
		return
	}

	s.render(finalLine)
	fmt.Fprintln(os.Stdout)
}

func (s *Spinner) duration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.startedAt.IsZero() {
		return 0
	}
	return time.Since(s.startedAt)
}

func (s *Spinner) render(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	width := utf8.RuneCountInString(line)
	padding := ""
	for i := width; i < s.lastWidth; i++ {
		padding += " "
	}

	fmt.Fprintf(os.Stdout, "\r%s%s", line, padding)
	s.lastWidth = width
}

func StdoutIsTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
