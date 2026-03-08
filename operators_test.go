package chans_test

import (
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/empijei/chans"
	"github.com/empijei/tst"
)

func TestDebounceDiscard(t *testing.T) {
	done := tst.Go(t).Done()
	const size = 10
	in := make(chan int, size)
	for i := range size {
		in <- i
	}
	close(in)
	got := chans.Debounce(done, in, 100*time.Millisecond)
	tst.Is(9, <-got, t)
	_, ok := <-got
	tst.Is(false, ok, t)
}

func TestDebounceKeep(t *testing.T) {
	done := tst.Go(t).Done()
	const size = 5
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range size {
			chans.Sleep(done, 10*time.Millisecond)
			in <- i
		}
	}()
	res := chans.Debounce(done, in, 1*time.Microsecond)
	got := chans.ToSlice(done, res)
	tst.Is([]int{0, 1, 2, 3, 4}, got, t)
}

func TestThrottleDrop(t *testing.T) {
	done := tst.Go(t).Done()
	const size = 10
	in := make(chan int, size)
	for i := range size {
		in <- i + 1
	}
	close(in)
	got := chans.Throttle(done, in, 100*time.Millisecond)
	tst.Is(1, <-got, t)
	_, ok := <-got
	tst.Is(false, ok, t)
}

func TestThrottleKeep(t *testing.T) {
	done := tst.Go(t).Done()
	const size = 5
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range size {
			chans.Sleep(done, 10*time.Millisecond)
			in <- i
		}
	}()
	res := chans.Throttle(done, in, 1*time.Microsecond)
	got := chans.ToSlice(done, res)
	tst.Is([]int{0, 1, 2, 3, 4}, got, t)
}

func TestWindowDrop(t *testing.T) {
	done := tst.Go(t).Done()
	const size = 10
	in := make(chan int, size)
	for i := range size {
		in <- i + 1
	}
	got := chans.Window(done, in, 100*time.Millisecond)
	tst.Is(1, <-got, t)
	tst.Is(10, <-got, t)
	close(in)
	_, ok := <-got
	tst.Is(false, ok, t)
}

func TestWindowKeep(t *testing.T) {
	done := tst.Go(t).Done()
	const size = 5
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range size {
			chans.Sleep(done, 10*time.Millisecond)
			in <- i
		}
	}()
	res := chans.Window(done, in, 1*time.Microsecond)
	got := chans.ToSlice(done, res)
	tst.Is([]int{0, 1, 2, 3, 4}, got, t)
}

func TestConcat(t *testing.T) {
	done := tst.Go(t).Done()

	t.Run("MultipleSources", func(t *testing.T) {
		s1 := chans.FromSlice(done, []int{1, 2})
		s2 := chans.FromSlice(done, []int{3, 4})
		s3 := chans.FromSlice(done, []int{5})

		res := chans.Concat(done, s1, s2, s3)
		got := chans.ToSlice(done, res)
		tst.Is([]int{1, 2, 3, 4, 5}, got, t)
	})

	t.Run("EmptySources", func(t *testing.T) {
		res := chans.Concat[int](done)
		got := chans.ToSlice(done, res)
		tst.Is([]int(nil), got, t)
	})

	t.Run("ClosedSources", func(t *testing.T) {
		s1 := make(chan int)
		close(s1)
		s2 := chans.FromSlice(done, []int{1})
		res := chans.Concat(done, s1, s2)
		got := chans.ToSlice(done, res)
		tst.Is([]int{1}, got, t)
	})
}

func TestMap(t *testing.T) {
	done := tst.Go(t).Done()
	src := chans.FromSlice(done, []int{1, 2, 3})
	res := chans.Map(done, src, func(i int) int { return i * 2 })
	got := chans.ToSlice(done, res)
	tst.Is([]int{2, 4, 6}, got, t)
}

func TestMapTry(t *testing.T) {
	done := tst.Go(t).Done()
	t.Run("Success", func(t *testing.T) {
		src := chans.FromSlice(done, []int{1, 2, 3})
		res, errs := chans.MapTry(done, src, func(i int) (int, error) {
			return i * 2, nil
		})
		got, gotErrs := chans.ToSlices(done, res, errs)
		tst.Is([]int{2, 4, 6}, got, t)
		tst.Is([]error(nil), gotErrs, t)
	})
	t.Run("Failure", func(t *testing.T) {
		src := chans.FromSlice(done, []int{1, 2, 3})
		res, errs := chans.MapTry(done, src, func(i int) (int, error) {
			if i == 2 {
				return 0, errors.New("fail")
			}
			return i * 2, nil
		})
		got, gotErrs := chans.ToSlices(done, res, errs)
		tst.Is([]int{2, 6}, got, t)
		tst.Is(1, len(gotErrs), t)
		tst.Err("fail", gotErrs[0], t)
	})
}

func TestMapParallel(t *testing.T) {
	done := tst.Go(t).Done()
	src := chans.FromSlice(done, []int{1, 2, 3, 4, 5})
	res := chans.MapParallel(done, src, 3, func(i int) int { return i * 2 })
	got := chans.ToSlice(done, res)
	sort.Ints(got)
	tst.Is([]int{2, 4, 6, 8, 10}, got, t)
}

func TestMapParallelTry(t *testing.T) {
	done := tst.Go(t).Done()
	t.Run("Success", func(t *testing.T) {
		src := chans.FromSlice(done, []int{1, 2, 3})
		res, errs := chans.MapParallelTry(done, src, 2, func(i int) (int, error) {
			return i * 2, nil
		})
		got, gotErrs := chans.ToSlices(done, res, errs)
		sort.Ints(got)
		tst.Is([]int{2, 4, 6}, got, t)
		tst.Is([]error(nil), gotErrs, t)
	})
	t.Run("Failure", func(t *testing.T) {
		src := chans.FromSlice(done, []int{1, 2, 3})
		res, errs := chans.MapParallelTry(done, src, 2, func(i int) (int, error) {
			if i == 2 {
				return 0, errors.New("fail")
			}
			return i * 2, nil
		})
		got, gotErrs := chans.ToSlices(done, res, errs)
		sort.Ints(got)
		tst.Is([]int{2, 6}, got, t)
		tst.Is(1, len(gotErrs), t)
		tst.Err("fail", gotErrs[0], t)
	})
}

func TestFilter(t *testing.T) {
	done := tst.Go(t).Done()
	src := chans.FromSlice(done, []int{1, 2, 3, 4, 5})
	res := chans.Filter(done, src, func(i int) bool { return i%2 == 0 })
	got := chans.ToSlice(done, res)
	tst.Is([]int{2, 4}, got, t)
}

func TestMerge(t *testing.T) {
	done := tst.Go(t).Done()
	t.Run("MultipleSources", func(t *testing.T) {
		s1 := chans.FromSlice(done, []int{1, 2})
		s2 := chans.FromSlice(done, []int{3, 4})
		s3 := chans.FromSlice(done, []int{5})

		res := chans.Merge(done, s1, s2, s3)
		got := chans.ToSlice(done, res)
		sort.Ints(got)
		tst.Is([]int{1, 2, 3, 4, 5}, got, t)
	})
	t.Run("EmptySources", func(t *testing.T) {
		res := chans.Merge[int](done)
		got := chans.ToSlice(done, res)
		tst.Is([]int(nil), got, t)
	})
}
