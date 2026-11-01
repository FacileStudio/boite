package worker

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"strconv"
	"strings"
	"testing"
)

const testToken = "s3cret-token"

const validKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItest user@host"

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func seeded(t *testing.T, status string) (*stubLife, http.Handler) {
	t.Helper()
	stub := newStub()
	stub.boxes["dev"] = &Box{ID: "dev", Name: "dev", Owner: "ann", Status: status}
	return stub, NewServer(stub, testToken).Handler()
}

func TestAuth(t *testing.T) {
	cases := []struct {
		name, header string
		want         int
	}{
		{"none", "", http.StatusUnauthorized},
		{"wrong", "Bearer nope", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + testToken, http.StatusUnauthorized},
		{"empty bearer", "Bearer ", http.StatusUnauthorized},
		{"valid", "Bearer " + testToken, http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, h := seeded(t, StatusStopped)
			req := httptest.NewRequest("GET", "/boxes", nil)
			if c.header != "" {
				req.Header.Set("Authorization", c.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Fatalf("got %d, want %d", rec.Code, c.want)
			}
		})
	}
}

type routeCase struct {
	name, method, path, body, status string
	want                             int
	code                             string
}

func routeCases() []routeCase {
	return []routeCase{
		{"create", "POST", "/boxes", `{"name":"new","owner":"bob"}`, StatusStopped, 201, ""},
		{"create exists", "POST", "/boxes", `{"name":"dev","owner":"bob"}`, StatusStopped, 409, "exists"},
		{"create bad name", "POST", "/boxes", `{"name":"../x","owner":"bob"}`, StatusStopped, 400, "invalid"},
		{"create no owner", "POST", "/boxes", `{"name":"new"}`, StatusStopped, 400, "invalid"},
		{"create bad cpus", "POST", "/boxes", `{"name":"new","owner":"b","cpus":99}`, StatusStopped, 400, "invalid"},
		{"create bad memory", "POST", "/boxes", `{"name":"new","owner":"b","memory":"lots"}`, StatusStopped, 400, "invalid"},
		{"create bad key", "POST", "/boxes", `{"name":"new","owner":"b","ssh_key":"nope"}`, StatusStopped, 400, "invalid"},
		{"create unknown field", "POST", "/boxes", `{"name":"new","owner":"b","x":1}`, StatusStopped, 400, "invalid"},
		{"get", "GET", "/boxes/dev", "", StatusStopped, 200, ""},
		{"get missing", "GET", "/boxes/ghost", "", StatusStopped, 404, "not_found"},
		{"start", "POST", "/boxes/dev/start", "", StatusStopped, 200, ""},
		{"stop", "POST", "/boxes/dev/stop", "", StatusRunning, 200, ""},
		{"stop missing", "POST", "/boxes/ghost/stop", "", StatusRunning, 404, "not_found"},
		{"rm", "POST", "/boxes/dev/rm", "", StatusRunning, 204, ""},
		{"ssh key", "PUT", "/boxes/dev/ssh-key", `{"ssh_key":"` + validKey + `"}`, StatusRunning, 200, ""},
		{"snapshot stopped", "POST", "/boxes/dev/snapshot", `{"tag":"a"}`, StatusStopped, 201, ""},
		{"snapshot running", "POST", "/boxes/dev/snapshot", `{"tag":"a"}`, StatusRunning, 409, "not_stopped"},
		{"rollback running", "POST", "/boxes/dev/rollback", `{"tag":"a"}`, StatusRunning, 409, "not_stopped"},
		{"rollback unknown tag", "POST", "/boxes/dev/rollback", `{"tag":"a"}`, StatusStopped, 400, "invalid"},
		{"snapshots running", "GET", "/boxes/dev/snapshots", "", StatusRunning, 409, "not_stopped"},
		{"snapshots stopped", "GET", "/boxes/dev/snapshots", "", StatusStopped, 200, ""},
		{"snapshot missing box", "POST", "/boxes/ghost/snapshot", `{"tag":"a"}`, StatusStopped, 404, "not_found"},
	}
}

func TestRoutes(t *testing.T) {
	for _, c := range routeCases() {
		t.Run(c.name, func(t *testing.T) {
			_, h := seeded(t, c.status)
			rec := do(h, c.method, c.path, c.body)
			if rec.Code != c.want {
				t.Fatalf("got %d, want %d: %s", rec.Code, c.want, rec.Body)
			}
			if c.code != "" && !strings.Contains(rec.Body.String(), `"code":"`+c.code+`"`) {
				t.Fatalf("want code %q in %s", c.code, rec.Body)
			}
		})
	}
}

func TestSnapshotNotCalledWhenRunning(t *testing.T) {
	stub, h := seeded(t, StatusRunning)
	do(h, "POST", "/boxes/dev/snapshot", `{"tag":"a"}`)
	if stub.called("snapshot dev a") {
		t.Fatal("backend snapshot ran against a running box")
	}
}

