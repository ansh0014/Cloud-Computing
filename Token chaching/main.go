package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"sync"
	"time"
)

// Token Cachet
type TokenCache struct {
	mu     sync.RWMutex
	tokens map[string]TokenData
}
type TokenData struct {
	UserId    string
	ExpiresAt time.Time
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		tokens: make(map[string]TokenData),
	}
}

// Read (Many readers allowed )
func (c *TokenCache) set(token, userID string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[token] = TokenData{
		UserId:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// cleanup Expired Tokens (background job)
func (c *TokenCache) cleanupExpired() {
	for {
		time.Sleep(2 * time.Second)
		c.mu.Lock()
		for k, v := range c.tokens {
			if time.Now().After(v.ExpiresAt) {
				fmt.Println("cleaning expired token:", k)
				delete(c.tokens, k)
			}

		}
		c.mu.Unlock()
	}
}

// Token Generation
// generateToken creates a random 32 byte hex string (like JWT)
func generateToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
func (c *TokenCache) Get(token string) (string, bool) {
	c.mu.RLock()
	td, ok := c.tokens[token]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(td.ExpiresAt) {
		return "", false
	}
	return td.UserId, true
}

func (c *TokenCache) Tokens() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.tokens))
	for k := range c.tokens {
		keys = append(keys, k)
	}
	return keys
}

// main function
func main() {
	cache := NewTokenCache()
	//  start background cleanup for expired tokens
	go cache.cleanupExpired()
	// writer goroutine: issues new tokens for users
	go func() {
		for i := 1; i <= 3; i++ {
			token := generateToken()
			cache.set(token, fmt.Sprintf("user-%d", i), 5*time.Second)
			fmt.Printf("Issued token for user-%d: %s\n", i, token)
			time.Sleep(3 * time.Second)

		}
	}()
	// reader goroutines: simulate clients validating tokens
	for j := 0; j < 4; j++ {
		go func(id int) {
			for {
				time.Sleep(1 * time.Second)
				tokens := cache.Tokens()
				for _, token := range tokens {
					user, ok := cache.Get(token)
					if ok {
						fmt.Printf("reader -%d validation token for %s\n", id, user)
					} else {
						fmt.Printf("reader-%d: invalid or expired token\n", id)
					}
				}
			}
		}(j)
	}
	// run for 20 second
	time.Sleep(20 * time.Second)
	fmt.Println("simulated end")
}
