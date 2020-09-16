package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/bkrebsbach/simple-job-queue/internal/domain"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog/hlog"
)

const (
	HeaderQueueConsumer = "QUEUE_CONSUMER"
)

// only the consumer that dequeued can conclude a job
// - consumer provides identity token (QUEUE_CONSUMER header)
// - store that with the job
// - validate the value in the conclude step

// job defines the JSON payload for a job.
type job struct {
	ID     int    `json:"ID"`
	Type   string `json:"Type"`
	Status string `json:"Status"`
}

type enqueueResponse struct {
	ID int `json:"ID"`
}

// JobQueuer defines an basic interface for a job queue
type JobQueuer interface {
	Enqueue(ctx context.Context, job domain.Job) (int, error)
	Dequeue(ctx context.Context, consumerID string) (domain.Job, error)
	Conclude(ctx context.Context, jobID int, consumerID string) error
	FetchJob(ctx context.Context, jobID int) (domain.Job, error)
	CancelJob(ctx context.Context, jobID int) error
}

// JobHandler provides the HTTP interface for queuing, dequeuing, and retrieving
// jobs from a job queue.
type JobHandler struct {
	JobQueuer JobQueuer
}

// EnqueueJob takes a job payload and adds it to the job queue.
func (h *JobHandler) EnqueueJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("handler", "EnqueueJob").Logger()

	var payload job
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Info().Err(err).Msg("unable to decode payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate job status (TODO: use validator)
	if _, ok := domain.JobStatuses[payload.Status]; !ok {
		log.Info().Msgf("invalid job status: %s", payload.Status)
		WriteErrorResponse(w, ErrInvalidInput, http.StatusBadRequest)
		return
	}

	// validate job type (TODO: use validator)
	if _, ok := domain.JobTypes[payload.Type]; !ok {
		log.Info().Msgf("invalid job type: %s", payload.Type)
		WriteErrorResponse(w, ErrInvalidInput, http.StatusBadRequest)
		return
	}

	// enqueue the job
	jobID, err := h.JobQueuer.Enqueue(ctx, domain.Job{
		Type:   payload.Type,
		Status: payload.Status,
	})
	if err != nil {
		log.Error().Err(err).
			Str("job_type", payload.Type).
			Str("job_status", payload.Status).
			Msgf("error enqueuing job")
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// marshal and return the job ID in the response
	response, err := json.Marshal(enqueueResponse{ID: jobID})
	if err != nil {
		log.Error().Err(err).
			Str("job_id", strconv.Itoa(jobID)).
			Msgf("error marshalling response")
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// DequeueJob returns the first avaiable job from the queue. It returns an not f
// found error if there are no available jobs.
func (h *JobHandler) DequeueJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("handler", "DequeueJob").Logger()

	// get the consumer id from the header
	consumerID := r.Header.Get(HeaderQueueConsumer)
	if consumerID == "" {
		log.Info().Msg("no valid queue consumer ID")
		WriteErrorResponse(w, ErrInvalidInput, http.StatusBadRequest)
		return
	}

	// dequeue a job
	dequeuedJob, err := h.JobQueuer.Dequeue(ctx, consumerID)
	if err != nil {
		if errors.Is(err, domain.ErrQueueEmpty) {
			log.Info().Err(err).Msg("no available jobs in queue")
			WriteErrorResponse(w, ErrQueueEmpty, http.StatusNotFound)
			return
		}

		log.Error().Err(err).Msgf("error dequeuing job")
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// marshal and return the job in the response
	response, err := json.Marshal(job{
		ID:     dequeuedJob.ID,
		Status: dequeuedJob.Status,
		Type:   dequeuedJob.Type,
	})
	if err != nil {
		log.Error().Err(err).
			Str("job_id", strconv.Itoa(dequeuedJob.ID)).
			Str("job_type", dequeuedJob.Type).
			Str("job_status", dequeuedJob.Status).
			Msgf("error marshalling response")
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// ConcludeJob finishes execution on a job.
func (h *JobHandler) ConcludeJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("handler", "ConcludeJob").Logger()

	// validate jobID param is a non-negative integer
	paramJobID := chi.URLParam(r, "jobID")
	jobID, err := strconv.Atoi(paramJobID)
	if err != nil || jobID < 0 {
		log.Info().Err(err).
			Str("job_id", paramJobID).
			Msg("invalid job id")
		WriteErrorResponse(w, ErrInvalidInput, http.StatusBadRequest)
		return
	}

	// get the consumer id from the header
	consumerID := r.Header.Get(HeaderQueueConsumer)
	if consumerID == "" {
		log.Info().Msg("no valid queue consumer ID")
		WriteErrorResponse(w, ErrInvalidInput, http.StatusBadRequest)
		return
	}

	// conclude the job
	if err := h.JobQueuer.Conclude(ctx, jobID, consumerID); err != nil {
		if errors.Is(err, domain.ErrJobNotFound{}) {
			WriteErrorResponse(w, ErrNotFound, http.StatusNotFound)
			return
		}

		log.Error().Err(err).Msgf("error concluding job")
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, http.StatusNoContent, nil)
}

// GetJobStatus returns the status of a job.
func (h *JobHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("handler", "GetJobStatus").Logger()

	// validate jobID param is a non-negative integer
	paramJobID := chi.URLParam(r, "jobID")
	jobID, err := strconv.Atoi(paramJobID)
	if err != nil || jobID < 0 {
		log.Info().Err(err).
			Str("job_id", paramJobID).
			Msg("invalid job id")
		WriteErrorResponse(w, ErrInvalidInput, http.StatusBadRequest)
		return
	}

	// fetch the job status
	queuedJob, err := h.JobQueuer.FetchJob(ctx, jobID)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound{}) {
			WriteErrorResponse(w, ErrNotFound, http.StatusNotFound)
			return
		}

		log.Printf("internal server error: %s", err) // TODO: use zerolog from request context
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// marshal and return the job in the response
	response, err := json.Marshal(job{
		ID:     queuedJob.ID,
		Status: queuedJob.Status,
		Type:   queuedJob.Type,
	})
	if err != nil {
		log.Error().Err(err).
			Str("job_id", strconv.Itoa(queuedJob.ID)).
			Str("job_type", queuedJob.Type).
			Str("job_status", queuedJob.Status).
			Msgf("error marshalling response")
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// CancelJob cancels a job that is not yet concluded.
func (h *JobHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("handler", "CancelJob").Logger()

	// validate jobID param is a non-negative integer
	paramJobID := chi.URLParam(r, "jobID")
	jobID, err := strconv.Atoi(paramJobID)
	if err != nil || jobID < 0 {
		log.Info().Err(err).
			Str("job_id", paramJobID).
			Msg("invalid job id")
		WriteErrorResponse(w, ErrInvalidInput, http.StatusBadRequest)
		return
	}

	// fetch the job status
	err = h.JobQueuer.CancelJob(ctx, jobID)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound{}) {
			WriteErrorResponse(w, ErrNotFound, http.StatusNotFound)
			return
		}

		log.Printf("internal server error: %s", err) // TODO: use zerolog from request context
		WriteErrorResponse(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, http.StatusOK, nil)
}
