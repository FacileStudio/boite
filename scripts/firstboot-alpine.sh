#!/sbin/openrc-run

name="boite-firstboot"
description="Boite firstboot provisioning"

depend() {
	before sshd
	need dev localmount
}

start() {
	marker="/var/lib/boite/firstboot.done"
	if [ -f "$marker" ]; then
		return 0
	fi

	mkdir -p /mnt/boitecfg
	blk=""
	label_dev="$(findfs LABEL=BOITECFG 2>/dev/null || true)"
	for attempt in 1 2 3 4 5 6; do
		for cand in "$label_dev" /dev/disk/by-label/BOITECFG /dev/vdb /dev/vdc /dev/sdb /dev/sdc /dev/sr0 /dev/sr1; do
			if [ -n "$cand" ] && [ -b "$cand" ] && mount -o ro "$cand" /mnt/boitecfg 2>/dev/null && [ -f /mnt/boitecfg/authorized_keys ]; then
				blk="$cand"
				break
			fi
			umount /mnt/boitecfg 2>/dev/null || true
		done
		if [ -n "$blk" ]; then
			break
		fi
		sleep 2
	done

	if [ -z "$blk" ] || [ ! -f /mnt/boitecfg/authorized_keys ]; then
		return 1
	fi

	mkdir -p /home/boite/.ssh /var/lib/boite
	cp /mnt/boitecfg/authorized_keys /home/boite/.ssh/authorized_keys
	chown boite:boite /home/boite/.ssh/authorized_keys
	chmod 0600 /home/boite/.ssh/authorized_keys

	for f in /mnt/boitecfg/.tiroir*; do
		if [ -f "$f" ]; then
			fname="$(basename "$f")"
			cp "$f" "/home/boite/$fname"
			chown boite:boite "/home/boite/$fname"
			chmod 0600 "/home/boite/$fname"
		fi
	done

	umount /mnt/boitecfg
	touch "$marker"
	return 0
}

if [ -z "${RC_VERSION:-}" ]; then
	start "$@"
fi