func TestListFiltersByOwner(t *testing.T) {
	_, h := seeded(t, StatusStopped)
	rec := do(h, "GET", "/boxes?owner=zed", "")
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("want empty list, got %s", rec.Body)
	}
}

func TestBusyBoxIsRefused(t *testing.T) {
	srv := NewServer(newStub(), testToken)
	if !srv.acquire("dev") || srv.acquire("dev") {
		t.Fatal("second acquire of the same box must fail")
	}
	srv.release("dev")
	if !srv.acquire("dev") {
		t.Fatal("release must free the box")
	}
}

func TestNormalizeDefaults(t *testing.T) {
	req, err := CreateRequest{Name: "a", Owner: " ann "}.Normalize()
	if err != nil {
		t.Fatal(err)
	}
	if req.CPUs != 2 || req.Memory != "2G" || req.Disk != "20G" || req.Owner != "ann" {
		t.Fatalf("defaults not applied: %+v", req)
	}
	req, err = CreateRequest{Name: "a", Owner: "o", Memory: "512m", Disk: "40"}.Normalize()
	if err != nil || req.Memory != "512M" || req.Disk != "40G" {
		t.Fatalf("normalize: %+v %v", req, err)
	}
}

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		wantErr bool
		sock    string
		host    string
	}{
		{"no token", map[string]string{}, true, "", ""},
		{"defaults", map[string]string{"BOITE_WORKER_TOKEN": "t"}, false, "/run/boite/worker.sock", "127.0.0.1"},
		{"bind becomes host", map[string]string{"BOITE_WORKER_TOKEN": "t", "BOITE_SSH_BIND": "172.18.0.1"}, false, "/run/boite/worker.sock", "172.18.0.1"},
		{"host override", map[string]string{"BOITE_WORKER_TOKEN": "t", "BOITE_SSH_BIND": "0.0.0.0", "BOITE_WORKER_HOST": "box.example"}, false, "/run/boite/worker.sock", "box.example"},
		{"bad bind", map[string]string{"BOITE_WORKER_TOKEN": "t", "BOITE_SSH_BIND": "lan"}, true, "", ""},
		{"custom sock", map[string]string{"BOITE_WORKER_TOKEN": "t", "BOITE_WORKER_SOCK": "/tmp/w.sock"}, false, "/tmp/w.sock", "127.0.0.1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, err := LoadConfig(func(k string) string { return c.env[k] })
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr && (cfg.Sock != c.sock || cfg.Host != c.host) {
				t.Fatalf("got %+v", cfg)
			}
		})
	}
}

func TestLoadConfigSocketAccess(t *testing.T) {
	grp, err := user.LookupGroupId(strconv.Itoa(os.Getgid()))
	if err != nil {
		t.Skip("current group has no name")
	}
	cases := []struct {
		name, mode, group string
		wantErr           bool
		wantMode          os.FileMode
		wantGID           int
	}{
		{"defaults", "", "", false, 0o600, noGroup},
		{"group mode", "0660", "", false, 0o660, noGroup},
		{"mode without leading zero", "660", "", false, 0o660, noGroup},
		{"group by name", "0660", grp.Name, false, 0o660, os.Getgid()},
		{"group by gid", "0660", "65532", false, 0o660, 65532},
		{"not octal", "0689", "", true, 0, 0},
		{"not a number", "rw", "", true, 0, 0},
		{"too wide", "01777", "", true, 0, 0},
		{"owner cannot write", "0460", "", true, 0, 0},
		{"world access", "0666", "", true, 0, 0},
		{"unknown group", "0660", "no-such-group-boite", true, 0, 0},
		{"negative gid", "0660", "-5", true, 0, 0},
		{"group with default mode", "", "65532", true, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			env := map[string]string{
				"BOITE_WORKER_TOKEN": "t", "BOITE_WORKER_SOCK_MODE": c.mode, "BOITE_WORKER_SOCK_GROUP": c.group,
			}
			cfg, err := LoadConfig(func(k string) string { return env[k] })
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr && (cfg.SockMode != c.wantMode || cfg.SockGID != c.wantGID) {
				t.Fatalf("got mode %o gid %d", cfg.SockMode, cfg.SockGID)
			}
		})
	}
}

func TestParseSnapshotTable(t *testing.T) {
	table := "Snapshot list:\nID        TAG               VM SIZE                DATE     VM CLOCK     ICOUNT\n" +
		"1         before            0 B  2026-11-01 10:00:00  00:00:00.000         0\n" +
		"2         after             1.5 MiB  2026-11-01 11:30:05  00:00:00.000         0\n"
	snaps := ParseSnapshotTable(table)
	if len(snaps) != 2 || snaps[0].Tag != "before" || snaps[1].CreatedAt != "2026-11-01 11:30:05" {
		t.Fatalf("got %+v", snaps)
	}
	if len(ParseSnapshotTable("")) != 0 {
		t.Fatal("empty table must yield no snapshots")
	}
}
