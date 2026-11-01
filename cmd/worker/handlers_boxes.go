package worker

import (
	"net/http"
	"strings"
)

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := decodeBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	req, err := req.Normalize()
	if err != nil {
		writeError(w, err)
		return
	}
	var box Box
	err = s.exclusive(req.Name, func() (err error) {
		box, err = s.life.Create(req)
		return err
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, box)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	boxes, err := s.life.List()
	if err != nil {
		writeError(w, err)
		return
	}
	owner := r.URL.Query().Get("owner")
	out := make([]Box, 0, len(boxes))
	for _, box := range boxes {
		if owner == "" || box.Owner == owner {
			out = append(out, box)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !ValidID(id) {
		writeError(w, ErrNotFound)
		return
	}
	box, err := s.life.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, box)
}

func (s *Server) boxAction(w http.ResponseWriter, r *http.Request, act func(id string) (Box, error)) {
	var box Box
	err := s.exclusive(r.PathValue("id"), func() (err error) {
		box, err = act(r.PathValue("id"))
		return err
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, box)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.boxAction(w, r, s.life.Start)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.boxAction(w, r, s.life.Stop)
}

func (s *Server) handleRemove(w http.ResponseWriter, r *http.Request) {
	err := s.exclusive(r.PathValue("id"), func() error {
		return s.life.Remove(r.PathValue("id"))
	})
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSSHKey(w http.ResponseWriter, r *http.Request) {
	var req keyRequest
	if err := decodeBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	key := strings.TrimSpace(req.SSHKey)
	s.boxAction(w, r, func(id string) (Box, error) {
		return s.life.SetSSHKey(id, key)
	})
}
