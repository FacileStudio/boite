package worker

import (
	"fmt"
	"slices"
	"sync"
)

type stubLife struct {
	mu      sync.Mutex
	boxes   map[string]*Box
	snaps   map[string][]string
	calls   []string
	gate    chan struct{}
	failure error
}

func newStub() *stubLife {
	return &stubLife{boxes: map[string]*Box{}, snaps: map[string][]string{}}
}

func (s *stubLife) record(call string) {
	s.calls = append(s.calls, call)
}

func (s *stubLife) Reserve(req CreateRequest) (Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("reserve " + req.Name)
	if _, ok := s.boxes[req.Name]; ok {
		return Box{}, ErrExists
	}
	box := &Box{
		ID: req.Name, Name: req.Name, Owner: req.Owner, Status: StatusCreating,
		CPUs: req.CPUs, Memory: req.Memory, Disk: req.Disk, Net: req.Net,
		SSH: SSHInfo{Host: "10.0.0.1", User: sshUser},
	}
	s.boxes[req.Name] = box
	return *box, nil
}

func (s *stubLife) Create(req CreateRequest) (Box, error) {
	s.mu.Lock()
	s.record("create " + req.Name)
	gate, failure := s.gate, s.failure
	s.mu.Unlock()
	if gate != nil {
		<-gate
	}
	if failure != nil {
		return Box{}, failure
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	box, ok := s.boxes[req.Name]
	if !ok {
		return Box{}, ErrNotFound
	}
	box.Status = StatusRunning
	box.SSH.Port = 2223
	return *box, nil
}

func (s *stubLife) Fail(box Box, msg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("fail " + box.ID)
	stored, ok := s.boxes[box.ID]
	if !ok {
		return ErrNotFound
	}
	stored.Status, stored.Error = StatusError, msg
	return nil
}

func (s *stubLife) Recover() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, box := range s.boxes {
		if box.Status == StatusCreating {
			box.Status, box.Error = StatusError, "interrupted"
		}
	}
	return nil
}

func (s *stubLife) setStatus(id, status string) (Box, error) {
	box, ok := s.boxes[id]
	if !ok {
		return Box{}, ErrNotFound
	}
	box.Status = status
	return *box, nil
}

func (s *stubLife) Start(id string) (Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("start " + id)
	return s.setStatus(id, StatusRunning)
}

func (s *stubLife) Stop(id string) (Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("stop " + id)
	return s.setStatus(id, StatusStopped)
}

func (s *stubLife) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("rm " + id)
	if _, ok := s.boxes[id]; !ok {
		return ErrNotFound
	}
	delete(s.boxes, id)
	return nil
}

func (s *stubLife) Get(id string) (Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	box, ok := s.boxes[id]
	if !ok {
		return Box{}, ErrNotFound
	}
	return *box, nil
}

func (s *stubLife) List() ([]Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Box{}
	for _, box := range s.boxes {
		out = append(out, *box)
	}
	return out, nil
}

func (s *stubLife) Snapshot(id, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("snapshot " + id + " " + tag)
	s.snaps[id] = append(s.snaps[id], tag)
	return nil
}

func (s *stubLife) Rollback(id, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("rollback " + id + " " + tag)
	if slices.Contains(s.snaps[id], tag) {
		return nil
	}
	return invalid("no snapshot %q", tag)
}

func (s *stubLife) Snapshots(id string) ([]Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Snapshot{}
	for i, tag := range s.snaps[id] {
		out = append(out, Snapshot{ID: fmt.Sprint(i + 1), Tag: tag})
	}
	return out, nil
}

func (s *stubLife) SetSSHKey(id, key string) (Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("ssh-key " + id)
	box, ok := s.boxes[id]
	if !ok {
		return Box{}, ErrNotFound
	}
	return *box, nil
}

func (s *stubLife) SetNet(id, net string) (Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record("net " + id + " " + net)
	box, ok := s.boxes[id]
	if !ok {
		return Box{}, ErrNotFound
	}
	box.Net = net
	return *box, nil
}

func (s *stubLife) called(call string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Contains(s.calls, call)
}
