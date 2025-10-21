package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// fetchData simulates calling an external API or database.
func fetchData(ctx context.Context, id int) (string, error) {
	// Simulate random processing time between 1–4 seconds
	delay := time.Duration(rand.Intn(4)+1) * time.Second
	fmt.Printf("Worker %d: starting (will take %v)...\n", id, delay)

	select {
	case <-time.After(delay):
		// Finished before timeout
		fmt.Printf("Worker %d: done!\n", id)
		return fmt.Sprintf("data-%d", id), nil
	case <-ctx.Done():
		// Context cancelled or timeout
		fmt.Printf("Worker %d: stopped (%v)\n", id, ctx.Err())
		return "", ctx.Err()
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create a parent context with timeout (3s total)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // Always cancel to free resources

	results := make(chan string)
	errors := make(chan error)

	// Launch multiple workers concurrently
	for i := 1; i <= 3; i++ {
		go func(id int) {
			data, err := fetchData(ctx, id)
			if err != nil {
				errors <- err
				return
			}
			results <- data
		}(i)
	}

	// Collect the first successful result or stop if context is done
	select {
	case res := <-results:
		fmt.Println("✅ Got result:", res)
	case err := <-errors:
		fmt.Println("❌ Error:", err)
	case <-ctx.Done():
		fmt.Println("⏰ Timeout reached:", ctx.Err())
	}

	fmt.Println("Main finished.")
}
// func WithDeadline(context, time.Duration)(context, context.CancelFunc)

/*
accepts specific time at which the context will cancelled and the done channel will be closed 
func withtimeout(context, time.Duration)(Context, CncelFunc)
Accepts a durition after which the context will cancelled and the done channel will be closed
func WithCancle (Context )(Context, CancelFunc)

unlike the previous function , withcancel acceptes noting and only return fucntion that can called to explicity cnacel the context.
*/

