# orderedmap  

A Golang data structure that preserves insertion order while behaving like a standard map, similar to LinkedHashMap in Java.

- Maintains key insertion order
- Supports JSON serialization/deserialization

# Usage

```go
package main

import (
	"encoding/json" // or any other lib of your choice
	"fmt"
	"github.com/emretanriverdi/orderedmap"
)

func main() {
	// Create a new OrderedMap
	om := orderedmap.New[string, int]()

	// Insert key-value pairs
	om.Set("a", 1)
	om.Set("b", 2)
	om.Set("c", 3)

	// Retrieve values
	val := om.GetOrDefault("b") // Returns 2
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

# Caveats

* Not optimized for concurrent access, use sync.Mutex if needed.

# Tests

```
go test
```

# Disclaimer  

This project was created for fun and as a simple exercise.

If you plan to use it in production environment, please do so at your own risk.
