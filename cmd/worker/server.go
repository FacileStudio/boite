package worker

import (
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Server is the HTTP control API over a Lifecycle.
type Server struct {
	life  Lifecycle
	token [sha256.Size]byte
	mu    sync.Mutex
	busy  map[string]bool
	jobs  map[string]*createJob
	queue chan *createJob
}

// NewServer builds a Server that authorizes requests against token. It marks
// boxes a previous worker left in status creating as failed, then starts the
// queue that runs creates one at a time.
func NewServer(life Lifecycle, token string) *Server {
	s := &Server{
		life: life, token: sha256.Sum256([]byte(token)), busy: map[string]bool{},
		jobs: map[string]*createJob{}, queue: make(chan *createJob, queueSize),
	}
	s.startJobs()
	return s
}

// Handler returns the routed, authenticated, logged API handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /boxes", s.handleCreate)
	mux.HandleFunc("GET /boxes", s.handleList)
	mux.HandleFunc("GET /boxes/{id}", s.handleGet)
	mux.HandleFunc("POST /boxes/{id}/start", s.handleStart)
	mux.HandleFunc("POST /boxes/{id}/stop", s.handleStop)
	mux.HandleFunc("POST /boxes/{id}/rm", s.handleRemove)
	mux.HandleFunc("PUT /boxes/{id}/ssh-key", s.handleSSHKey)
	mux.HandleFunc("POST /boxes/{id}/snapshot", s.handleSnapshot)
	mux.HandleFunc("POST /boxes/{id}/rollback", s.handleRollback)
	mux.HandleFunc("GET /boxes/{id}/snapshots", s.handleSnapshots)
	return s.logged(s.authorized(mux))
}

func (s *Server) authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		got := sha256.Sum256([]byte(bearer))
		if !ok || subtle.ConstantTimeCompare(got[:], s.token[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeJSON(w, http.StatusUnauthorized, errorBody{Error: "missing or invalid bearer token", Code: "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *Server) logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "ms", time.Since(start).Milliseconds())
	})
}

func (s *Server) acquire(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy[id] {
		return false
	}
	s.busy[id] = true
	return true
}

func (s *Server) release(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.busy, id)
}

func (s *Server) exclusive(id string, fn func() error) error {
	if !ValidID(id) {
		return invalid("box id %q is not a valid name", id)
	}
	if !s.acquire(id) {
		return ErrBusy
	}
	defer s.release(id)
	return fn()
}
