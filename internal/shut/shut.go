package shut

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

func NewShutter(log *slog.Logger) *Shut {
	return &Shut{
		log: log.With(slog.String("module", "shutter")),
	}
}

// Add thing that requires shutdown
func (s *Shut) Up(to ...Shutter) {
	if len(to) == 0 {
		return
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	s.to = append(s.to, to...)
}

func (s *Shut) UpFunc(f ...func()) {
	if len(f) == 0 {
		return
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	s.tof = append(s.tof, f...)
}

// Down walks shutdown list in reverse and call Close() one by one
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

type Shutter interface {
	Close() error
}
