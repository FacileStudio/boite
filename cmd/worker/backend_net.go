package worker

// SetNet records the box's egress mode and extra allow-list hosts in worker.json and its boite.yml. The
// QEMU process reads the mode at boot, so a running box keeps its current
// mode until it is stopped and started again.
func (b *Backend) SetNet(id, net string, allow []string) (Box, error) {
	inst, meta, err := b.load(id)
	if err != nil {
		return Box{}, err
	}
	meta.Net, meta.NetAllow = net, allow
	if err := writeBoxConfig(id, meta); err != nil {
		return Box{}, err
	}
	if err := saveMeta(id, meta); err != nil {
		return Box{}, err
	}
	return b.box(inst, meta), nil
}
