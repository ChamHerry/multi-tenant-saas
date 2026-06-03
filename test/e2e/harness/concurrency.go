//go:build e2e

package harness

import "sync"

func RunConcurrent(n int, fn func(i int)) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			fn(i)
		}()
	}
	close(start)
	wg.Wait()
}
