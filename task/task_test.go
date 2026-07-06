package task

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestFuncRun(t *testing.T) {
	f := Func{Label: "greet", Fn: func(context.Context) (any, error) { return "hi", nil }}
	if f.Name() != "greet" {
		t.Fatalf("name = %q", f.Name())
	}
	res, err := f.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Task != "greet" || res.Output != "hi" {
		t.Fatalf("res = %+v", res)
	}
	if res.Started.IsZero() {
		t.Fatal("Started not set")
	}
}

func TestFuncRunError(t *testing.T) {
	boom := errors.New("boom")
	f := Func{Label: "bad", Fn: func(context.Context) (any, error) { return nil, boom }}
	_, err := f.Run(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("got %v", err)
	}
}

func mkTask(name string, out any, err error) Task {
	return Func{Label: name, Fn: func(context.Context) (any, error) { return out, err }}
}

func TestSequential(t *testing.T) {
	res, err := Sequential{}.Run(context.Background(),
		mkTask("a", 1, nil),
		mkTask("b", 2, nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 || res[0].Output != 1 || res[1].Output != 2 {
		t.Fatalf("res = %+v", res)
	}
}

func TestSequentialStopOnError(t *testing.T) {
	var ran atomic.Int32
	count := func(name string, err error) Task {
		return Func{Label: name, Fn: func(context.Context) (any, error) {
			ran.Add(1)
			return nil, err
		}}
	}
	_, err := Sequential{StopOnError: true}.Run(context.Background(),
		count("a", nil),
		count("b", errors.New("stop")),
		count("c", nil),
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if ran.Load() != 2 {
		t.Fatalf("ran %d tasks, expected to stop after 2", ran.Load())
	}
}

func TestSequentialJoinsErrors(t *testing.T) {
	e1, e2 := errors.New("e1"), errors.New("e2")
	_, err := Sequential{}.Run(context.Background(),
		mkTask("a", nil, e1),
		mkTask("b", nil, e2),
	)
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("expected both errors, got %v", err)
	}
}

func TestConcurrent(t *testing.T) {
	var running, maxSeen atomic.Int32
	slow := func(name string) Task {
		return Func{Label: name, Fn: func(context.Context) (any, error) {
			n := running.Add(1)
			for {
				m := maxSeen.Load()
				if n <= m || maxSeen.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			running.Add(-1)
			return name, nil
		}}
	}
	tasks := []Task{slow("a"), slow("b"), slow("c"), slow("d")}
	res, err := Concurrent{Limit: 2}.Run(context.Background(), tasks...)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 4 || res[0].Output != "a" || res[3].Output != "d" {
		t.Fatalf("results out of order: %+v", res)
	}
	if maxSeen.Load() > 2 {
		t.Fatalf("concurrency limit exceeded: %d", maxSeen.Load())
	}
}

func TestScheduleRun(t *testing.T) {
	var count atomic.Int32
	tk := Func{Label: "tick", Fn: func(context.Context) (any, error) {
		count.Add(1)
		return nil, nil
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	err := Every(10*time.Millisecond).WithImmediate().Run(ctx, tk, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	// Immediate fire + at least one tick.
	if count.Load() < 2 {
		t.Fatalf("fired %d times", count.Load())
	}
}

func TestScheduleNoInterval(t *testing.T) {
	if err := (Schedule{}).Run(context.Background(), mkTask("x", nil, nil), nil); err != nil {
		t.Fatalf("zero schedule should be a no-op, got %v", err)
	}
}
