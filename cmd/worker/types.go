package worker

import "errors"

var (
	ErrNotFound   = errors.New("box not found")
	ErrExists     = errors.New("box already exists")
	ErrInvalid    = errors.New("invalid request")
	ErrNotStopped = errors.New("box is running: stop it first")
	ErrBusy       = errors.New("box is busy with another operation")
)

const sshUser = "boite"

const (
	StatusRunning = "running"
	StatusStopped = "stopped"
)

// CreateRequest is the body of POST /boxes. Zero resources take defaults.
type CreateRequest struct {
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	CPUs   int    `json:"cpus"`
	Memory string `json:"memory"`
	Disk   string `json:"disk"`
	SSHKey string `json:"ssh_key"`
}

// SSHInfo tells a client where and as whom to open an SSH session.
type SSHInfo struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
}

// Box is the API view of one sandbox. The ID is the sandbox name.
type Box struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Owner     string  `json:"owner"`
	Status    string  `json:"status"`
	CPUs      int     `json:"cpus"`
	Memory    string  `json:"memory"`
	Disk      string  `json:"disk"`
	SSH       SSHInfo `json:"ssh"`
	CreatedAt string  `json:"created_at"`
}

// Snapshot is one internal qcow2 snapshot on a box's overlay disk.
type Snapshot struct {
	ID        string `json:"id"`
	Tag       string `json:"tag"`
	CreatedAt string `json:"created_at"`
}

type tagRequest struct {
	Tag string `json:"tag"`
}

type keyRequest struct {
	SSHKey string `json:"ssh_key"`
}

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
