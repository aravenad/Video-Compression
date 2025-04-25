package queue

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Mock the compressor package to avoid real compression calls
func mockCompress(source, destination string, args []string) error {
	if source == "error.mp4" {
		return errors.New("mock compression error")
	}
	if source == "slow.mp4" {
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// TestNew verifies that New creates a Queue with the correct concurrency and empty tasks
func TestNew(t *testing.T) {
	q := New(5)
	if q.concurrency != 5 {
		t.Errorf("New(5): concurrency = %d; want 5", q.concurrency)
	}
	if len(q.tasks) != 0 {
		t.Errorf("New(5): tasks slice has %d elements; want 0", len(q.tasks))
	}
}

// TestAdd verifies that tasks are correctly added to the queue
func TestAdd(t *testing.T) {
	q := New(1)

	task1 := Task{Source: "file1.mp4", Destination: "out1.mp4", Args: []string{"-a", "1"}}
	q.Add(task1)

	if len(q.tasks) != 1 {
		t.Fatalf("After Add: len(tasks) = %d; want 1", len(q.tasks))
	}

	// Compare individual fields instead of the whole struct
	if q.tasks[0].Source != task1.Source ||
		q.tasks[0].Destination != task1.Destination ||
		!reflect.DeepEqual(q.tasks[0].Args, task1.Args) {
		t.Errorf("After Add: tasks[0] = %+v; want %+v", q.tasks[0], task1)
	}

	task2 := Task{Source: "file2.mp4", Destination: "out2.mp4", Args: []string{"-b", "2"}}
	q.Add(task2)

	if len(q.tasks) != 2 {
		t.Fatalf("After second Add: len(tasks) = %d; want 2", len(q.tasks))
	}

	// Compare individual fields instead of the whole struct
	if q.tasks[1].Source != task2.Source ||
		q.tasks[1].Destination != task2.Destination ||
		!reflect.DeepEqual(q.tasks[1].Args, task2.Args) {
		t.Errorf("After second Add: tasks[1] = %+v; want %+v", q.tasks[1], task2)
	}
}

// TestRun_Success verifies that all tasks are processed and results collected
func TestRun_Success(t *testing.T) {
	// Replace the compressor.Compress function with our mock
	origCompress := compressFunc
	compressFunc = mockCompress
	defer func() { compressFunc = origCompress }()

	q := New(2)
	task1 := Task{Source: "file1.mp4", Destination: "out1.mp4", Args: []string{"-a", "1"}}
	task2 := Task{Source: "file2.mp4", Destination: "out2.mp4", Args: []string{"-b", "2"}}
	q.Add(task1)
	q.Add(task2)

	results := q.Run()

	if len(results) != 2 {
		t.Fatalf("Run: got %d results; want 2", len(results))
	}

	// Map of expected results by source
	expected := map[string]Task{
		task1.Source: task1,
		task2.Source: task2,
	}

	for _, res := range results {
		task, ok := expected[res.Task.Source]
		if !ok {
			t.Errorf("Unexpected result for source: %s", res.Task.Source)
			continue
		}

		// Compare individual fields instead of the whole struct
		if res.Task.Source != task.Source ||
			res.Task.Destination != task.Destination ||
			!reflect.DeepEqual(res.Task.Args, task.Args) {
			t.Errorf("Result task = %+v; want %+v", res.Task, task)
		}

		if res.Err != nil {
			t.Errorf("Result error = %v; want nil", res.Err)
		}
		delete(expected, res.Task.Source)
	}

	if len(expected) > 0 {
		t.Errorf("Missing results for tasks: %v", expected)
	}
}

// TestRun_Error verifies that errors from compression tasks are properly collected
func TestRun_Error(t *testing.T) {
	// Replace the compressor.Compress function with our mock
	origCompress := compressFunc
	compressFunc = mockCompress
	defer func() { compressFunc = origCompress }()

	q := New(2)
	task1 := Task{Source: "file1.mp4", Destination: "out1.mp4", Args: []string{}}
	task2 := Task{Source: "error.mp4", Destination: "error-out.mp4", Args: []string{}}
	q.Add(task1)
	q.Add(task2)

	results := q.Run()

	if len(results) != 2 {
		t.Fatalf("Run: got %d results; want 2", len(results))
	}

	// Check that the expected error was returned
	for _, res := range results {
		if res.Task.Source == "error.mp4" {
			if res.Err == nil || res.Err.Error() != "mock compression error" {
				t.Errorf("For error.mp4: got error %v; want 'mock compression error'", res.Err)
			}
		} else if res.Err != nil {
			t.Errorf("For %s: got error %v; want nil", res.Task.Source, res.Err)
		}
	}
}

// TestRun_Concurrency verifies that the queue respects the concurrency limit
func TestRun_Concurrency(t *testing.T) {
	// Replace the compressor.Compress function with our instrumented version
	origCompress := compressFunc
	defer func() { compressFunc = origCompress }()

	// Track maximum concurrent executions
	var activeCount int32
	var maxActive int32
	var mu sync.Mutex

	compressFunc = func(source, destination string, args []string) error {
		// Increment active count and update max if needed
		newCount := atomic.AddInt32(&activeCount, 1)

		mu.Lock()
		if newCount > maxActive {
			maxActive = newCount
		}
		mu.Unlock()

		// Simulate work for a short time
		time.Sleep(50 * time.Millisecond)

		// Decrement active count
		atomic.AddInt32(&activeCount, -1)
		return nil
	}

	// Create a queue with 3 concurrent workers and 5 tasks
	concurrency := 3
	q := New(concurrency)

	for i := 0; i < 5; i++ {
		q.Add(Task{
			Source:      fmt.Sprintf("file%d.mp4", i),
			Destination: fmt.Sprintf("out%d.mp4", i),
			Args:        []string{},
		})
	}

	// Run the queue
	q.Run()

	// Check if the max concurrent executions matches expected concurrency
	if maxActive != int32(concurrency) {
		t.Errorf("Max concurrent executions = %d; want %d", maxActive, concurrency)
	}
}

// TestRun_EmptyQueue verifies that running an empty queue returns an empty slice
func TestRun_EmptyQueue(t *testing.T) {
	q := New(3)
	results := q.Run()

	if len(results) != 0 {
		t.Errorf("Empty queue Run() returned %d results; want 0", len(results))
	}
}

// TestGetCompressFunc verifies that GetCompressFunc returns the current compressFunc
func TestGetCompressFunc(t *testing.T) {
	// Save original function
	origCompress := compressFunc
	defer func() { compressFunc = origCompress }()

	// Set a known function
	testFunc := func(in, out string, args []string) error {
		return fmt.Errorf("test function")
	}
	compressFunc = testFunc

	// Get the function and verify it's the same one we set
	fn := GetCompressFunc()

	// Call both functions and compare their behavior
	err1 := testFunc("in", "out", []string{})
	err2 := fn("in", "out", []string{})

	if err1.Error() != err2.Error() {
		t.Errorf("GetCompressFunc didn't return the expected function: got %v, want %v", err2, err1)
	}
}

// TestSetCompressFunc verifies that SetCompressFunc changes the compressFunc variable
func TestSetCompressFunc(t *testing.T) {
	// Save original function
	origCompress := compressFunc
	defer func() { compressFunc = origCompress }()

	// Create a new function with known behavior
	testFunc := func(in, out string, args []string) error {
		return fmt.Errorf("set function test")
	}

	// Set the new function
	SetCompressFunc(testFunc)

	// Verify that compressFunc has been changed
	err := compressFunc("in", "out", []string{})
	if err == nil || err.Error() != "set function test" {
		t.Errorf("SetCompressFunc didn't change compressFunc: got %v", err)
	}
}
