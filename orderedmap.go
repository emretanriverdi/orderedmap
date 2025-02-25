package orderedmap

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"sort"
)

type orderedMap[K comparable, V any] struct {
	keys   []K
	values map[K]V
}

func New[K comparable, V any]() *orderedMap[K, V] { // intentionally hidden
	return NewWithCapacity[K, V](16) // pre-allocate
}

func NewWithCapacity[K comparable, V any](capacity int) *orderedMap[K, V] {
	return &orderedMap[K, V]{
		keys:   make([]K, 0, capacity),
		values: make(map[K]V, capacity),
	}
}

func (om *orderedMap[K, V]) Set(key K, value V) {
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
	}
	om.values[key] = value
}

func (om *orderedMap[K, V]) Get(key K) (V, error) {
	if value, exists := om.values[key]; exists {
		return value, nil
	}
	var zero V
	return zero, errors.New("key not found")
}

func (om *orderedMap[K, V]) GetOrEmpty(key K) V {
	if value, exists := om.values[key]; exists {
		return value
	}
	var zero V
	return zero
}

func (om *orderedMap[K, V]) Delete(key K) {
	if _, exists := om.values[key]; !exists {
		return
	}
	for i, k := range om.keys {
		if k == key {
			om.keys = append(om.keys[:i], om.keys[i+1:]...)
			break
		}
	}
	delete(om.values, key)
}

func (om *orderedMap[K, V]) ForEach(f func(K, V)) {
	for _, key := range om.keys {
		f(key, om.values[key])
	}
}

func (om *orderedMap[K, V]) ForEachReverse(f func(K, V)) {
	for i := len(om.keys) - 1; i >= 0; i-- {
		f(om.keys[i], om.values[om.keys[i]])
	}
}

func (om *orderedMap[K, V]) Clear() {
	om.keys = om.keys[:0]
	om.values = make(map[K]V, len(om.values))
}

func (om *orderedMap[K, V]) Keys() []K {
	keysCopy := make([]K, len(om.keys))
	copy(keysCopy, om.keys)
	return keysCopy
}

func (om *orderedMap[K, V]) Values() []V {
	valuesCopy := make([]V, len(om.keys))
	for i, key := range om.keys {
		valuesCopy[i] = om.values[key]
	}
	return valuesCopy
}

func (om *orderedMap[K, V]) Reverse() {
	for i := 0; i < len(om.keys)/2; i++ {
		j := len(om.keys) - 1 - i
		om.keys[i], om.keys[j] = om.keys[j], om.keys[i]
	}
}

func (om *orderedMap[K, V]) ContainsKey(key K) bool {
	_, exists := om.values[key]
	return exists
}

func (om *orderedMap[K, V]) ContainsValue(value V, equal func(a, b V) bool) bool {
	for _, v := range om.values {
		if equal(v, value) {
			return true
		}
	}
	return false
}

func (om *orderedMap[K, V]) ContainsValueReflect(value V) bool {
	return om.ContainsValue(value, func(a, b V) bool {
		return reflect.DeepEqual(a, b)
	})
}

func (om *orderedMap[K, V]) IndexOf(key K) int {
	for i, k := range om.keys {
		if k == key {
			return i
		}
	}
	return -1
}

func (om *orderedMap[K, V]) Pop(key K) (V, bool) {
	value, exists := om.values[key]
	if !exists {
		var zero V
		return zero, false
	}
	om.Delete(key)
	return value, true
}

func (om *orderedMap[K, V]) Clone() *orderedMap[K, V] {
	newMap := &orderedMap[K, V]{
		keys:   make([]K, len(om.keys)),
		values: make(map[K]V, len(om.values)),
	}
	copy(newMap.keys, om.keys)
	for k, v := range om.values {
		newMap.values[k] = v
	}
	return newMap
}

func (om *orderedMap[K, V]) Merge(other *orderedMap[K, V]) {
	for _, key := range other.keys {
		om.Set(key, other.values[key]) // if a key exists, update without changing location
	}
}

func (om *orderedMap[K, V]) SortAsc() error {
	return om.sortByKey(true)
}

func (om *orderedMap[K, V]) SortDesc() error {
	return om.sortByKey(false)
}

func (om *orderedMap[K, V]) sortByKey(isAsc bool) error {
	if err := om.validateKey(); err != nil {
		return err
	}
	sort.SliceStable(om.keys, func(i, j int) bool {
		keyI, keyJ := any(om.keys[i]).(string), any(om.keys[j]).(string) // safe to cast after validateKey.
		if isAsc {
			return keyI < keyJ
		}
		return keyI > keyJ
	})
	return nil
}

func (om *orderedMap[K, V]) MarshalJSON() ([]byte, error) {
	if err := om.validateKey(); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Grow(64 * len(om.keys)) // heuristic pre-allocation
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
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valueBytes, err := json.Marshal(om.values[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (om *orderedMap[K, V]) UnmarshalJSON(data []byte) error {
	if err := om.validateKey(); err != nil {
		return err
	}
	trimmedData := bytes.TrimSpace(data)
	if bytes.Equal(trimmedData, []byte("null")) {
		om.Clear()
		return nil
	}

	om.Clear()

	dec := json.NewDecoder(bytes.NewReader(data))
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("error reading opening token: %w", err)
	}

	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return fmt.Errorf("error in json: %w", err)
		}

		keyStr, ok := token.(string)
		if !ok {
			return fmt.Errorf("expected key token to be string, got: %T", token)
		}

		key := any(keyStr).(K)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return fmt.Errorf("error in json: %w", err)
		}

		var value V
		switch any(value).(type) {
		case *orderedMap[string, any]:
			nested := New[string, any]()
			if err := json.Unmarshal(raw, nested); err != nil {
				return fmt.Errorf("error unmarshaling nested map for key %s: %w", keyStr, err)
			}
			value = any(nested).(V)
		default:
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("error unmarshaling value for key %s: %w", keyStr, err)
			}
		}
		om.Set(key, value)
	}

	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("error reading closing token: %w", err)
	}
	return nil
}

func (om *orderedMap[K, V]) Len() int {
	return len(om.keys)
}

func (om *orderedMap[K, V]) IsEmpty() bool {
	return om == nil || om.Len() == 0
}

func (om *orderedMap[K, V]) String() string {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range om.keys {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v: %v", key, om.values[key]))
	}
	buf.WriteByte('}')
	return buf.String()
}

func (om *orderedMap[K, V]) validateKey() error {
	var key K
	switch any(key).(type) {
	case string:
		return nil
	default:
		return errors.New("error in json: key must be string")
	}
}
