package chans

import "time"

// Done is the type for the done channels.
type Done = <-chan struct{}

// Sleep waits dur and returns.
func Sleep(done Done, dur time.Duration) {
	select {
	case <-done:
	case <-time.After(dur):
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
