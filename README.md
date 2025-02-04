# orderedmap  

A Golang data structure that preserves insertion order while behaving like a standard map, similar to LinkedHashMap in Java.

- Maintains key insertion order
- Supports JSON serialization/deserialization

# Usage

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/emretanriverdi/orderedmap"
)

func main() {
	// Create a new OrderedMap
	om := orderedmap.NewOrderedMap[string, int]()

	// Insert key-value pairs
	om.Set("a", 1)
	om.Set("b", 2)
	om.Set("c", 3)

	// Retrieve values
	val := om.Get("b") // Returns 2
	fmt.Println(val)

	// Get all keys in insertion order
	keys := om.Keys()
	fmt.Println(keys) // Output: [a b c]

	// Delete a key
	om.Delete("b")

	// Serialize to JSON
	jsonData, _ := json.Marshal(om)
	fmt.Println(string(jsonData)) // Output: {"a":1,"c":3}

	// Deserialize from JSON
	json.Unmarshal([]byte(`{"x": 100, "y": 200}`), &om)

	// Print updated map
	fmt.Println(om.Keys()) // Output: [x y]
}
```

# Features

* Order Preservation: Unlike Go’s built-in maps, orderedmap keeps track of the order in which keys were inserted.
* JSON Support: Can be marshaled/unmarshaled while preserving order.

# Caveats

* Only supports JSON serialization when keys are of type string, as per JSON specifications.
* Not optimized for concurrent access, use sync.Mutex if needed.

# Tests

```
go test
```

# Disclaimer  

This project was created for fun and as a simple exercise. While it works as intended, there are other alternatives available.  

If you plan to use it in a production environment, please do so at your own risk.
