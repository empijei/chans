package chans_test

import (
	"errors"
	"testing"

	"github.com/empijei/chans"
	"github.com/empijei/tst"
)

func TestFromSlice(t *testing.T) {
	done := tst.Go(t).Done()
	data := []int{1, 2, 3}
	c1 := chans.FromSlice(done, data)
	c2 := chans.FromSliceCopy(data)
	got1, got2 := chans.ToSlice(done, c1), chans.ToSlice(done, c2)
	tst.Is(got1, got2, t)
}

func TestFromSeq(t *testing.T) {
	done := tst.Go(t).Done()
	data := []int{1, 2, 3}
	src := func(yield func(int) bool) {
		for _, v := range data {
			if !yield(v) {
				return
			}
		}
	}
	got := chans.ToSlice(done, chans.FromSeq(done, src))
	tst.Is(data, got, t)
}

func TestFromSeq2(t *testing.T) {
	done := tst.Go(t).Done()
	type pair struct {
		k int
		v string
	}
	data := []pair{{1, "one"}, {2, "two"}}
	src := func(yield func(int, string) bool) {
		for _, p := range data {
			if !yield(p.k, p.v) {
				return
			}
		}
	}
	c1, c2 := chans.FromSeq2(done, src)
	got1, got2 := chans.ToSlices(done, c1, c2)
	tst.Is([]int{1, 2}, got1, t)
	tst.Is([]string{"one", "two"}, got2, t)
}

func TestFromSeqUnwrap(t *testing.T) {
	done := tst.Go(t).Done()
	err := errors.New("ops")
	src := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, err) {
			return
		}
		if !yield(2, nil) {
			return
		}
	}
	c, errs := chans.FromSeqUnwrap(done, src)
	got, gotErrs := chans.ToSlices(done, c, errs)
	tst.Is([]int{1, 2}, got, t)
	tst.Is([]error{err}, gotErrs, t)
}
