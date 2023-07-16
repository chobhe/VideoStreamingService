package main

import (
	"fmt"
	"net/http"
	"sync"
)

// Map of pubsub objects that have one client streaming/publishing and a bunch of clients subscribing/receiving
type RelayService struct {
	streams    map[string]*Pubsub
	m          sync.Mutex
	httpClient *http.Client
}

// NewRelayService constructs a new RelayService.
// The RelayService keeps track of all active pubsubs (streams) and provides functionality to
// create, retrieve, and remove pubsubs.
func NewRelayService() *RelayService {
	return &RelayService{
		streams: make(map[string]*Pubsub),
	}
}

// NewPubsub creates a new Pubsub associated with the given key and adds it to the RelayService.
// It returns an error if a Pubsub with the same key already exists in the RelayService.
// The function is thread-safe; it locks the RelayService while it's adding the Pubsub.
func (s *RelayService) NewPubsub(key string) (*Pubsub, error) {
	s.m.Lock()
	defer s.m.Unlock()

	// If the stream key is already in the relayservice error out
	if _, ok := s.streams[key]; ok {
		return nil, fmt.Errorf("already published: %s", key)
	}

	pubsub := NewPubsub(s, key)

	s.streams[key] = pubsub

	return pubsub, nil
}

// GetPubsub retrieves the Pubsub associated with the given key from the RelayService.
// It returns an error if no such Pubsub exists.
// The function is thread-safe; it locks the RelayService while it's retrieving the Pubsub.
func (s *RelayService) GetPubsub(key string) (*Pubsub, error) {
	s.m.Lock()
	defer s.m.Unlock()

	pubsub, ok := s.streams[key]
	if !ok {
		return nil, fmt.Errorf("not published: %s", key)
	}

	return pubsub, nil
}

// RemovePubsub removes the Pubsub associated with the given key from the RelayService.
// It returns an error if no such Pubsub exists.
// The function is thread-safe; it locks the RelayService while it's removing the Pubsub.
func (s *RelayService) RemovePubsub(key string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.streams[key]; !ok {
		return fmt.Errorf("not published: %s", key)
	}

	delete(s.streams, key)

	return nil
}
