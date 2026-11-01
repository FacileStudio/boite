package worker

import (
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
)

const queueSize = 256

type createJob struct {
	req       CreateRequest
	box       Box
	cancelled atomic.Bool
	done      chan struct{}
}

func (s *Server) startJobs() {
	if err := s.life.Recover(); err != nil {
		slog.Error("recover interrupted creates", "err", err)
	}
	go func() {
		for job := range s.queue {
			s.execute(job)
		}
	}()
}

func (s *Server) enqueue(box Box, req CreateRequest) error {
	job := &createJob{req: req, box: box, done: make(chan struct{})}
	s.mu.Lock()
	s.jobs[req.Name] = job
	s.mu.Unlock()
	select {
	case s.queue <- job:
		return nil
	default:
		s.finish(job)
		if err := s.life.Remove(req.Name); err != nil {
			slog.Warn("drop rejected create", "id", req.Name, "err", err)
		}
		return ErrBusy
	}
}

func (s *Server) execute(job *createJob) {
	defer s.finish(job)
	if job.cancelled.Load() {
		s.drop(job)
		return
	}
	_, err := s.life.Create(job.req)
	switch {
	case job.cancelled.Load():
		s.drop(job)
	case err != nil:
		slog.Error("create box", "id", job.req.Name, "err", err)
		if ferr := s.life.Fail(job.box, err.Error()); ferr != nil {
			slog.Error("record failed create", "id", job.req.Name, "err", ferr)
		}
	}
}

func (s *Server) drop(job *createJob) {
	if err := s.life.Remove(job.req.Name); err != nil && !errors.Is(err, ErrNotFound) {
		slog.Error("remove cancelled create", "id", job.req.Name, "err", err)
	}
}

func (s *Server) finish(job *createJob) {
	s.mu.Lock()
	delete(s.jobs, job.req.Name)
	s.mu.Unlock()
	close(job.done)
}

func (s *Server) pendingJob(id string) *createJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jobs[id]
}

func (s *Server) cancelCreate(w http.ResponseWriter, r *http.Request, job *createJob) {
	job.cancelled.Store(true)
	select {
	case <-job.done:
		w.WriteHeader(http.StatusNoContent)
	case <-r.Context().Done():
	}
}
