package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"

	"github.com/bkrebsbach/simple-job-queue/internal/handler"
	"github.com/bkrebsbach/simple-job-queue/internal/queue"
)

func main() {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "simple-job-queue").
		Logger()

	var router = chi.NewRouter()

	// define router middlewares
	router.Use(middleware.RequestID)
	router.Use(hlog.NewHandler(log))
	router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	router.Use(hlog.RemoteAddrHandler("ip"))
	router.Use(hlog.UserAgentHandler("user_agent"))
	router.Use(hlog.RefererHandler("referer"))
	router.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

	// setup queue
	inMemoryQueue := queue.NewInMemoryQueue()

	// setup HTTP job handler
	jobHandler := &handler.JobHandler{JobQueuer: inMemoryQueue}

	// define routes
	router.Route("/jobs", func(router chi.Router) {
		router.Post("/enqueue", jobHandler.EnqueueJob)
		router.Post("/dequeue", jobHandler.DequeueJob)
		router.Post("/{jobID}/conclude", jobHandler.ConcludeJob)
		router.Get("/{jobID}", jobHandler.GetJobStatus)
	})

	var stop = make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	var s = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// start server
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatal().Err(err)
		}
	}()

	<-stop

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = s.Shutdown(ctx)
}
