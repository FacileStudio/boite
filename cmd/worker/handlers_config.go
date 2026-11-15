package worker

import (
	"net/http"
	"strings"
)

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

func (s *Server) handleNet(w http.ResponseWriter, r *http.Request) {
	var req netRequest
	if err := decodeBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	net, allow, err := NormalizeNetConfig(req.Net, req.NetAllow)
	if req.Net == "" {
		err = invalid("net is required")
	}
	if err != nil {
		writeError(w, err)
		return
	}
	s.boxAction(w, r, func(id string) (Box, error) {
		return s.life.SetNet(id, net, allow)
	})
}
