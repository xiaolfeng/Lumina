package logic

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

func TestWithLpwFileLockMutualExclusion(t *testing.T) {
	t.Parallel()

	// Q-04 回归：同 key 并发临界区必须完全串行（计数无丢失、无数据竞争）
	var counter int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			withLpwFileLock(context.Background(), 101, "a.lpw", func(_ context.Context) {
				// 非原子自增：若锁未生效，并发下结果必然 < 100 或触发 -race
				cur := atomic.LoadInt64(&counter)
				time.Sleep(time.Microsecond) // 放大竞争窗口
				atomic.StoreInt64(&counter, cur+1)
			})
		}()
	}
	wg.Wait()
	if counter != 100 {
		t.Errorf("expected counter=100 under mutual exclusion, got %d", counter)
	}
}

func TestWithLpwFileLockReentrant(t *testing.T) {
	t.Parallel()

	// Q-04 回归：持锁上下文内再进同 key（模拟 withDocument → SaveFile → UploadFile）
	// 必须直接执行而不自死锁；不同 key 仍需真正加锁。
	done := make(chan struct{})
	withLpwFileLock(context.Background(), 202, "b.lpw", func(ctx context.Context) {
		withLpwFileLock(ctx, 202, "b.lpw", func(ctx context.Context) {
			// 第三层嵌套重入也应直接执行
			withLpwFileLock(ctx, 202, "b.lpw", func(_ context.Context) {
				close(done)
			})
		})
	})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("同 key 重入发生自死锁")
	}
}

func TestWithLpwFileLockDifferentKeysIndependent(t *testing.T) {
	t.Parallel()

	// 不同 key 互不阻塞：b1 持锁期间 b2 应能进入
	entered := make(chan struct{})
	released := make(chan struct{})
	go func() {
		withLpwFileLock(context.Background(), 303, "b1.lpw", func(_ context.Context) {
			close(entered)
			<-released
		})
	}()
	<-entered

	done := make(chan struct{})
	go func() {
		withLpwFileLock(context.Background(), 303, "b2.lpw", func(_ context.Context) {
			close(done)
		})
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		close(released)
		t.Fatal("不同 key 被错误地相互阻塞")
	}
	close(released)
}

func TestLpwCrossChannelSerialization(t *testing.T) {
	t.Parallel()

	// Q-04 核心场景：模拟「lpw 读改验写」与「整体覆写」交错执行，
	// 共享锁必须保证任一通道的完整执行期间另一通道不得插入。
	var sequence []string
	var mu sync.Mutex // 仅保护 slice 追加的可见性顺序断言

	runPipeline := func(name string) {
		withLpwFileLock(context.Background(), xSnowflake.SnowflakeID(404), "cross.lpw", func(ctx context.Context) {
			mu.Lock()
			sequence = append(sequence, name+"-begin")
			mu.Unlock()
			time.Sleep(2 * time.Millisecond) // 模拟读-改-验耗时
			// 模拟 withDocument 内部落库触发 UploadFile（同 key 重入，发生在持锁期内）
			if name == "lpw" {
				withLpwFileLock(ctx, xSnowflake.SnowflakeID(404), "cross.lpw", func(_ context.Context) {
					mu.Lock()
					sequence = append(sequence, "lpw-save(reentrant)")
					mu.Unlock()
				})
			}
			mu.Lock()
			sequence = append(sequence, name+"-end")
			mu.Unlock()
		})
	}

	var wg sync.WaitGroup
	for _, name := range []string{"lpw", "upload", "lpw", "upload"} {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			runPipeline(n)
		}(name)
	}
	wg.Wait()

	// 不变量：任何通道的 begin 与 end 之间不得插入另一通道的操作；
	// 重入保存必须发生在其所属 lpw 临界区内。
	type span struct{ owner string; begin, end int }
	var spans []span
	reentrantOwners := make(map[int]struct{})
	var current map[string]*span = map[string]*span{}
	for i, op := range sequence {
		switch {
		case op == "lpw-begin":
			current["lpw"] = &span{owner: "lpw", begin: i}
		case op == "upload-begin":
			current["upload"] = &span{owner: "upload", begin: i}
		case op == "lpw-end":
			if s := current["lpw"]; s != nil {
				s.end = i
				spans = append(spans, *s)
				delete(current, "lpw")
			}
		case op == "upload-end":
			if s := current["upload"]; s != nil {
				s.end = i
				spans = append(spans, *s)
				delete(current, "upload")
			}
		case op == "lpw-save(reentrant)":
			reentrantOwners[i] = struct{}{}
		}
	}

	for a := 0; a < len(spans); a++ {
		for b := a + 1; b < len(spans); b++ {
			x, y := spans[a], spans[b]
			if x.begin < y.begin && y.begin < x.end {
				t.Fatalf("%s 侵入 %s 临界区: %v", y.owner, x.owner, sequence)
			}
			if y.begin < x.begin && x.begin < y.end {
				t.Fatalf("%s 侵入 %s 临界区: %v", x.owner, y.owner, sequence)
			}
		}
	}

	// 重入保存必须落在某个 lpw 临界区内部
	lpwSpans := 0
	for _, s := range spans {
		if s.owner != "lpw" {
			continue
		}
		lpwSpans++
	}
	if len(reentrantOwners) != lpwSpans {
		t.Fatalf("重入保存次数(%d)应等于 lpw 临界区数(%d): %v", len(reentrantOwners), lpwSpans, sequence)
	}
	for i := range reentrantOwners {
		inside := false
		for _, s := range spans {
			if s.owner == "lpw" && i > s.begin && i < s.end {
				inside = true
				break
			}
		}
		if !inside {
			t.Fatalf("重入保存(i=%d)未发生在 lpw 临界区内: %v", i, sequence)
		}
	}
}
