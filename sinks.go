package chans

// ToSlice returns all values emitted by src.
func ToSlice[T any](done <-chan struct{}, src <-chan T) []T {
	var accum []T
	for {
		select {
		case v, ok := <-src:
			if !ok {
				return accum
			}
			accum = append(accum, v)
		case <-done:
			return accum
		}
	}
}
