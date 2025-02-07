package orderedmap

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

func TestOrderedMap(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		om := NewOrderedMap[string, int]()

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

		// Test Update
		om.Set("b", 5)
		val, _ = om.Get("b")
		assert.Equal(t, 5, val)
	})

	t.Run("delete", func(t *testing.T) {
		om := NewOrderedMap[string, int]()
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
		om := NewOrderedMap[string, int]()
		om.Set("a", 1)
		om.Set("c", 3)
		jsonData, err := json.Marshal(om)
		assert.Nil(t, err)
		expectedJSON := `{"a":1,"c":3}`
		assert.JSONEq(t, expectedJSON, string(jsonData))
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		jsonInput := `{"x":10,"y":20,"z":30}`
		om2 := NewOrderedMap[string, int]()
		err := json.Unmarshal([]byte(jsonInput), om2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"x", "y", "z"}, om2.Keys())
		assert.Equal(t, []int{10, 20, 30}, om2.Values())
	})

	t.Run("empty map", func(t *testing.T) {
		emptyOM := NewOrderedMap[string, int]()
		emptyJSON, err := json.Marshal(emptyOM)
		assert.Nil(t, err)
		assert.Equal(t, "{}", string(emptyJSON))
	})

	t.Run("single-element", func(t *testing.T) {
		om := NewOrderedMap[string, int]()
		om.Set("single", 42)
		oneJSON, err := json.Marshal(om)
		assert.Nil(t, err)
		assert.Contains(t, string(oneJSON), `"single":42`)
	})

	t.Run("nested maps", func(t *testing.T) {
		inner := NewOrderedMap[string, int]()
		inner.Set("inner1", 100)
		inner.Set("inner2", 200)

		outer := NewOrderedMap[string, *orderedMap[string, int]]()
		outer.Set("outer", inner)

		assert.Equal(t, []string{"outer"}, outer.Keys())
		nestedMap := outer.GetOrEmpty("outer")
		assert.Equal(t, []string{"inner1", "inner2"}, nestedMap.Keys())
		assert.Equal(t, []int{100, 200}, nestedMap.Values())

		jsonData, err := json.Marshal(outer)
		assert.Nil(t, err)
		expectedJSON := `{"outer":{"inner1":100,"inner2":200}}`
		assert.JSONEq(t, expectedJSON, string(jsonData))

		outer2 := NewOrderedMap[string, *orderedMap[string, int]]()
		err = json.Unmarshal([]byte(expectedJSON), outer2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"outer"}, outer2.Keys())
		nestedMap2 := outer2.GetOrEmpty("outer")
		assert.Equal(t, []string{"inner1", "inner2"}, nestedMap2.Keys())
		assert.Equal(t, []int{100, 200}, nestedMap2.Values())
	})

	t.Run("keys with commas", func(t *testing.T) {
		omComma := NewOrderedMap[string, int]()
		omComma.Set("a,b", 10)
		omComma.Set("c,d", 20)

		assert.Equal(t, []string{"a,b", "c,d"}, omComma.Keys())
		assert.Equal(t, []int{10, 20}, omComma.Values())

		jsonData, err := json.Marshal(omComma)
		assert.Nil(t, err)
		expectedJSON := `{"a,b":10,"c,d":20}`
		assert.JSONEq(t, expectedJSON, string(jsonData))

		omComma2 := NewOrderedMap[string, int]()
		err = json.Unmarshal([]byte(expectedJSON), omComma2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"a,b", "c,d"}, omComma2.Keys())
		assert.Equal(t, []int{10, 20}, omComma2.Values())
	})

	t.Run("keys and values with commas", func(t *testing.T) {
		omCommaStr := NewOrderedMap[string, string]()
		omCommaStr.Set("a,b", "val,1")
		omCommaStr.Set("c,d", "val,2")

		assert.Equal(t, []string{"a,b", "c,d"}, omCommaStr.Keys())
		assert.Equal(t, []string{"val,1", "val,2"}, omCommaStr.Values())

		jsonData, err := json.Marshal(omCommaStr)
		assert.Nil(t, err)
		expectedJSON := `{"a,b":"val,1","c,d":"val,2"}`
		assert.JSONEq(t, expectedJSON, string(jsonData))

		omCommaStr2 := NewOrderedMap[string, string]()
		err = json.Unmarshal([]byte(expectedJSON), omCommaStr2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"a,b", "c,d"}, omCommaStr2.Keys())
		assert.Equal(t, []string{"val,1", "val,2"}, omCommaStr2.Values())
	})
}
