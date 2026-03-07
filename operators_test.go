package chans_test

import (
	"testing"
	"time"

	"github.com/empijei/chans"
	"github.com/empijei/tst"
)

func TestDebounceDiscard(t *testing.T) {
	ctx := tst.Go(t)
	const size = 10
	in := make(chan int, size)
	for i := range size {
		in <- i
	}
	close(in)
	got := chans.Debounce(ctx.Done(), in, 100*time.Millisecond)
	tst.Is(9, <-got, t)
	_, ok := <-got
	tst.Is(false, ok, t)
}

func TestDebounceKeep(t *testing.T) {
	ctx := tst.Go(t)
	const size = 5
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range size {
			chans.Sleep(ctx.Done(), 10*time.Millisecond)
			in <- i
		}
	}()
	res := chans.Debounce(ctx.Done(), in, 1*time.Microsecond)
	got := chans.ToSlice(ctx.Done(), res)
	tst.Is([]int{0, 1, 2, 3, 4}, got, t)
}

func TestThrottleDrop(t *testing.T) {
	ctx := tst.Go(t)
	const size = 10
	in := make(chan int, size)
	for i := range size {
		in <- i + 1
	}
	close(in)
	got := chans.Throttle(ctx.Done(), in, 100*time.Millisecond)
	tst.Is(1, <-got, t)
	_, ok := <-got
	tst.Is(false, ok, t)
}

func TestThrottleKeep(t *testing.T) {
	ctx := tst.Go(t)
	const size = 5
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range size {
			chans.Sleep(ctx.Done(), 10*time.Millisecond)
			in <- i
		}
	}()
	res := chans.Throttle(ctx.Done(), in, 1*time.Microsecond)
	got := chans.ToSlice(ctx.Done(), res)
	tst.Is([]int{0, 1, 2, 3, 4}, got, t)
}

func TestWindowDrop(t *testing.T) {
	ctx := tst.Go(t)
	const size = 10
	in := make(chan int, size)
	for i := range size {
		in <- i + 1
	}
	got := chans.Window(ctx.Done(), in, 100*time.Millisecond)
	tst.Is(1, <-got, t)
	tst.Is(10, <-got, t)
	close(in)
	_, ok := <-got
	tst.Is(false, ok, t)
}

func TestWindowKeep(t *testing.T) {
	ctx := tst.Go(t)
	const size = 5
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range size {
			chans.Sleep(ctx.Done(), 10*time.Millisecond)
			in <- i
		}
	}()
	res := chans.Window(ctx.Done(), in, 1*time.Microsecond)
	got := chans.ToSlice(ctx.Done(), res)
	tst.Is([]int{0, 1, 2, 3, 4}, got, t)
}
