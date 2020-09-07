package queue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bkrebsbach/simple-job-queue/internal/domain"
)

// NOTE: I ran out of time writing tests, so I only have the first few functions
// captured here.

func TestEnqueue(t *testing.T) {
	mq := NewInMemoryQueue()

	job := domain.Job{
		Type:   domain.JobTypeTimeCritical,
		Status: domain.JobStatusQueued,
	}

	// check that job is created successfully
	jobID, err := mq.Enqueue(context.Background(), job)
	require.Equal(t, jobID, 1)
	require.Nil(t, err)

	// check that job ID increments
	jobID, err = mq.Enqueue(context.Background(), job)
	require.Equal(t, jobID, 2)
	require.Nil(t, err)

	// check that job ID increments as expected
	mq.maxID = 1
	jobID, err = mq.Enqueue(context.Background(), job)
	require.Equal(t, jobID, 3)
	require.Nil(t, err)
}

func TestDequeue(t *testing.T) {
	mq := NewInMemoryQueue()

	jobs := map[int]domain.Job{
		1: {
			ID:     1,
			Type:   domain.JobTypeTimeCritical,
			Status: domain.JobStatusInProgress,
		},
		2: {
			ID:     2,
			Type:   domain.JobTypeNotTimeCritical,
			Status: domain.JobStatusConcluded,
		},
		3: {
			ID:     3,
			Type:   domain.JobTypeNotTimeCritical,
			Status: domain.JobStatusQueued,
		},
	}
	queue := []int{1, 2, 3}

	mq.jobs = jobs
	mq.queue = queue

	// check that jobs are dequeued successfully
	job, err := mq.Dequeue(context.Background())
	require.Nil(t, err)

	require.Equal(t, len(mq.queue), 0)
	require.Equal(t, job, jobs[3])
}
