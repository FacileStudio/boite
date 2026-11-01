package worker

import (
	"os"
	"os/user"
	"strconv"
	"testing"
)

type sockAccessCase struct {
	name, mode, group string
	wantErr           bool
	wantMode          os.FileMode
	wantGID           int
}

func sockAccessCases(groupName string) []sockAccessCase {
	return []sockAccessCase{
		{"defaults", "", "", false, 0o600, noGroup},
		{"group mode", "0660", "", false, 0o660, noGroup},
		{"mode without leading zero", "660", "", false, 0o660, noGroup},
		{"group by name", "0660", groupName, false, 0o660, os.Getgid()},
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
}

func TestLoadConfigSocketAccess(t *testing.T) {
	grp, err := user.LookupGroupId(strconv.Itoa(os.Getgid()))
	if err != nil {
		t.Skip("current group has no name")
	}
	for _, c := range sockAccessCases(grp.Name) {
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
