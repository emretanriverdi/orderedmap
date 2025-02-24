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
	values []V
	pos    map[K]int
}

func New[K comparable, V any]() *orderedMap[K, V] { // intentionally hidden
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

func (om *orderedMap[K, V]) ForEachReverse(f func(K, V)) {
	for i := len(om.keys) - 1; i >= 0; i-- {
		f(om.keys[i], om.values[i])
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

func (om *orderedMap[K, V]) Reverse() {
	n := len(om.keys)
	for i := 0; i < n/2; i++ {
		j := n - 1 - i
		om.keys[i], om.keys[j] = om.keys[j], om.keys[i]
		om.values[i], om.values[j] = om.values[j], om.values[i]
		om.pos[om.keys[i]] = i
		om.pos[om.keys[j]] = j
	}
}

func (om *orderedMap[K, V]) ContainsKey(key K) bool {
	_, exists := om.pos[key]
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

func (om *orderedMap[K, V]) Pop(key K) (V, bool) {
	value, err := om.Get(key)
	if err != nil {
		var zero V
		return zero, false
	}
	om.Delete(key)
	return value, true
}

func (om *orderedMap[K, V]) Clone() *orderedMap[K, V] {
	newMap := New[K, V]()
	newMap.keys = append(newMap.keys, om.keys...)
	newMap.values = append(newMap.values, om.values...)
	newMap.pos = make(map[K]int, len(om.pos))
	for key, index := range om.pos {
		newMap.pos[key] = index
	}
	return newMap
}

func (om *orderedMap[K, V]) Merge(other *orderedMap[K, V]) {
	for _, key := range other.Keys() {
		value, _ := other.Get(key)
		om.Set(key, value) // if a key exists, update without changing location
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

	type pair struct {
		key    K
		value  V
		strKey string
	}

	pairs := make([]pair, len(om.keys))
	for i, key := range om.keys {
		strKey := any(key).(string) // safe to cast after validateKey.
		pairs[i] = pair{key: key, value: om.values[i], strKey: strKey}
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		if isAsc {
			return pairs[i].strKey < pairs[j].strKey
		}
		return pairs[i].strKey > pairs[j].strKey
	})

	for i, p := range pairs {
		om.keys[i] = p.key
		om.values[i] = p.value
		om.pos[p.key] = i
	}
	return nil
}

func (om *orderedMap[K, V]) MarshalJSON() ([]byte, error) {
	if err := om.validateKey(); err != nil {
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
	if err := om.validateKey(); err != nil {
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

func (om *orderedMap[K, V]) validateKey() error {
	var key K
	switch any(key).(type) {
	case string:
		return nil
	default:
		return errors.New("error in json: key must be string")
	}
}
