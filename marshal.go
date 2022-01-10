package gsession

import (
	"encoding/json"
	"fmt"
)

type Marshaler interface {
	Marshal(values map[interface{}]interface{}) ([]byte, error)
	Unmarshal(data []byte) (map[interface{}]interface{}, error)
	ContentType() string
}

type JSONMarshaler struct{}

func (m *JSONMarshaler) Marshal(values map[interface{}]interface{}) ([]byte, error) {
	compatValues := make(map[string]interface{})
	for k, v := range values {
		compatK, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("JSONMarshaler: only string keys supported: %T", k)
		}
		compatValues[compatK] = v
	}

	return json.Marshal(compatValues)
}

func (m *JSONMarshaler) Unmarshal(data []byte) (map[interface{}]interface{}, error) {
	var compatValues map[string]interface{}
	err := json.Unmarshal(data, &compatValues)

	values := make(map[interface{}]interface{})
	for k, v := range compatValues {
		values[k] = v
	}

	return values, err
}

func (m *JSONMarshaler) ContentType() string {
	return "application/json"
}
