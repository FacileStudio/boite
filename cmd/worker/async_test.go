package worker

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestCreateTransitionsToRunning(t *testing.T) {
	stub := newStub()
	stub.gate = make(chan struct{})
	client, _ := startSocketServer(t, stub)
	body := client.expect(http.StatusAccepted, "POST", "/boxes", `{"name":"dev","owner":"ann"}`, testToken)
	if !strings.Contains(body, `"status":"creating"`) {
		t.Fatalf("create body: %s", body)
	}
	if got := client.expect(http.StatusOK, "GET", "/boxes", "", testToken); !strings.Contains(got, `"creating"`) {
		t.Fatalf("list must show the creating box: %s", got)
	}
	client.expect(http.StatusConflict, "POST", "/boxes", `{"name":"dev","owner":"ann"}`, testToken)
	close(stub.gate)
	client.waitStatus("dev", StatusRunning)
}

func TestCreateFailureKeepsErrorBox(t *testing.T) {
	stub := newStub()
	stub.failure = errors.New("qemu exploded")
	client, _ := startSocketServer(t, stub)
	client.expect(http.StatusAccepted, "POST", "/boxes", `{"name":"dev","owner":"ann"}`, testToken)
	box := client.waitStatus("dev", StatusError)
	if box.Error != "qemu exploded" {
		t.Fatalf("error = %q", box.Error)
	}
	client.expect(http.StatusConflict, "POST", "/boxes", `{"name":"dev","owner":"ann"}`, testToken)
	client.expect(http.StatusNoContent, "POST", "/boxes/dev/rm", "", testToken)
	client.expect(http.StatusNotFound, "GET", "/boxes/dev", "", testToken)
}

func TestRemoveWhileQueuedCancels(t *testing.T) {
	stub := newStub()
	stub.gate = make(chan struct{})
	client, _ := startSocketServer(t, stub)
	client.expect(http.StatusAccepted, "POST", "/boxes", `{"name":"first","owner":"ann"}`, testToken)
	client.expect(http.StatusAccepted, "POST", "/boxes", `{"name":"second","owner":"ann"}`, testToken)
	done := make(chan struct{})
	go func() {
		client.expect(http.StatusNoContent, "POST", "/boxes/second/rm", "", testToken)
		close(done)
	}()
	close(stub.gate)
	<-done
	client.waitStatus("first", StatusRunning)
	client.expect(http.StatusNotFound, "GET", "/boxes/second", "", testToken)
}

func TestRemoveWhileCreatingCleansUp(t *testing.T) {
	stub := newStub()
	stub.gate = make(chan struct{})
	client, _ := startSocketServer(t, stub)
	client.expect(http.StatusAccepted, "POST", "/boxes", `{"name":"dev","owner":"ann"}`, testToken)
	done := make(chan struct{})
	go func() {
		client.expect(http.StatusNoContent, "POST", "/boxes/dev/rm", "", testToken)
		close(done)
	}()
	close(stub.gate)
	<-done
	client.expect(http.StatusNotFound, "GET", "/boxes/dev", "", testToken)
}

func TestRestartMarksCreatingAsInterrupted(t *testing.T) {
	stub := newStub()
	stub.boxes["stuck"] = &Box{ID: "stuck", Name: "stuck", Owner: "ann", Status: StatusCreating}
	stub.boxes["up"] = &Box{ID: "up", Name: "up", Owner: "ann", Status: StatusRunning}
	h := NewServer(stub, testToken).Handler()
	rec := do(h, "GET", "/boxes/stuck", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"status":"error"`) || !strings.Contains(rec.Body.String(), `"error":"interrupted"`) {
		t.Fatalf("stuck box: %d %s", rec.Code, rec.Body)
	}
	if rec := do(h, "GET", "/boxes/up", ""); !strings.Contains(rec.Body.String(), `"status":"running"`) {
		t.Fatalf("running box must be untouched: %s", rec.Body)
	}
}
