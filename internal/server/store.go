package server

import (
	"context"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
)

type VKStorage struct {
	vk cache.Cache
}

func NewVKStorage(c cache.Cache) *VKStorage {
	return &VKStorage{vk: c}
}

func (s *VKStorage) Get(key string) ([]byte, error) {
	return s.vk.GetBytes(context.Background(), key)
}

// Set stores the given value for the given key along
// with an expiration value, 0 means no expiration.
// Empty key or value will be ignored without an error.
func (s *VKStorage) Set(key string, val []byte, exp time.Duration) error {
	return s.vk.SetTimed(context.Background(), key, string(val), exp.Milliseconds()/1000)
}

// Delete deletes the value for the given key.
// It returns no error if the storage does not contain the key,
func (s *VKStorage) Delete(key string) error {
	return s.vk.Delete(context.Background(), key)
}

// Reset resets the storage and delete all keys.
func (s *VKStorage) Reset() error {
	return nil
}

// Close closes the storage and will stop any running garbage
// collectors and open connections.
func (s *VKStorage) Close() error {
	return nil
}
