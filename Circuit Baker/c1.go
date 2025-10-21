package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)



type CircuitBreaker struct {
	mu               sync.Mutex
	state            string // "closed", "open", "half-open"
	failureCount     int
	failureThreshold int
	resetTimeout     time.Duration
	lastFailureTime  time.Time
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            "closed",
		failureThreshold: threshold,
		resetTimeout:     timeout,
	}
}

// Call wraps external API call
func (cb *CircuitBreaker) Call(ctx context.Context, fn func(context.Context) error) error {
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()

	switch state {
	case "open":
		// Check if we can move to half-open (cooldown passed)
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			fmt.Println("🔄 Moving to HALF-OPEN state")
			cb.setState("half-open")
		} else {
			return errors.New(" circuit breaker OPEN — failing fast")
		}
	}

	// Execute the function
	err := fn(ctx)
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	fmt.Printf(" Failure #%d\n", cb.failureCount)

	if cb.failureCount >= cb.failureThreshold {
		cb.setState("open")
		cb.lastFailureTime = time.Now()
		fmt.Println(" Circuit OPENED! All requests will now fail fast.")
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	if cb.state == "half-open" {
		fmt.Println("✅ Service recovered! Circuit CLOSED again.")
	}
	cb.setState("closed")
}

func (cb *CircuitBreaker) setState(s string) {
	cb.state = s
}


func unstableExternalAPI(ctx context.Context) error {
	delay := time.Duration(rand.Intn(1000)+300) * time.Millisecond

	select {
	case <-time.After(delay):
		if rand.Float32() < 0.5 {
			return errors.New("external API failure")
		}
		fmt.Println("🌍 External API succeeded in", delay)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}


func paymentHandler(cb *CircuitBreaker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err := cb.Call(ctx, unstableExternalAPI)
		if err != nil {
			// Fallback response if circuit is open or API failed
			log.Println("Fallback:", err)
			http.Error(w, " Payment Service temporarily unavailable (fallback triggered)", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintln(w, " Payment processed successfully!")
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	cb := NewCircuitBreaker(3, 5*time.Second) // 3 failures -> open circuit for 5s

	http.HandleFunc("/payment", paymentHandler(cb))

	fmt.Println(" Server running on http://localhost:8080")
	fmt.Println(" Hit /payment multiple times to observe circuit behavior.")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
