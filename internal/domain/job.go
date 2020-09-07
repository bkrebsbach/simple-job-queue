package domain

const (
	JobTypeTimeCritical    = "TIME_CRITICAL"
	JobTypeNotTimeCritical = "NOT_TIME_CRITICAL"

	JobStatusQueued     = "QUEUED"
	JobStatusInProgress = "IN_PROGRESS"
	JobStatusConcluded  = "CONCLUDED"
)

var (
	JobStatuses = map[string]bool{
		JobStatusQueued:     true,
		JobStatusInProgress: true,
		JobStatusConcluded:  true,
	}

	JobTypes = map[string]bool{
		JobTypeTimeCritical:    true,
		JobTypeNotTimeCritical: true,
	}
)

type Job struct {
	ID     int
	Type   string
	Status string
}
