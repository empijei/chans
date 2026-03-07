package chans_test

import (
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
