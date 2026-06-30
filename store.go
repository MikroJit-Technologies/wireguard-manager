package main

import (
	"sync"
	"time"
)

type storedPeer struct {
	ClientConfig string
	SavedAt      time.Time
}

type configStore struct {
	mu sync.RWMutex
	m  map[string]storedPeer
}

var cfgStore = &configStore{m: make(map[string]storedPeer)}

func (s *configStore) Set(pubkey, clientConfig string) {
	s.mu.Lock()
	s.m[pubkey] = storedPeer{clientConfig, time.Now()}
	s.mu.Unlock()
}

func (s *configStore) Get(pubkey string) (string, bool) {
	s.mu.RLock()
	v, ok := s.m[pubkey]
	s.mu.RUnlock()
	return v.ClientConfig, ok
}

func (s *configStore) startCleaner() {
	go func() {
		for range time.Tick(10 * time.Minute) {
			s.mu.Lock()
			for k, v := range s.m {
				if time.Since(v.SavedAt) > 24*time.Hour {
					delete(s.m, k)
				}
			}
			s.mu.Unlock()
		}
	}()
}
