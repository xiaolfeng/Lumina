package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

func TestBiometricChallengeUsesDynamicTTLAndCanOnlyBeConsumedOnce(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := &BiometricCredentialCache{Base: &Base{RDB: rdb}}
	ctx := context.Background()
	ttl := 330 * time.Second

	if xErr := cache.SetChallenge(ctx, "reg", "session", []byte("session-data"), ttl); xErr != nil {
		t.Fatalf("SetChallenge() error = %v", xErr)
	}
	key := bConst.CacheBiometricChallengeRegister.Get("session").String()
	if got := mr.TTL(key); got != ttl {
		t.Fatalf("challenge TTL = %s, want %s", got, ttl)
	}

	type result struct {
		data []byte
		ok   bool
		err  error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, ok, xErr := cache.ConsumeChallenge(ctx, "reg", "session")
			var err error
			if xErr != nil {
				err = xErr
			}
			results <- result{data: data, ok: ok, err: err}
		}()
	}
	wg.Wait()
	close(results)

	hits := 0
	for got := range results {
		if got.err != nil {
			t.Fatalf("ConsumeChallenge() error = %v", got.err)
		}
		if got.ok {
			hits++
			if string(got.data) != "session-data" {
				t.Fatalf("challenge data = %q", got.data)
			}
		}
	}
	if hits != 1 {
		t.Fatalf("challenge consumed %d times, want exactly once", hits)
	}
}
