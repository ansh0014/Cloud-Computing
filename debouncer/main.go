package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Circuit is a function type that performs some external call or computation.
// It returns a result string and an error.
type Circuit func(context.Context) (string, error)

// DebounceLast wraps a Circuit function with debounce logic.
// It ensures the function executes only once after a period of inactivity (d).
func DebounceLast(circuit Circuit, d time.Duration) Circuit {
	var (
		mu        sync.Mutex
		ticker    *time.Ticker
		threshold time.Time
		result    string
		err       error
		once      sync.Once
	)

	return func(ctx context.Context) (string, error) {
		mu.Lock()
		// Every new call resets the debounce threshold
		threshold = time.Now().Add(d)
		mu.Unlock()

		// Ensure ticker goroutine starts only once
		once.Do(func() {
			ticker = time.NewTicker(100 * time.Millisecond) // check every 100ms

			go func() {
				defer func() {
					mu.Lock()
					ticker.Stop()
					once = sync.Once{} // reset Once for future debounce cycles
					mu.Unlock()
				}()

				for {
					select {
					case <-ticker.C:
						mu.Lock()
						// If the current time passes threshold => execute circuit
						if time.Now().After(threshold) {
							fmt.Println(" Quiet period over — running debounced circuit...")
							result, err = circuit(ctx)
							mu.Unlock()
							return
						}
						mu.Unlock()

					case <-ctx.Done():
						mu.Lock()
						result, err = "", ctx.Err()
						mu.Unlock()
						return
					}
				}
			}()
		})

		// Return current known result (not updated until circuit runs)
		mu.Lock()
		defer mu.Unlock()
		return result, err
	}
}


// Example Circuit: simulates an expensive or external API call.

func apiCall(ctx context.Context) (string, error) {
	fmt.Println(" Executing API call at:", time.Now().Format("15:04:05"))
	return " Success", nil
}



// Main function — Demonstrates the DebounceLast behavior

func main() {
	ctx := context.Background()
	debouncedAPI := DebounceLast(apiCall, 2*time.Second) // wait 2s of inactivity

	fmt.Println("📘 Simulating multiple rapid API trigger calls...")

	for i := 1; i <= 5; i++ {
		fmt.Printf("Trigger #%d at %v\n", i, time.Now().Format("15:04:05"))
		debouncedAPI(ctx)
		time.Sleep(500 * time.Millisecond) // user keeps triggering every 0.5s
	}
	fmt.Println(" Waiting 3s (quiet period)...")
	time.Sleep(3 * time.Second)

	fmt.Println(" Program ended.")
}
