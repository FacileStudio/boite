package worker

import "net/http"

func (s *Server) requireStopped(id string) error {
	box, err := s.life.Get(id)
	if err != nil {
		return err
	}
	if box.Status != StatusStopped {
		return ErrNotStopped
	}
	return nil
}

func (s *Server) tagAction(w http.ResponseWriter, r *http.Request, status int, act func(id, tag string) error) {
	var req tagRequest
	if err := decodeBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	id := r.PathValue("id")
	err := s.exclusive(id, func() error {
		if err := s.requireStopped(id); err != nil {
			return err
		}
		return act(id, req.Tag)
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, status, map[string]string{"id": id, "tag": req.Tag})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	s.tagAction(w, r, http.StatusCreated, s.life.Snapshot)
}

func (s *Server) handleRollback(w http.ResponseWriter, r *http.Request) {
	s.tagAction(w, r, http.StatusOK, s.life.Rollback)
}

func (s *Server) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var snaps []Snapshot
	err := s.exclusive(id, func() error {
		if err := s.requireStopped(id); err != nil {
			return err
		}
		var err error
		snaps, err = s.life.Snapshots(id)
		return err
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snaps)
}
