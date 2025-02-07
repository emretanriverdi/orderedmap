package orderedmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedMap(t *testing.T) {
	om := OrderedMap[string, int]{}

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

	// Test Delete (O(1) swap-based)
	om.Delete("b")
	_, err = om.Get("b")
	assert.NotNil(t, err)
	assert.ElementsMatch(t, []string{"a", "c"}, om.Keys())
	assert.ElementsMatch(t, []int{1, 3}, om.Values())

	// Test JSON Marshaling
	expectedJSON := `{"a":1,"c":3}`
	jsonData, err := json.Marshal(om)
	assert.Nil(t, err)
	assert.JSONEq(t, expectedJSON, string(jsonData))

	// Test JSON Unmarshaling
	jsonInput := `{"x":10,"y":20,"z":30}`
	om2 := OrderedMap[string, int]{}
	err = json.Unmarshal([]byte(jsonInput), om2)
	assert.Nil(t, err)
	assert.Equal(t, []string{"x", "y", "z"}, om2.Keys())
	assert.Equal(t, []int{10, 20, 30}, om2.Values())

	// Empty OrderedMap
	emptyOM := OrderedMap[string, int]{}
	emptyJSON, err := json.Marshal(emptyOM)
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(emptyJSON))

	// Single-element OrderedMap
	om.Set("single", 42)
	oneJSON, err := json.Marshal(om)
	assert.Nil(t, err)
	assert.Contains(t, string(oneJSON), `"single":42`)
}
