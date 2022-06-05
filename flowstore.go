package scheduler

import (
	"errors"
	"sync"
)

type ChangeNotify interface {
	Added(*Flow)
	Updated(*Flow)
	Deleted(*Flow)
}

// todo
type Persistent interface {
}

type FlowStore struct {
	sync.RWMutex
	notify ChangeNotify
	entity map[string]*Flow
	saver  Persistent
}

// Store store a kv into flowstore
func (s *FlowStore) Store(k string, f *Flow) (err error) {
	s.Lock()
	defer func() {
		s.Unlock()
		if err == nil {
			s.notify.Added(f)
		}
	}()

	_, ok := s.entity[k]
	if ok {
		err = errors.New("repeat to store flow: " + k)
		return
	}
	s.entity[k] = f
	return nil
}

func (s *FlowStore) Update(k string, f *Flow) (err error) {
	s.Lock()
	defer func() {
		s.Unlock()
		if err == nil {
			s.notify.Updated(f)
		}
	}()

	_, ok := s.entity[k]
	if !ok {
		return errors.New("can't update, not found flow: " + k)
	}
	s.entity[k] = f
	return nil
}

func (s *FlowStore) Get(k string) (*Flow, error) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.entity[k]
	if !ok {
		return nil, errors.New("not found flow: " + k)
	}
	return v, nil
}

// If 'do' return a error, will stop the 'Foreach'
func (s *FlowStore) Foreach(do func(string, *Flow) error) {
	s.RLock()
	defer s.RUnlock()
	for k, v := range s.entity {
		if err := do(k, v); err != nil {
			return
		}
	}
}
