package worker

import (
	"regexp"
	"strings"

	"github.com/FacileStudio/boite/cmd/qemu"
)

var snapshotDate = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)

func (b *Backend) checkSnapshot(id, tag string) error {
	if !qemu.InstanceExists(id) {
		return ErrNotFound
	}
	if err := qemu.ValidSnapshotTag(tag); err != nil {
		return invalid("%v", err)
	}
	return nil
}

// Snapshot records tag on a stopped box's overlay disk.
func (b *Backend) Snapshot(id, tag string) error {
	if err := b.checkSnapshot(id, tag); err != nil {
		return err
	}
	return qemu.SnapshotCreate(id, tag)
}

// Rollback restores a stopped box's overlay disk to tag.
func (b *Backend) Rollback(id, tag string) error {
	if err := b.checkSnapshot(id, tag); err != nil {
		return err
	}
	return qemu.SnapshotRollback(id, tag)
}

// Snapshots lists the snapshots on a stopped box's overlay disk.
func (b *Backend) Snapshots(id string) ([]Snapshot, error) {
	if !qemu.InstanceExists(id) {
		return nil, ErrNotFound
	}
	table, err := qemu.SnapshotList(id)
	if err != nil {
		return nil, err
	}
	return ParseSnapshotTable(table), nil
}

// ParseSnapshotTable reads the output of 'qemu-img snapshot -l' into
// snapshots. Header and blank lines are skipped.
func ParseSnapshotTable(table string) []Snapshot {
	snaps := []Snapshot{}
	for line := range strings.SplitSeq(table, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] == "ID" || fields[0] == "Snapshot" {
			continue
		}
		snaps = append(snaps, Snapshot{ID: fields[0], Tag: fields[1], CreatedAt: snapshotDate.FindString(line)})
	}
	return snaps
}
