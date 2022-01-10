package gsession

import (
	"context"
	"fmt"
)

type MemStorage struct {
	SessionValues map[string]map[interface{}]interface{}
}

func (s *MemStorage) Save(ctx context.Context, id string, values map[interface{}]interface{}) error {
	if s.SessionValues == nil {
		s.SessionValues = make(map[string]map[interface{}]interface{})
	}
	s.SessionValues[id] = values
	return nil
}

func (s *MemStorage) Load(ctx context.Context, id string) (map[interface{}]interface{}, error) {
	values, ok := s.SessionValues[id]
	if !ok {
		return nil, fmt.Errorf("no value found for id: %s", id)
	}
	return values, nil
}

func (s *MemStorage) Delete(ctx context.Context, id string) error {
	delete(s.SessionValues, id)
	return nil
}
