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
