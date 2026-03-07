package chans

import "math"

// Unbound introduces an unbounded buffer to src.
//
// Since this is a dangerous operation, it allows to provide an optional warn function that
// is called every time the buffer goes above warnThreshold, and called again with
// false once it goes below warnThreshold/2.
//
// Unbound spawns a goroutine.
func Unbound[T any](done Done, src <-chan T, warnThreshold uint, warn func(aboveThreshold bool)) <-chan T {
	ret := make(chan T)
	if warn == nil || warnThreshold <= 0 {
		warnThreshold = math.MaxUint64
		warn = func(bool) {}
	}
	go func() {
		defer close(ret)
		var (
			buf     []T
			sendRet chan T
			warned  bool
		)

		getVal := func() (t T) {
			if len(buf) == 0 {
				// We'll never send this because when buf is empty
				// sendRet is nil.
				return t
			}

			size := uint(len(buf))
			if !warned && size > warnThreshold {
				warned = true
				warn(true)
			}
			if warned && size < warnThreshold/2 {
				warned = false
				warn(false)
			}
			return buf[0]
		}

		for src != nil || sendRet != nil {
			select {
			case <-done:
				return
			case v, ok := <-src:
				if !ok {
					src = nil
					continue
				}
				buf = append(buf, v)
				sendRet = ret
			case sendRet <- getVal():
				buf = buf[1:]
				if len(buf) == 0 {
					sendRet = nil
				}
			}
		}
	}()
	return ret
}
