package orderedmap

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

func TestOrderedMap(t *testing.T) {
	t.Run("basic operations (get-set)", func(t *testing.T) {
		om := New[string, int]()

		// Test Set
		om.Set("a", 1)
		om.Set("b", 2)
		om.Set("c", 3)

		assert.Equal(t, []string{"a", "b", "c"}, om.Keys())
		assert.Equal(t, []int{1, 2, 3}, om.Values())

		// Test Get
		val, err := om.Get("b")
		assert.Nil(t, err)
		assert.Equal(t, 2, val)

		// Test Get non-existent key
		_, err = om.Get("d")
		assert.NotNil(t, err)

		// Test GetOrDefault
		val = om.GetOrDefault("d")
		assert.Equal(t, 0, val)

		// Test Update
		om.Set("b", 5)
		val, _ = om.Get("b")
		assert.Equal(t, 5, val)
	})

	t.Run("delete", func(t *testing.T) {
		om := New[string, int]()
		om.Set("a", 1)
		om.Set("b", 2)
		om.Set("c", 3)

		om.Delete("b")
		_, err := om.Get("b")
		assert.NotNil(t, err)

		assert.Equal(t, []string{"a", "c"}, om.Keys())
		assert.Equal(t, []int{1, 3}, om.Values())
	})

	t.Run("JSON marshaling", func(t *testing.T) {
		om := New[string, int]()
		om.Set("a", 1)
		om.Set("c", 3)

		jsonData, err := json.Marshal(om)
		assert.Nil(t, err)

		expectedJSON := `{"a":1,"c":3}`
		assert.JSONEq(t, expectedJSON, string(jsonData))
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		jsonInput := `{"x":10,"y":20,"z":30}`
		om2 := New[string, int]()

		err := json.Unmarshal([]byte(jsonInput), om2)
		assert.Nil(t, err)

		assert.Equal(t, []string{"x", "y", "z"}, om2.Keys())
		assert.Equal(t, []int{10, 20, 30}, om2.Values())
	})

	t.Run("empty map", func(t *testing.T) {
		emptyOM := New[string, int]()

		emptyJSON, err := json.Marshal(emptyOM)
		assert.Nil(t, err)
		assert.Equal(t, "{}", string(emptyJSON))
	})

	t.Run("single-element", func(t *testing.T) {
		om := New[string, int]()
		om.Set("single", 42)

		oneJSON, err := json.Marshal(om)
		assert.Nil(t, err)
		assert.Contains(t, string(oneJSON), `"single":42`)
	})

	t.Run("nested maps", func(t *testing.T) {
		inner := New[string, int]()
		inner.Set("inner1", 100)
		inner.Set("inner2", 200)

		outer := New[string, *orderedMap[string, int]]()
		outer.Set("outer", inner)

		assert.Equal(t, []string{"outer"}, outer.Keys())

		nestedMap := outer.GetOrDefault("outer")
		assert.Equal(t, []string{"inner1", "inner2"}, nestedMap.Keys())
		assert.Equal(t, []int{100, 200}, nestedMap.Values())

		jsonData, err := json.Marshal(outer)
		assert.Nil(t, err)

		expectedJSON := `{"outer":{"inner1":100,"inner2":200}}`
		assert.JSONEq(t, expectedJSON, string(jsonData))

		outer2 := New[string, *orderedMap[string, int]]()
		err = json.Unmarshal([]byte(expectedJSON), outer2)
		assert.Nil(t, err)

		assert.Equal(t, []string{"outer"}, outer2.Keys())
		nestedMap2 := outer2.GetOrDefault("outer")
		assert.Equal(t, []string{"inner1", "inner2"}, nestedMap2.Keys())
		assert.Equal(t, []int{100, 200}, nestedMap2.Values())
	})

	t.Run("keys and values with commas", func(t *testing.T) {
		omCommaStr := New[string, string]()
		omCommaStr.Set("a,b", "val,1")
		omCommaStr.Set("c,d", "val,2")

		assert.Equal(t, []string{"a,b", "c,d"}, omCommaStr.Keys())
		assert.Equal(t, []string{"val,1", "val,2"}, omCommaStr.Values())

		jsonData, err := json.Marshal(omCommaStr)
		assert.Nil(t, err)

		expectedJSON := `{"a,b":"val,1","c,d":"val,2"}`
		assert.JSONEq(t, expectedJSON, string(jsonData))

		omCommaStr2 := New[string, string]()
		err = json.Unmarshal([]byte(expectedJSON), omCommaStr2)
		assert.Nil(t, err)

		assert.Equal(t, []string{"a,b", "c,d"}, omCommaStr2.Keys())
		assert.Equal(t, []string{"val,1", "val,2"}, omCommaStr2.Values())
	})

	t.Run("ForEach", func(t *testing.T) {
		om := New[string, int]()
		om.Set("first", 1)
		om.Set("second", 2)
		om.Set("third", 3)

		var keys []string
		var values []int

		om.ForEach(func(k string, v int) {
			keys = append(keys, k)
			values = append(values, v)
		})

		assert.Equal(t, []string{"first", "second", "third"}, keys)
		assert.Equal(t, []int{1, 2, 3}, values)
	})

	t.Run("ForEachReverse", func(t *testing.T) {
		om := New[string, int]()
		om.Set("first", 1)
		om.Set("second", 2)
		om.Set("third", 3)

		var keys []string
		var values []int

		om.ForEachReverse(func(k string, v int) {
			keys = append(keys, k)
			values = append(values, v)
		})

		assert.Equal(t, []string{"third", "second", "first"}, keys)
		assert.Equal(t, []int{3, 2, 1}, values)
	})

	t.Run("Iter", func(t *testing.T) {
		om := New[string, int]()
		om.Set("first", 1)
		om.Set("second", 2)
		om.Set("third", 3)

		var keys []string
		var values []int

		om.Iter()(func(k string, v int) bool {
			keys = append(keys, k)
			values = append(values, v)
			return true
		})

		assert.Equal(t, []string{"first", "second", "third"}, keys)
		assert.Equal(t, []int{1, 2, 3}, values)
	})

	t.Run("IterReverse", func(t *testing.T) {
		om := New[string, int]()
		om.Set("first", 1)
		om.Set("second", 2)
		om.Set("third", 3)

		var keys []string
		var values []int

		om.IterReverse()(func(k string, v int) bool {
			keys = append(keys, k)
			values = append(values, v)
			return true
		})

		assert.Equal(t, []string{"third", "second", "first"}, keys)
		assert.Equal(t, []int{3, 2, 1}, values)
	})

	t.Run("ContainsKey", func(t *testing.T) {
		om := New[string, int]()
		om.Set("exists", 100)

		assert.True(t, om.ContainsKey("exists"))
		assert.False(t, om.ContainsKey("missing"))
	})

	t.Run("ContainsValue", func(t *testing.T) {
		om := New[string, int]()
		om.Set("a", 1)
		om.Set("b", 2)
		om.Set("c", 3)

		eq := func(a, b int) bool { return a == b }

		assert.True(t, om.ContainsValue(2, eq))
		assert.False(t, om.ContainsValue(4, eq))
	})

	t.Run("ContainsValueReflect", func(t *testing.T) {
		om := New[string, int]()
		om.Set("a", 1)
		om.Set("b", 2)
		om.Set("c", 3)

		assert.True(t, om.ContainsValueReflect(2))
		assert.False(t, om.ContainsValueReflect(4))
	})

	t.Run("Pop", func(t *testing.T) {
		om := New[string, int]()
		om.Set("key", 42)

		val, ok := om.Pop("key")
		assert.True(t, ok)
		assert.Equal(t, 42, val)

		_, err := om.Get("key")
		assert.NotNil(t, err)
	})

	t.Run("Clone", func(t *testing.T) {
		om := New[string, int]()
		om.Set("a", 1)
		om.Set("b", 2)

		clone := om.Clone()
		assert.Equal(t, om.Keys(), clone.Keys())
		assert.Equal(t, om.Values(), clone.Values())

		om.Set("c", 3)
		assert.NotEqual(t, om.Keys(), clone.Keys())
	})

	t.Run("Merge", func(t *testing.T) {
		om1 := New[string, int]()
		om1.Set("a", 1)
		om1.Set("b", 2)

		om2 := New[string, int]()
		om2.Set("b", 20)
		om2.Set("c", 3)

		om1.Merge(om2)
		assert.Equal(t, []string{"a", "b", "c"}, om1.Keys())

		valB, _ := om1.Get("b")
		valC, _ := om1.Get("c")
		assert.Equal(t, 20, valB)
		assert.Equal(t, 3, valC)
	})

	t.Run("Reverse", func(t *testing.T) {
		om := New[string, int]()
		om.Set("first", 1)
		om.Set("second", 2)
		om.Set("third", 3)

		om.Reverse()
		assert.Equal(t, []string{"third", "second", "first"}, om.Keys())
		assert.Equal(t, []int{3, 2, 1}, om.Values())
	})

	t.Run("SortAsc", func(t *testing.T) {
		om := New[string, int]()
		om.Set("delta", 4)
		om.Set("alpha", 1)
		om.Set("charlie", 3)
		om.Set("bravo", 2)

		err := om.SortAsc()
		assert.Nil(t, err)
		assert.Equal(t, []string{"alpha", "bravo", "charlie", "delta"}, om.Keys())
		assert.Equal(t, []int{1, 2, 3, 4}, om.Values())
	})

	t.Run("SortDesc", func(t *testing.T) {
		om := New[string, int]()
		om.Set("delta", 4)
		om.Set("alpha", 1)
		om.Set("charlie", 3)
		om.Set("bravo", 2)

		err := om.SortDesc()
		assert.Nil(t, err)
		assert.Equal(t, []string{"delta", "charlie", "bravo", "alpha"}, om.Keys())
		assert.Equal(t, []int{4, 3, 2, 1}, om.Values())
	})

	t.Run("IsEmpty (empty)", func(t *testing.T) {
		emptyOM := New[string, int]()
		assert.True(t, emptyOM.IsEmpty())
	})

	t.Run("IsEmpty (non-empty)", func(t *testing.T) {
		nonEmptyOM := New[string, int]()
		nonEmptyOM.Set("a", 1)
		assert.False(t, nonEmptyOM.IsEmpty())
	})

	t.Run("IndexOf", func(t *testing.T) {
		om := New[string, int]()
		om.Set("a", 1)
		om.Set("b", 2)

		assert.Equal(t, 0, om.IndexOf("a"))
		assert.Equal(t, 1, om.IndexOf("b"))
		assert.Equal(t, -1, om.IndexOf("c"))
	})
}
