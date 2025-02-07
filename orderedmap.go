package orderedmap

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
)

type orderedMap[K comparable, V any] struct {
	keys   []K
	values []V
	pos    map[K]int
}

func NewOrderedMap[K comparable, V any]() *orderedMap[K, V] { // intentionally hidden
	return &orderedMap[K, V]{
		keys:   make([]K, 0),
		values: make([]V, 0),
		pos:    make(map[K]int),
	}
}

func (om *orderedMap[K, V]) Set(key K, value V) {
	if i, exists := om.pos[key]; exists {
		om.values[i] = value
		return
	}
	om.pos[key] = len(om.keys)
	om.keys = append(om.keys, key)
	om.values = append(om.values, value)
}

func (om *orderedMap[K, V]) Get(key K) (V, error) {
	if i, exists := om.pos[key]; exists {
		return om.values[i], nil
	}
	var zero V
	return zero, errors.New("key not found")
}

func (om *orderedMap[K, V]) GetOrEmpty(key K) V {
	if i, exists := om.pos[key]; exists {
		return om.values[i]
	}
	var zero V
	return zero
}

func (om *orderedMap[K, V]) Delete(key K) {
	i, exists := om.pos[key]
	if !exists {
		return
	}

	copy(om.keys[i:], om.keys[i+1:])
	om.keys = om.keys[:len(om.keys)-1]

	copy(om.values[i:], om.values[i+1:])
	om.values = om.values[:len(om.values)-1]

	delete(om.pos, key)

	for j := i; j < len(om.keys); j++ {
		om.pos[om.keys[j]] = j
	}
}

func (om *orderedMap[K, V]) ForEach(f func(K, V)) {
	for i, key := range om.keys {
		f(key, om.values[i])
	}
}

func (om *orderedMap[K, V]) Clear() {
	om.keys = om.keys[:0]
	om.values = om.values[:0]
	om.pos = make(map[K]int)
}

func (om *orderedMap[K, V]) Keys() []K {
	keysCopy := make([]K, len(om.keys))
	copy(keysCopy, om.keys)
	return keysCopy
}

func (om *orderedMap[K, V]) Values() []V {
	valuesCopy := make([]V, len(om.values))
	copy(valuesCopy, om.values)
	return valuesCopy
}

func (om *orderedMap[K, V]) MarshalJSON() ([]byte, error) {
	if err := om.checkKey(); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range om.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyStr := any(key).(string)
		keyBytes, err := json.Marshal(keyStr)
		if err != nil {
			return nil, err
		}
		valueBytes, err := json.Marshal(om.values[i])
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (om *orderedMap[K, V]) UnmarshalJSON(data []byte) error {
	if err := om.checkKey(); err != nil {
		return err
	}

	trimmedData := bytes.TrimSpace(data)
	if bytes.Equal(trimmedData, []byte("null")) {
		om.keys = nil
		om.values = nil
		om.pos = make(map[K]int)
		return nil
	}

	om.Clear()

	dec := json.NewDecoder(bytes.NewReader(data))
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("error reading opening token: %w", err)
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != '{' {
		return errors.New("error in json: expected '{' at the beginning of JSON object")
	}

	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return fmt.Errorf("error reading key token: %w", err)
		}

		keyStr, ok := token.(string)
		if !ok {
			return fmt.Errorf("expected key token to be string, got %T", token)
		}

		var key = any(keyStr).(K)
		var value V
		if err := dec.Decode(&value); err != nil {
			return fmt.Errorf("error decoding value for key %v: %w", key, err)
		}

		om.Set(key, value)
	}
	
	token, err = dec.Token()
	if err != nil {
		return fmt.Errorf("error reading closing token: %w", err)
	}
	delim, ok = token.(json.Delim)
	if !ok || delim != '}' {
		return errors.New("error in json: expected '}' at the end of JSON object")
	}
	return nil
}

func (om *orderedMap[K, V]) Len() int {
	return len(om.keys)
}

func (om *orderedMap[K, V]) String() string {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range om.keys {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v: %v", key, om.values[i]))
	}
	buf.WriteByte('}')
	return buf.String()
}

func (om *orderedMap[K, V]) checkKey() error {
	var key K
	switch any(key).(type) {
	case string:
		return nil
	default:
		return errors.New("error in json: key must be string")
	}
}
