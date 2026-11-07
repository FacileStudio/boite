package worker

// Provisioner creates, starts, stops and removes boxes.
type Provisioner interface {
	Create(req CreateRequest) (Box, error)
	Start(id string) (Box, error)
	Stop(id string) (Box, error)
	Remove(id string) error
}

// Registry persists boxes that are being created or failed to create, so
// they stay listable while the slow Create runs and after it fails.
type Registry interface {
	Reserve(req CreateRequest) (Box, error)
	Fail(box Box, msg string) error
	Recover() error
}

// Inventory reads boxes back with their live status.
type Inventory interface {
	Get(id string) (Box, error)
	List() ([]Box, error)
}

// Snapshotter manages qcow2 snapshots on stopped boxes.
type Snapshotter interface {
	Snapshot(id, tag string) error
	Rollback(id, tag string) error
	Snapshots(id string) ([]Snapshot, error)
}

// Configurer replaces a box's user SSH key or its egress mode.
type Configurer interface {
	SetSSHKey(id, key string) (Box, error)
	SetNet(id, net string) (Box, error)
}

// Maintainer manages snapshots and per-box settings.
type Maintainer interface {
	Snapshotter
	Configurer
}

// Lifecycle is everything the control API needs from the QEMU layer. The
// production implementation is Backend; tests substitute a stub.
type Lifecycle interface {
	Provisioner
	Registry
	Inventory
	Maintainer
}
