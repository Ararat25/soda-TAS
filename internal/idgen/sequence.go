package idgen

import (
	"fmt"
	"sync"
)

// Sequence реализует интерфейс IDGenerator
type Sequence struct {
	mu       sync.Mutex
	counters map[string]int
}

// New конструктор Sequence
func New() *Sequence { return &Sequence{counters: make(map[string]int)} }

// NextID гененрирует ID в зависимоти от префикса
func (s *Sequence) NextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[prefix]++

	return fmt.Sprintf("%s%d", prefix, s.counters[prefix])
}
