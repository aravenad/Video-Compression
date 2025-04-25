package queue

import (
	"sync"

	"github.com/yourorg/video-compressor/internal/compressor"
)

// Task represents one compression job, including the ffmpeg arguments to use.
type Task struct {
	Source      string   // input path
	Destination string   // output path
	Args        []string // ffmpeg args (e.g. from presets.BuildFFArgs)
}

// Result holds the outcome of a Task.
type Result struct {
	Task Task
	Err  error
}

// Queue manages and runs tasks concurrently.
type Queue struct {
	tasks       []Task
	concurrency int
}

// Define a variable for the compress function to allow mocking in tests
var compressFunc = compressor.Compress

// GetCompressFunc returns the current compressFunc - useful for testing
func GetCompressFunc() func(string, string, []string) error {
	return compressFunc
}

// SetCompressFunc replaces the current compressFunc - useful for testing
func SetCompressFunc(fn func(string, string, []string) error) {
	compressFunc = fn
}

// New creates a Queue with a given concurrency level.
func New(concurrency int) *Queue {
	return &Queue{
		concurrency: concurrency,
		tasks:       make([]Task, 0),
	}
}

// Add enqueues a Task.
func (q *Queue) Add(task Task) {
	q.tasks = append(q.tasks, task)
}

// Run processes all queued tasks and blocks until done.
// It returns a slice of Results corresponding to each Task.
func (q *Queue) Run() []Result {
	taskChan := make(chan Task)
	resultChan := make(chan Result)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < q.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				err := compressFunc(task.Source, task.Destination, task.Args)
				resultChan <- Result{Task: task, Err: err}
			}
		}()
	}

	// Feed tasks into taskChan
	go func() {
		for _, task := range q.tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	// Close resultChan once all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]Result, 0, len(q.tasks))
	for res := range resultChan {
		results = append(results, res)
	}

	return results
}
