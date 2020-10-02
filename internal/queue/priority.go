package queue

import (
	"context"
	"errors"
	"fmt"

	"github.com/bkrebsbach/simple-job-queue/internal/domain"
)

// JobQueuer defines an basic interface for a job queue
type JobQueuer interface {
	Enqueue(ctx context.Context, job domain.Job) (int, error)
	Dequeue(ctx context.Context, consumerID string) (domain.Job, error)
	Conclude(ctx context.Context, jobID int, consumerID string) error
	FetchJob(ctx context.Context, jobID int) (domain.Job, error)
	CancelJob(ctx context.Context, jobID int) error
}

// NOTE: job IDs are currently constructed in the InMemoryQueue implemenation.
// Because lookups of jobs are done by ID (without reference to type), they need
// to be unique between both high and low priority queues. These IDs should be constructed
// in the MultiPriorityQueue, and passed to the JobQueuer implementations. The
// MultiPriorityQueue would need to maintain an in-memory map (as a substitute
// for a lookup table) of jobID to queue (and possibly other metadata as time goes on).
//
// The InMemoryQueue would need to be updated to take a job ID if passed in.
//

type MultiPriorityQueue struct {
	LowPriorityQueue  JobQueuer
	HighPriorityQueue JobQueuer
}

func (q *MultiPriorityQueue) Enqueue(
	ctx context.Context,
	job domain.Job,
) (int, error) {
	switch job.Type {
	case domain.JobTypeTimeCritical:
		return q.HighPriorityQueue.Enqueue(ctx, job)
	case domain.JobTypeNotTimeCritical:
		return q.LowPriorityQueue.Enqueue(ctx, job)
	}

	return 0, fmt.Errorf("unable to enqueue job")
}

func (q *MultiPriorityQueue) Dequeue(ctx context.Context, consumerID string) (domain.Job, error) {
	job, err := q.HighPriorityQueue.Dequeue(ctx, consumerID)
	if err != nil {
		switch {
		// if there are no jobs in the high priority queue, dequeue from
		// the low priority queue
		case errors.Is(err, domain.ErrQueueEmpty):
			job, err = q.HighPriorityQueue.Dequeue(ctx, consumerID)
			if err != nil {
				return domain.Job{}, err
			}
		default:
			return domain.Job{}, fmt.Errorf("error dequeueing from high priority queue: %w", err)
		}
	}

	return job, nil
}

func (q *MultiPriorityQueue) Conclude(ctx context.Context, jobID int, consumerID string) {
	return nil
}
func (q *MultiPriorityQueue) FetchJob(ctx context.Context, jobID int) (domain.Job, error) {
	return nil, nil
}
func (q *MultiPriorityQueue) CancelJob(ctx context.Context, jobID int) error {
	return nil
}
