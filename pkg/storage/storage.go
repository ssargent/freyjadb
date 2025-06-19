package storage

import (
	"github.com/cockroachdb/pebble"
	"github.com/segmentio/ksuid"
)

type DefaultStorage struct {
	db *pebble.DB
}

func NewDefaultStorage(path string) (*DefaultStorage, error) {
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &DefaultStorage{db: db}, nil
}

func (s *DefaultStorage) Create(data []byte) (*ksuid.KSUID, error) {
	id := ksuid.New()
	key := id.Bytes()
	if err := s.db.Set(key, data, pebble.NoSync); err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *DefaultStorage) Read(id *ksuid.KSUID) ([]byte, error) {
	data, closer, err := s.db.Get(id.Bytes())
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	return data, nil
}

func (s *DefaultStorage) Update(id *ksuid.KSUID, data []byte) error {
	return s.db.Set(id.Bytes(), data, pebble.NoSync)
}

func (s *DefaultStorage) Delete(id *ksuid.KSUID) error {
	return s.db.Delete(id.Bytes(), pebble.NoSync)
}

func (s *DefaultStorage) Close() error {
	return s.db.Close()
}
