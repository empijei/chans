package chans

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
