package queue

import (
	"context"
	"sync"

	"github.com/bkrebsbach/simple-job-queue/internal/domain"
)

// InMemoryQueue is an in-memory implementation of a job queue. Job IDs are stored
// in a slice, and the job definitions are stored in a map with the job IDs as
// keys. Maps are unordered, so the slice is necessary to preserve ordering.
type InMemoryQueue struct {
	queue []int
	jobs  map[int]domain.Job
	maxID int

	lock sync.RWMutex
}

// NewInMemoryQueue returns an in-memory job queue.
func NewInMemoryQueue() *InMemoryQueue {
	queue := make([]int, 0)
	jobs := make(map[int]domain.Job)

	return &InMemoryQueue{
		queue: queue,
		jobs:  jobs,
		maxID: 0,
		lock:  sync.RWMutex{},
	}
}

// Enqueue adds a job to the queue, and returns the ID of the job.
func (q *InMemoryQueue) Enqueue(ctx context.Context, job domain.Job) (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// get the next available job ID
	id := q.maxID + 1
	_, ok := q.jobs[id]
	for ok {
		id++
		_, ok = q.jobs[id]
	}

	// add the ID to the job
	// add the job to the job list
	// add the job to the queue
	// update the max ID
	job.ID = id
	q.jobs[job.ID] = job
	q.queue = append(q.queue, job.ID)
	q.maxID = job.ID

	return job.ID, nil
}

// Dequeue returns a job from the queue. Jobs are considered available for
// Dequeue if the job has not been concluded and has not dequeued already.
func (q *InMemoryQueue) Dequeue(ctx context.Context) (domain.Job, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// pop jobs off the queue until one is available for processing
	for len(q.queue) > 0 {
		// pop the first job off the queue
		var jobID int
		jobID, q.queue = q.queue[0], q.queue[1:]

		// get the job definition
		job, ok := q.jobs[jobID]
		if !ok {
			return domain.Job{}, domain.ErrJobNotFound{JobID: jobID}
		}

		// if the job status is queued, dequeue the job, mark it as in progress,
		// and return it
		if job.Status == domain.JobStatusQueued {
			job.Status = domain.JobStatusInProgress
			q.jobs[job.ID] = job

			return job, nil
		}
	}

	// if there are no jobs in the queue, return an error
	return domain.Job{}, domain.ErrQueueEmpty
}

// Conclude finishes execution on the job.
func (q *InMemoryQueue) Conclude(ctx context.Context, jobID int) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	// check if the job is defined
	job, ok := q.jobs[jobID]
	if !ok {
		return domain.ErrJobNotFound{JobID: jobID}
	}

	job.Status = domain.JobStatusConcluded
	q.jobs[job.ID] = job

	return nil
}

// FetchJob returns a job definition for the given job ID.
func (q *InMemoryQueue) FetchJob(ctx context.Context, jobID int) (domain.Job, error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	// check if the job is defined
	job, ok := q.jobs[jobID]
	if !ok {
		return domain.Job{}, domain.ErrJobNotFound{JobID: jobID}
	}

	return job, nil
}
