package orderedmap

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"sort"
	"sync"
)

var errKeyNotFound = errors.New("key not found")
var errKeyMustBeStringForJson = errors.New("error in json: key must be string")

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type orderedMap[K comparable, V any] struct {
	head        *node[K, V]
	tail        *node[K, V]
	values      map[K]*node[K, V]
	size        int
	pool        *sync.Pool
	isKeyString bool // pre-calculate key's type to check if it's parseable (and save it to avoid multiple calculations)
}

func New[K comparable, V any]() *orderedMap[K, V] { // intentionally hidden
	return NewWithCapacity[K, V](16) // pre-allocate
}

func NewWithCapacity[K comparable, V any](capacity int) *orderedMap[K, V] {
	var zero K
	om := &orderedMap[K, V]{
		values: make(map[K]*node[K, V], capacity),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(node[K, V])
			},
		},
		isKeyString: isKeyString(zero),
	}
	return om
}

func (om *orderedMap[K, V]) Set(key K, value V) {
	if existingNode, exists := om.values[key]; exists {
		existingNode.value = value
		return
	}

	n := om.pool.Get().(*node[K, V])
	n.key = key
	n.value = value
	n.prev = nil
	n.next = nil

	om.values[key] = n
	if om.head == nil {
		om.head = n
		om.tail = n
	} else {
		om.tail.next = n
		n.prev = om.tail
		om.tail = n
	}
	om.size++
}

func (om *orderedMap[K, V]) Get(key K) (V, error) {
	if n, exists := om.values[key]; exists {
		return n.value, nil
	}
	var zero V
	return zero, errKeyNotFound
}

func (om *orderedMap[K, V]) GetOrDefault(key K) V {
	if n, exists := om.values[key]; exists {
		return n.value
	}
	var zero V
	return zero
}

func (om *orderedMap[K, V]) Delete(key K) {
	if n, exists := om.values[key]; exists {
		if n == om.head {
			om.head = n.next
		} else {
			n.prev.next = n.next
		}
		if n == om.tail {
			om.tail = n.prev
		} else if n.next != nil {
			n.next.prev = n.prev
		}
		delete(om.values, key)
		n.prev = nil
		n.next = nil
		om.pool.Put(n)
		om.size--
	}
}

func (om *orderedMap[K, V]) ForEach(f func(K, V)) {
	for n := om.head; n != nil; n = n.next {
		f(n.key, n.value)
	}
}

func (om *orderedMap[K, V]) ForEachReverse(f func(K, V)) {
	for n := om.tail; n != nil; n = n.prev {
		f(n.key, n.value)
	}
}

func (om *orderedMap[K, V]) Clear() {
	for n := om.head; n != nil; {
		next := n.next
		n.prev = nil
		n.next = nil
		om.pool.Put(n)
		n = next
	}
	om.head = nil
	om.tail = nil
	om.values = make(map[K]*node[K, V], om.size)
	om.size = 0
}

func (om *orderedMap[K, V]) Keys() []K {
	keys := make([]K, om.size)
	index := 0
	for n := om.head; n != nil; n = n.next {
		keys[index] = n.key
		index++
	}
	return keys
}

func (om *orderedMap[K, V]) Values() []V {
	values := make([]V, om.size)
	index := 0
	for n := om.head; n != nil; n = n.next {
		values[index] = n.value
		index++
	}
	return values
}

func (om *orderedMap[K, V]) Reverse() {
	om.head, om.tail = om.tail, om.head
	for n := om.head; n != nil; n = n.next {
		n.prev, n.next = n.next, n.prev
	}
}

func (om *orderedMap[K, V]) ContainsKey(key K) bool {
	_, exists := om.values[key]
	return exists
}

