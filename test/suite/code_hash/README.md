# Code Hash Test Suite

Test cases that check if we are computing the code hash properly.

A good way to manually check for what the hash should be is to use a program
[like this](https://go.dev/play/p/rXkdYzMXgvw):

```go
package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	tokens := []string{"function", "main", "(", ")", ":", "void", "end"}
	hasher := sha256.New()

	for _, token := range tokens {
		hasher.Write([]byte(token))
		hasher.Write([]byte{0})
	}
	hash := hasher.Sum(nil)
	fmt.Printf("%x", hash)
}
```
