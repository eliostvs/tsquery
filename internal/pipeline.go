package tsquery

import (
	"context"
	"sync"
)

func fanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	output := make(chan T)

	go func() {
		defer close(output)

		var wg sync.WaitGroup
		wg.Add(len(inputs))

		for _, input := range inputs {
			go func(in <-chan T) {
				defer wg.Done()

				for {
					select {
					case item, ok := <-in:
						if !ok {
							return
						}

						select {
						case output <- item:
							return
						case <-ctx.Done():
							return
						}

					case <-ctx.Done():
						return
					}
				}
			}(input)
		}

		wg.Wait()
	}()

	return output
}

func fanOut[T any](ctx context.Context, input <-chan T, numWorkers int) []<-chan T {
	outputs := make([]<-chan T, numWorkers)

	for i := range numWorkers {
		output := make(chan T)
		outputs[i] = output

		go func(out chan<- T) {
			defer close(out)

			for {
				select {
				case item, ok := <-input:
					if !ok {
						return
					}

					select {
					case out <- item:
						return
					case <-ctx.Done():
						return
					}

				case <-ctx.Done():
					return
				}
			}
		}(output)
	}

	return outputs
}
