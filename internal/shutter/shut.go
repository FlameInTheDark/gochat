package shutter

import (
	"log/slog"
	"sync"
)

type Shut struct {
	to  []Shutter
	tof []func()

	mx sync.Mutex

	log *slog.Logger
}

// NewShutter creates a new shutter instance. If the log is nil, then discard logger will be used.
func NewShutter(log *slog.Logger) *Shut {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Shut{
		log: log.With(slog.String("module", "shutter")),
	}
}

// Up add thing that requires shutdown
func (s *Shut) Up(to ...Shutter) {
	if len(to) == 0 {
		return
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	s.to = append(s.to, to...)
}

// UpFunc add blank shutdown function
func (s *Shut) UpFunc(f ...func()) {
	if len(f) == 0 {
		return
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	s.tof = append(s.tof, f...)
}

// Down walks a shutdown list in reverse and call Close() one by one
func (s *Shut) Down() {
	s.mx.Lock()
	defer s.mx.Unlock()
	for i := len(s.to) - 1; i >= 0; i-- {
		err := s.to[i].Close()
		if err != nil {
			s.log.Error("Failed to stop", slog.String("error", err.Error()))
		}
	}
	for i := len(s.tof) - 1; i >= 0; i-- {
		s.tof[i]()
	}
}

// Shutter is an interface for something that can be shutdown. Same as io.Closer interface.
type Shutter interface {
	Close() error
}
