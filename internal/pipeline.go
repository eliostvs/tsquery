package tsquery

import (
	"sync"
)

func fanIn[T any](sources ...<-chan T) <-chan T {
	dest := make(chan T)

	var wg sync.WaitGroup
	wg.Add(len(sources))

	for _, ch := range sources {
		go func(c <-chan T) {
			defer wg.Done()

			for n := range c {
				dest <- n
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		defer close(dest)
	}()

	return dest
}

func fanOut[T any](source <-chan T, n int) []<-chan T {
	dest := make([]<-chan T, 0)

	for range n {
		ch := make(chan T)
		dest = append(dest, ch)

		go func() {
			defer close(ch)

			for i := range source {
				ch <- i
			}
		}()
	}

	return dest
}