func (om *orderedMap[K, V]) ContainsValue(value V, equal func(a, b V) bool) bool {
	for n := om.head; n != nil; n = n.next {
		if equal(n.value, value) {
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
	index := 0
	for n := om.head; n != nil; n = n.next {
		if n.key == key {
			return index
		}
		index++
	}
	return -1
}

func (om *orderedMap[K, V]) Pop(key K) (V, bool) {
	if n, exists := om.values[key]; exists {
		value := n.value
		om.Delete(key)
		return value, true
	}
	var zero V
	return zero, false
}

func (om *orderedMap[K, V]) Clone() *orderedMap[K, V] {
	newMap := NewWithCapacity[K, V](om.size)
	for n := om.head; n != nil; n = n.next {
		newMap.Set(n.key, n.value)
	}
	return newMap
}

func (om *orderedMap[K, V]) Merge(other *orderedMap[K, V]) {
	for n := other.head; n != nil; n = n.next {
		om.Set(n.key, n.value) // if a key exists, update without changing location
	}
}

func (om *orderedMap[K, V]) SortAsc() error {
	return om.sortKeys(func(i, j int, keys []K) bool {
		return any(keys[i]).(string) < any(keys[j]).(string) // safe to cast after validateKey.
	})
}

func (om *orderedMap[K, V]) SortDesc() error {
	return om.sortKeys(func(i, j int, keys []K) bool {
		return any(keys[i]).(string) > any(keys[j]).(string) // safe to cast after validateKey.
	})
}

func (om *orderedMap[K, V]) sortKeys(less func(i, j int, keys []K) bool) error {
	if err := om.validateKey(); err != nil {
		return err
	}
	keys := om.Keys()
	sort.SliceStable(keys, func(i, j int) bool {
		return less(i, j, keys)
	})
	om.rebuildFromKeys(keys)
	return nil
}

func (om *orderedMap[K, V]) rebuildFromKeys(keys []K) {
	om.head = nil
	om.tail = nil
	om.size = 0
	for _, key := range keys {
		if n, exists := om.values[key]; exists {
			n.prev = om.tail
			n.next = nil
			if om.tail != nil {
				om.tail.next = n
			}
			if om.head == nil {
				om.head = n
			}
			om.tail = n
			om.size++
		}
	}
}

func (om *orderedMap[K, V]) MarshalJSON() ([]byte, error) {
	if err := om.validateKey(); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Grow(64 * om.size) // heuristic pre-allocation
	buf.WriteByte('{')
	first := true
	for n := om.head; n != nil; n = n.next {
		if !first {
			buf.WriteByte(',')
		}
		// We can safely cast key to string here because we know it is one.
		keyBytes, err := json.Marshal(any(n.key).(string))
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valueBytes, err := json.Marshal(n.value)
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
		first = false
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (om *orderedMap[K, V]) UnmarshalJSON(data []byte) error {
	if err := om.validateKey(); err != nil {
		return err
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
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
			return fmt.Errorf("expected string key, got: %T", token)
		}

		key := any(keyStr).(K)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return fmt.Errorf("error in json: %w", err)
		}

		var value V // tricky force-casting by checking if value can be treated as orderedmap to use its own unmarshal
		if _, isMap := any(value).(*orderedMap[string, any]); isMap {
			nested := New[string, any]()
			if err := json.Unmarshal(raw, nested); err != nil {
				return fmt.Errorf("error unmarshaling nested map for key %s: %w", keyStr, err)
			}
			value = any(nested).(V)
		} else {
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
	return om.size
}

func (om *orderedMap[K, V]) IsEmpty() bool {
	return om.size == 0
}

func (om *orderedMap[K, V]) String() string {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	for n := om.head; n != nil; n = n.next {
		if !first {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprint(n.key))
		buf.WriteString(": ")
		buf.WriteString(fmt.Sprint(n.value))
		first = false
	}
	buf.WriteByte('}')
	return buf.String()
}

func (om *orderedMap[K, V]) validateKey() error {
	if om.isKeyString {
		return nil
	}
	return errKeyMustBeStringForJson
}

func isKeyString(key any) bool {
	if _, ok := key.(string); ok {
		return true
	}
	return false
}
