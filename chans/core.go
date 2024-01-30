package chans

import (
	"sync"
	"sync/atomic"
)

func MapAndFilter[A, B any](in <-chan A, n int, f func(A) (B, bool)) <-chan B {
	if in == nil {
		return nil
	}

	out := make(chan B)

	loop(in, out, n, func(x A) {
		y, keep := f(x)
		if keep {
			out <- y
		}
	})

	return out
}

func OrderedMapAndFilter[A, B any](in <-chan A, n int, f func(A) (B, bool)) <-chan B {
	if in == nil {
		return nil
	}

	out := make(chan B)
	orderedLoop(in, out, n, func(a A, canWrite <-chan struct{}) {
		y, keep := f(a)
		<-canWrite
		if keep {
			out <- y
		}
	})

	return out
}

func Map[A, B any](in <-chan A, n int, f func(A) B) <-chan B {
	return MapAndFilter(in, n, func(a A) (B, bool) {
		return f(a), true
	})
}

func OrderedMap[A, B any](in <-chan A, n int, f func(A) B) <-chan B {
	return OrderedMapAndFilter(in, n, func(a A) (B, bool) {
		return f(a), true
	})
}

func Filter[A any](in <-chan A, n int, f func(A) bool) <-chan A {
	return MapAndFilter(in, n, func(a A) (A, bool) {
		return a, f(a)
	})
}

func OrderedFilter[A any](in <-chan A, n int, f func(A) bool) <-chan A {
	return OrderedMapAndFilter(in, n, func(a A) (A, bool) {
		return a, f(a)
	})
}

func FlatMap[A, B any](in <-chan A, n int, f func(A) <-chan B) <-chan B {
	if in == nil {
		return nil
	}

	out := make(chan B)

	loop(in, out, n, func(a A) {
		for b := range f(a) {
			out <- b
		}
	})

	return out
}

func OrderedFlatMap[A, B any](in <-chan A, n int, f func(A) <-chan B) <-chan B {
	if in == nil {
		return nil
	}

	out := make(chan B)
	orderedLoop(in, out, n, func(a A, canWrite <-chan struct{}) {
		bb := f(a)
		<-canWrite
		for b := range bb {
			out <- b
		}
	})

	return out
}

// blocking
// todo: explain that if false has been returned for item[i] that it's guranteed that function would have been called for all previous items
func ForEach[A any](in <-chan A, n int, f func(A) bool) {
	// In case of early exit some unconsumed items will be left in the 'in' channel.
	// To avoid leaks we need to consume everything until channel is closed.
	// On the other hand caller can close in, only after we return.
	// So drain also must happen only after we return. The correct order is:
	// early exit -> caller closes 'in' -> drain 'in'
	// That's why we're using non-blocking drain here.
	defer DrainNB(in)

	if n == 1 {
		for a := range in {
			if !f(a) {
				break
			}
		}

		return
	}

	var wg sync.WaitGroup
	earlyExit := int64(0)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for a := range in {
				ok := f(a)
				if !ok {
					atomic.AddInt64(&earlyExit, 1)
					break
				} else if atomic.LoadInt64(&earlyExit) > 0 {
					break
				}
			}
		}()
	}

	wg.Wait()
}
