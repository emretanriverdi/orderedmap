package orderedmap

import (
	"bytes"
	"encoding/json"
	"errors"
)

type OrderedMap[K comparable, V any] struct {
	keys   []K
	values []V
	pos    map[K]int
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		pos: make(map[K]int),
	}
}

func (om *OrderedMap[K, V]) Set(key K, value V) {
	if i, exists := om.pos[key]; exists {
		om.values[i] = value
		return
	}
	om.pos[key] = len(om.keys)
	om.keys = append(om.keys, key)
	om.values = append(om.values, value)
}

func (om *OrderedMap[K, V]) Get(key K) (V, error) {
	if i, exists := om.pos[key]; exists {
		return om.values[i], nil
	}
	var zero V
	return zero, errors.New("key not found")
}

func (om *OrderedMap[K, V]) GetOrEmpty(key K) V {
	if i, exists := om.pos[key]; exists {
		return om.values[i]
	}
	var zeroValue V
	return zeroValue
}

func (om *OrderedMap[K, V]) Delete(key K) {
	i, exists := om.pos[key]
	if !exists {
		return
	}

	lastIdx := len(om.keys) - 1
	if i != lastIdx {
		om.keys[i], om.keys[lastIdx] = om.keys[lastIdx], om.keys[i]
		om.values[i], om.values[lastIdx] = om.values[lastIdx], om.values[i]
		om.pos[om.keys[i]] = i
	}

	delete(om.pos, key)
	om.keys = om.keys[:lastIdx]
	om.values = om.values[:lastIdx]
}

func (om *OrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, key := range om.keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		keyBytes, err := json.Marshal(key)
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

func (om *OrderedMap[K, V]) UnmarshalJSON(data []byte) error {
	om.keys = om.keys[:0]
	om.values = om.values[:0]
	om.pos = make(map[K]int)

	data = bytes.Trim(data, "{}")
	if len(data) == 0 {
		return nil
	}

	pairs := bytes.Split(data, []byte(","))
	for _, pair := range pairs {
		colonIndex := bytes.IndexByte(pair, ':')
		if colonIndex == -1 {
			return errors.New("malformed JSON: missing colon in key-value pair")
		}

		var key K
		if err := json.Unmarshal(pair[:colonIndex], &key); err != nil {
			return err
		}

		var value V
		if err := json.Unmarshal(pair[colonIndex+1:], &value); err != nil {
			return err
		}
		om.Set(key, value)
	}
	return nil
}

func (om *OrderedMap[K, V]) Keys() []K {
	return om.keys
}

func (om *OrderedMap[K, V]) Values() []V {
	return om.values
}
