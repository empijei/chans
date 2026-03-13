package chans

import "time"

// Done is the type for the done channels.
type Done = <-chan struct{}

// Sleep waits dur and returns.
//
// It returns whether the time actually passed.
func Sleep(done Done, dur time.Duration) (waited bool) {
	select {
	case <-done:
		return false
	case <-time.After(dur):
		return true
	}
}

// Tap mirrors source emissions, calling tap on each one.
//
// Tap spawns a goroutine.
func Tap[T any](done Done, src <-chan T, tap func(T)) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		for {
			select {
			case <-done:
				return
			case v, ok := <-src:
				if !ok {
					return
				}
				tap(v)
				if !TrySend(done, ret, v) {
					return
				}
			}
		}
	}()
	return ret
}
