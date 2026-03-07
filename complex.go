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

type sub[T any] struct {
	done Done
	c    chan T
}

// Multicast replicates all emissions from a source on all currently active subscriptions.
type Multicast[T any] struct {
	subs chan sub[T]
	done Done
}

// Subscribe returns all emissions from src.
//
// If a subscription doesn't consume a value, the multicast blocks and waits for it
// to do so, so it's important that unused subscriptions are cancelled by closing done.
func (m *Multicast[T]) Subscribe(done Done) <-chan T {
	c := make(chan T)
	select {
	case m.subs <- sub[T]{done: done, c: c}:
	case <-done:
		close(c)
	case <-m.done:
		close(c)
	}
	return c
}

// NewMulticast returns a new Multicast from src.
//
// It spawns a goroutine.
func NewMulticast[T any](done Done, src <-chan T) Multicast[T] {
	m := Multicast[T]{
		subs: make(chan sub[T]),
		done: done,
	}
	go m.run(src)
	return m
}

func (m *Multicast[T]) run(src <-chan T) {
	subscriptions := make(map[chan T]Done)
	defer func() {
		for s := range subscriptions {
			close(s)
		}
	}()

	for {
		select {
		case <-m.done:
			return
		case newSub := <-m.subs:
			subscriptions[newSub.c] = newSub.done
		case v, ok := <-src:
			if !ok {
				return
			}
			for s, sDone := range subscriptions {
				select {
				case s <- v:
				case <-sDone:
					delete(subscriptions, s)
					close(s)
				case <-m.done:
					return
				}
			}
		}
	}
}
