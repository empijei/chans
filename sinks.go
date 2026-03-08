package chans

// ToSlice returns all values emitted by src.
func ToSlice[T any](done Done, src <-chan T) []T {
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

// ToSlices consumes both src at the same time, returning the accumulated set of
// values.
func ToSlices[T, V any](done Done, src1 <-chan T, src2 <-chan V) ([]T, []V) {
	var retT []T
	var retV []V
	for src1 != nil || src2 != nil {
		select {
		case <-done:
			return retT, retV
		case v, ok := <-src1:
			if !ok {
				src1 = nil
				continue
			}
			retT = append(retT, v)
		case v, ok := <-src2:
			if !ok {
				src2 = nil
				continue
			}
			retV = append(retV, v)
		}
	}
	return retT, retV
}
