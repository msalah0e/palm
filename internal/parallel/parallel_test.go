package parallel

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestRun_Success(t *testing.T) {
	tasks := []Task{
		{Name: "task1", Fn: func() (string, error) { return "", nil }},
		{Name: "task2", Fn: func() (string, error) { return "", nil }},
		{Name: "task3", Fn: func() (string, error) { return "", nil }},
	}

	results := Run(tasks, 4)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.OK {
			t.Errorf("task %s should be OK", r.Name)
		}
		if r.Err != nil {
			t.Errorf("task %s should have no error", r.Name)
		}
	}
}

func TestRun_WithErrors(t *testing.T) {
	tasks := []Task{
		{Name: "ok-task", Fn: func() (string, error) { return "", nil }},
		{Name: "fail-task", Fn: func() (string, error) { return "some output", fmt.Errorf("simulated failure") }},
	}

	results := Run(tasks, 4)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Results should be in order
	if !results[0].OK {
		t.Error("first task should be OK")
	}
	if results[1].OK {
		t.Error("second task should have failed")
	}
	if results[1].Err == nil {
		t.Error("second task should have error")
	}
	if results[1].Output != "some output" {
		t.Errorf("expected output %q, got %q", "some output", results[1].Output)
	}
}

func TestRun_Concurrency(t *testing.T) {
	var maxConcurrent int64
	var current int64

	tasks := make([]Task, 10)
	for i := range tasks {
		tasks[i] = Task{
			Name: fmt.Sprintf("task-%d", i),
			Fn: func() (string, error) {
				c := atomic.AddInt64(&current, 1)
				// Track max concurrent
				for {
					old := atomic.LoadInt64(&maxConcurrent)
					if c <= old || atomic.CompareAndSwapInt64(&maxConcurrent, old, c) {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt64(&current, -1)
				return "", nil
			},
		}
	}

	results := Run(tasks, 2) // Limit to 2 concurrent

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}

	if maxConcurrent > 2 {
		t.Errorf("max concurrent should be <= 2, got %d", maxConcurrent)
	}
}

func TestRun_DefaultConcurrency(t *testing.T) {
	tasks := []Task{
		{Name: "test", Fn: func() (string, error) { return "", nil }},
	}

	// Should not panic with 0 concurrency (defaults to 4)
	results := Run(tasks, 0)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestRun_TimingTracked(t *testing.T) {
	tasks := []Task{
		{Name: "slow", Fn: func() (string, error) {
			time.Sleep(50 * time.Millisecond)
			return "", nil
		}},
	}

	results := Run(tasks, 1)
	if results[0].Elapsed < 50*time.Millisecond {
		t.Errorf("expected elapsed >= 50ms, got %v", results[0].Elapsed)
	}
}

func TestRun_OutputCaptured(t *testing.T) {
	tasks := []Task{
		{Name: "with-output", Fn: func() (string, error) { return "hello world", nil }},
	}

	results := Run(tasks, 1)
	if results[0].Output != "hello world" {
		t.Errorf("expected output %q, got %q", "hello world", results[0].Output)
	}
}
