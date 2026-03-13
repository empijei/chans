package chans

import (
	"sync/atomic"
	"time"
)

// TrySend tries to send the value, blocking.
func TrySend[T any](done Done, sink chan<- T, v T) (ok bool) {
	select {
	case sink <- v:
	case <-done:
		return false
	}
	return true
}

// Debounce emits the last value from src when the time window has passed without emissions.
//
// This introduces an emission delay and buffering of the last element.
//
// Debounce spawns a goroutine.
func Debounce[T any](done Done, src <-chan T, window time.Duration) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		var (
			last T
			t    <-chan time.Time
		)
		for {
			select {
			case <-done:
				return
			case v, ok := <-src:
				if !ok {
					if t != nil {
						TrySend(done, ret, last)
					}
					return
				}
				last = v
				t = time.After(window)
			case <-t:
				TrySend(done, ret, last)
				t = nil
			}
		}
	}()
	return ret
}

// Throttle emits values generated from src at most once every window.
//
// All values after the first of the window are discarded.
//
// Throttle spawns a goroutine.
func Throttle[T any](done Done, src <-chan T, window time.Duration) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		t := time.NewTicker(window)
		var sent bool
		for {
			select {
			case <-done:
				return
			case v, ok := <-src:
				if !ok {
					return
				}
				if sent {
					continue
				}
				sent = TrySend(done, ret, v)
			case <-t.C:
				sent = false
			}
		}
	}()
	return ret
}

// Window emits values generated from src at most once every window.
//
// Values generated after the first of every window are discarded except for the last,
// which is deferred to the beginning of the next window.
//
// This introduces an emission delay and buffering of one element.
//
// Window spawns a goroutine.
func Window[T any](done Done, src <-chan T, window time.Duration) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		t := time.NewTicker(window)
		var (
			usedWindow bool
			lastValid  bool
			last       T
		)
		for {
			select {
			case <-done:
				return
			case v, ok := <-src:
				if !ok {
					return
				}
				if usedWindow {
					last = v
					lastValid = true
					continue
				}
				TrySend(done, ret, v)
				usedWindow = true
			case <-t.C:
				if lastValid {
					TrySend(done, ret, last)
					lastValid = false
					continue
				}

				usedWindow = false
			}
		}
	}()
	return ret
}

// Concat concatenates a series of channels.
func Concat[T any](done Done, srcs ...<-chan T) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
	srcsLoop:
		for _, src := range srcs {
			for {
				select {
				case <-done:
					return
				case v, ok := <-src:
					if !ok {
						continue srcsLoop
					}
					if !TrySend(done, ret, v) {
						return
					}
				}
			}
		}
	}()
	return ret
}

// Map transforms every emission of the source into a projection of that emission.
//
// Map spawns a new goroutine.
func Map[I, O any](done Done, src <-chan I, projection func(I) O) <-chan O {
	ret, _ := MapTry(done, src, func(i I) (o O, err error) {
		return projection(i), nil
	})
	return ret
}

// MapTry is like [Map] but with a projection function that may fail.
func MapTry[I, O any](done Done, src <-chan I, projection func(I) (O, error)) (<-chan O, <-chan error) {
	ret := make(chan O)
	errs := make(chan error)
	go func() {
		defer close(ret)
		defer close(errs)
		for {
			select {
			case i, ok := <-src:
				if !ok {
					return
				}
				o, err := projection(i)
				if err != nil {
					if !TrySend(done, errs, err) {
						return
					}
					break
				}
				if !TrySend(done, ret, o) {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return ret, errs
}

// MapParallel is like [Map] but calls projection with the given level of parallelism.
// Output order is not guaranteed.
//
// MapParallel spawns a new goroutine per parallelism level specified.
func MapParallel[I, O any](done Done, src <-chan I, parallelism int, projection func(I) O) <-chan O {
	ret, _ := MapParallelTry(done, src, parallelism, func(i I) (O, error) {
		return projection(i), nil
	})
	return ret
}

// MapParallelTry is like [MapParallel] and [MapTry], together.
func MapParallelTry[I, O any](done Done, src <-chan I, parallelism int, projection func(I) (O, error)) (<-chan O, <-chan error) {
	if parallelism <= 0 {
		parallelism = 1
	}

	ret := make(chan O)
	errs := make(chan error)
	var liveRoutines atomic.Int64
	liveRoutines.Add(int64(parallelism))
	for range parallelism {
		go func() {
			defer func() {
				if liveRoutines.Add(-1) > 0 {
					return
				}
				close(ret)
				close(errs)
			}()

			for {
				select {
				case i, ok := <-src:
					if !ok {
						return
					}
					o, err := projection(i)
					if err != nil {
						if !TrySend(done, errs, err) {
							return
						}
						break
					}
					if !TrySend(done, ret, o) {
						return
					}
				case <-done:
					return
				}
			}
		}()
	}
	return ret, errs
}

// Filter emits only values that filter returns true for.
//
// Filter spawns a new goroutine.
func Filter[T any](done Done, src <-chan T, filter func(T) bool) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		for {
			select {
			case i, ok := <-src:
				if !ok {
					return
				}
				if !filter(i) {
					continue
				}
				if !TrySend(done, ret, i) {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return ret
}

// Merge concurrently merges all the emissions from all the provided sources.
//
// It spawns len(srcs) goroutines.
func Merge[T any](done Done, srcs ...<-chan T) <-chan T {
	ret := make(chan T)
	if len(srcs) == 0 {
		close(ret)
		return ret
	}
	var liveRoutines atomic.Int64
	liveRoutines.Add(int64(len(srcs)))
	for _, src := range srcs {
		go func() {
			defer func() {
				if liveRoutines.Add(-1) > 0 {
					return
				}
				close(ret)
			}()
			for {
				select {
				case v, ok := <-src:
					if !ok {
						return
					}
					if !TrySend(done, ret, v) {
						return
					}
				case <-done:
					return
				}
			}
		}()
	}
	return ret
}
