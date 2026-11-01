package worker

// Provisioner creates, starts, stops and removes boxes.
type Provisioner interface {
	Create(req CreateRequest) (Box, error)
	Start(id string) (Box, error)
	Stop(id string) (Box, error)
	Remove(id string) error
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

// KeyStore replaces the user SSH key of a box.
type KeyStore interface {
	SetSSHKey(id, key string) (Box, error)
}

// Lifecycle is everything the control API needs from the QEMU layer. The
// production implementation is Backend; tests substitute a stub.
type Lifecycle interface {
	Provisioner
	Inventory
	Snapshotter
	KeyStore
}
