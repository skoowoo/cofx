package runtime

import (
	"errors"
	"sync"
)

type flowstore struct {
	sync.RWMutex
	entity map[string]*Flow
}

// store store a kv into flowstore
func (s *flowstore) store(k string, f *Flow) (err error) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.entity[k]
	if ok {
		err = errors.New("repeat to store flow: " + k)
		return
	}
	s.entity[k] = f
	return nil
}

func (s *flowstore) get(k string) (*Flow, error) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.entity[k]
	if !ok {
		return nil, errors.New("not found flow: " + k)
	}
	return v, nil
}
