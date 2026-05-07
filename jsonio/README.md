In **Go 1.23+**, using `iter.Seq[error]` is generally considered an **antipattern** for standard error handling. While technically functional, it breaks the idiomatic flow of the language and introduces ambiguity in how a consumer should interact with the sequence.

### Why it is discouraged

*   **Ambiguous Termination:** In Go, an error usually signals that a process has stopped. With `iter.Seq[error]`, it is unclear if yielding an error means "I found a problem, but keep going" or "I am finished because of this error."
*   **The "Double Check" Requirement:** If you yield an error, the caller must check the error *inside* the loop. However, if the iterator encounters a fatal error and stops yielding entirely, the caller might assume the loop finished successfully.
*   **API Consistency:** Standard library iterators (like those in `slices` or `maps`) do not yield errors. Introducing them via the sequence type makes your API inconsistent with the rest of the ecosystem.

---

### The Idiomatic Pattern: The "Sentinel" Approach
The recommended way to handle errors in iterators is to yield the data (or a pair) and provide a separate method or a closure-bound variable to check for an error **after** the loop completes. This mirrors the `bufio.Scanner` pattern.

#### Example: Recommended Pattern
```go
package main

import (
	"fmt"
	"iter"
)

// Result wraps the data and the error
type Result struct {
	Value string
	Err   error
}

// DataStream returns an iterator. 
// If a fatal error occurs, it stops yielding.
func DataStream(items []string) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		for _, item := range items {
			if item == "trigger-error" {
				yield(Result{Err: fmt.Errorf("bad item found")})
				return // Terminate the iterator
			}
			if !yield(Result{Value: item}) {
				return
			}
		}
	}
}

func main() {
	items := []string{"apple", "trigger-error", "banana"}

	for res := range DataStream(items) {
		if res.Err != nil {
			fmt.Printf("Stopped due to error: %v\n", res.Err)
			break
		}
		fmt.Println("Processing:", res.Value)
	}
}
```

---

### When `iter.Seq2[V, error]` makes sense
While `iter.Seq[error]` is almost always wrong, `iter.Seq2[V, error]` is occasionally acceptable for **row-based** operations (like database cursor results) where every single iteration could potentially fail independently.

| Approach | Use Case |
| :--- | :--- |
| **`iter.Seq[V]` + `err` method** | High-performance streams or scanners (Most common). |
| **`iter.Seq[Result{V, error}]`** | When you want to force the user to acknowledge the error at every step. |
| **`iter.Seq[error]`** | **Antipattern.** Use a simple `error` return or a channel instead. |

**Summary:** Avoid `iter.Seq[error]`. If your iterator can fail, yield a struct that contains the error or a `Seq2` where the second value is the error, and ensure the iterator terminates immediately after a fatal error is yielded.