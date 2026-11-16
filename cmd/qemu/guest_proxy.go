package qemu

import (
	"fmt"
	"strings"
)

const (
	guestProxyBegin = "# boite-egress begin"
	guestProxyEnd   = "# boite-egress end"
)

// GuestProxyScript returns the shell script that installs (enable) or removes
// the proxy settings in a Debian-style guest, in one sudo call because sudo
// stalls on a hostname lookup when the guest has no DNS (RES_OPTIONS keeps that
// to a few seconds, and the script adds the hostname to /etc/hosts): /etc/environment for login and
// ssh sessions, /etc/profile.d for login shells, and an apt drop-in. The
// settings are a convenience for tools. The enforcement is the host proxy and
// slirp restrict=on, which a guest root cannot change.
func GuestProxyScript(enable bool) string {
	cleanup := "sed -i '/^" + guestProxyBegin + "$/,/^" + guestProxyEnd + "$/d' /etc/environment"
	if !enable {
		return rootScript(cleanup + "; rm -f /etc/profile.d/boite-egress.sh /etc/apt/apt.conf.d/99boite-egress")
	}
	url := fmt.Sprintf("http://%s:%d", EgressProxyHost, EgressProxyPort)
	var env, prof []string
	for _, v := range []string{"http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY"} {
		env = append(env, fmt.Sprintf("%s=%s", v, url))
		prof = append(prof, fmt.Sprintf("export %s=%s", v, url))
	}
	for _, v := range []string{"no_proxy", "NO_PROXY"} {
		env = append(env, fmt.Sprintf("%s=localhost,127.0.0.1,::1", v))
		prof = append(prof, fmt.Sprintf("export %s=localhost,127.0.0.1,::1", v))
	}
	apt := []string{
		fmt.Sprintf("Acquire::http::Proxy \"%s/\";", url),
		fmt.Sprintf("Acquire::https::Proxy \"%s/\";", url),
	}
	return rootScript(strings.Join([]string{
		cleanup,
		"grep -q \" $(hostname)$\" /etc/hosts || echo \"127.0.1.1 $(hostname)\" >> /etc/hosts",
		"printf '%s\\n' " + joinQuoted(append(append([]string{guestProxyBegin}, env...), guestProxyEnd)) + " >> /etc/environment",
		"printf '%s\\n' " + joinQuoted(prof) + " > /etc/profile.d/boite-egress.sh",
		"mkdir -p /etc/apt/apt.conf.d",
		"printf '%s\\n' " + joinQuoted(apt) + " > /etc/apt/apt.conf.d/99boite-egress",
	}, " && "))
}

func rootScript(script string) string {
	return "RES_OPTIONS='timeout:1 attempts:1' sudo -n sh -c " + shellQuote(script)
}

func joinQuoted(lines []string) string {
	q := make([]string, len(lines))
	for i, l := range lines {
		q[i] = shellQuote(l)
	}
	return strings.Join(q, " ")
}

// ApplyGuestProxy writes or removes the guest proxy settings over SSH. The
// config disk only seeds a guest at first boot, so this runs after each start.
func ApplyGuestProxy(inst *Instance, enable bool) error {
	if _, err := SSHOutput(inst, GuestProxyScript(enable)); err != nil {
		return fmt.Errorf("write guest proxy settings: %w", err)
	}
	return nil
}
