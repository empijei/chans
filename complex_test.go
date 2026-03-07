package chans_test

import (
	"testing"
	"time"

	"github.com/empijei/chans"
	"github.com/empijei/tst"
)

func TestUnboundPass(t *testing.T) {
	ctx := tst.Go(t)
	data := []int{1, 2, 3, 4, 5}
	src := chans.FromSlice(ctx.Done(), data)
	got := chans.ToSlice(ctx.Done(), chans.Unbound(ctx.Done(), src, 0, nil))
	tst.Is(data, got, t)
}

func TestUnboundOverflow(t *testing.T) {
	ctx := tst.Go(t)
	src := make(chan int)
	const size = 1000
	unbound := chans.Unbound(ctx.Done(), src, 0, nil)

	// Send everything without reading.
	go func() {
		defer close(src)
		for i := range size {
			src <- i
		}
	}()

	// Wait a bit to ensure they are all in the buffer.
	chans.Sleep(ctx.Done(), 50*time.Millisecond)

	var got []int
	for range size {
		got = append(got, <-unbound)
	}
	tst.Is(size, len(got), t)
}

func TestUnboundWarn(t *testing.T) {
	ctx := tst.Go(t)
	src := make(chan int)
	warns := make(chan bool, 10)
	warn := func(above bool) {
		warns <- above
	}

	const threshold = 10
	unbound := chans.Unbound(ctx.Done(), src, threshold, warn)

	// Fill up to threshold + 1
	for i := range threshold + 1 {
		src <- i
	}

	// Should have received a warning upon first read.
	<-unbound
	select {
	case w := <-warns:
		tst.Is(true, w, t)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for warning")
	}

	// Drain until threshold / 2 - 1.
	// We already read 1. We need to reach size < 5 (so size 4).
	// Current size is 10.
	// We need 6 more reads to reach size 4.
	// 1st read: size 10 -> size 9
	// 2nd read: size 9 -> size 8
	// 3rd read: size 8 -> size 7
	// 4th read: size 7 -> size 6
	// 5th read: size 6 -> size 5
	// 6th read: size 5 -> size 4
	// 7th read: size 4 -> size 3. At this call, size is 4, so it should trigger warn(false).
	for range 7 {
		<-unbound
	}

	// Should have received an "all clear"
	select {
	case w := <-warns:
		tst.Is(false, w, t)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for all clear")
	}
	close(src)
	for range unbound {
	}
}

func TestMulticast(t *testing.T) {
	ctx := tst.Go(t)
	src := make(chan int)
	m := chans.NewMulticast(ctx.Done(), src)

	s1 := m.Subscribe(ctx.Done())
	s2 := m.Subscribe(ctx.Done())

	go func() {
		src <- 42
		src <- 43
		close(src)
	}()

	tst.Is(42, <-s1, t)
	tst.Is(42, <-s2, t)
	tst.Is(43, <-s1, t)
	tst.Is(43, <-s2, t)

	_, ok1 := <-s1
	_, ok2 := <-s2
	tst.Is(false, ok1, t)
	tst.Is(false, ok2, t)
}

func TestMulticastLateSubscribe(t *testing.T) {
	ctx := tst.Go(t)
	src := make(chan int)
	m := chans.NewMulticast(ctx.Done(), src)

	s1 := m.Subscribe(ctx.Done())

	go func() {
		src <- 42
	}()

	tst.Is(42, <-s1, t)

	s2 := m.Subscribe(ctx.Done())
	go func() {
		src <- 43
		close(src)
	}()

	tst.Is(43, <-s1, t)
	tst.Is(43, <-s2, t)
}
