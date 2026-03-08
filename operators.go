package chans

import (
	"time"
)

// TODO write in the docs that done means what it means.
// TODO write that closure is propagated from src to returned chans.

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
