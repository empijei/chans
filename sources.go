package chans

import "iter"

// FromSlice emits all values from src.
//
// FromSlice spawns a goroutine.
func FromSlice[T any](done Done, src []T) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		for _, v := range src {
			TrySend(done, ret, v)
		}
	}()
	return ret
}

// FromSliceCopy copies the slice into a buffered channel.
func FromSliceCopy[T any](src []T) <-chan T {
	ret := make(chan T, len(src))
	for _, v := range src {
		ret <- v
	}
	close(ret)
	return ret
}

// FromSeq emits every time the src emits.
//
// FromSeq spawns a goroutine.
func FromSeq[T any](done Done, src iter.Seq[T]) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		for v := range src {
			if !TrySend(done, ret, v) {
				return
			}
		}
	}()
	return ret
}

// FromSeq2 is like [FromSeq] but if firs sends the first value on the first channel,
// then the second value on the second channel.
func FromSeq2[T, V any](done Done, src iter.Seq2[T, V]) (<-chan T, <-chan V) {
	ret, rev := make(chan T), make(chan V)
	go func() {
		defer close(ret)
		defer close(rev)
		for t, v := range src {
			if !TrySend(done, ret, t) {
				return
			}
			if !TrySend(done, rev, v) {
				return
			}
		}
	}()
	return ret, rev
}

// FromSeqUnwrap is like [FromSeq2] but it only emits once per iteration, depending
// on whether there was an error.
func FromSeqUnwrap[T any](done Done, src iter.Seq2[T, error]) (<-chan T, <-chan error) {
	ret, rev := make(chan T), make(chan error)
	go func() {
		defer close(ret)
		defer close(rev)
		for t, err := range src {
			if err != nil {
				if !TrySend(done, rev, err) {
					return
				}
				continue
			}
			if !TrySend(done, ret, t) {
				return
			}
		}
	}()
	return ret, rev
}
