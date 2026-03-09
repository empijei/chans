package chans_test

import (
	"testing"

	"github.com/empijei/chans"
	"github.com/empijei/tst"
)

func TestTap(t *testing.T) {
	done := tst.Go(t).Done()
	want := []int{1, 2, 3, 4}
	src := chans.FromSlice(done, want)
	var got []int
	tap := func(i int) {
		got = append(got, i)
	}
	res := chans.Tap(done, src, tap)
	gotRes := chans.ToSlice(done, res)
	tst.Is(want, got, t)
	tst.Is(want, gotRes, t)
}
