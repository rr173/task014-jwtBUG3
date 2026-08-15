package jwt

import (
	"sync"
	"testing"
)

// TestSignRaceOnSharedExtra verifies that Sign does not panic or race when
// called concurrently with a Claims whose Extra map is being mutated externally.
// The fix should copy the Extra map before iterating.
func TestSignRaceOnSharedExtra(t *testing.T) {
	secret := []byte("secret")
	extra := map[string]any{"role": "admin", "tier": float64(1)}
	claims := Claims{Subject: "alice", Extra: extra}

	var wg sync.WaitGroup

	// Writer goroutine mutates Extra concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			extra["counter"] = i
		}
	}()

	// Reader goroutines call Sign which iterates Extra via MarshalJSON
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := Sign(claims, secret)
				if err != nil {
					t.Errorf("sign: %v", err)
				}
			}
		}()
	}
	wg.Wait()
}
