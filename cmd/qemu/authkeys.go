package qemu

import (
	"fmt"
	"os"
	"strings"
)

// ControlPubKeyEnv names the environment variable that holds the control
// public key injected into every sandbox beside the user's own keys.
const ControlPubKeyEnv = "BOITE_CONTROL_PUBKEY"

var publicKeyPrefixes = []string{"ssh-", "ecdsa-", "sk-ssh-", "sk-ecdsa-"}

// ValidatePublicKey rejects anything that is not a single-line OpenSSH public
// key, so a value can never smuggle extra authorized_keys entries or options.
func ValidatePublicKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" || strings.ContainsAny(key, "\r\n\x00") {
		return fmt.Errorf("public key must be a single non-empty line")
	}
	for _, prefix := range publicKeyPrefixes {
		if strings.HasPrefix(key, prefix) && len(strings.Fields(key)) >= 2 {
			return nil
		}
	}
	return fmt.Errorf("public key must start with a key type such as ssh-ed25519")
}

// JoinAuthorizedKeys builds the authorized_keys content: the primary key, any
// extra keys, then the control key from BOITE_CONTROL_PUBKEY. Blank entries
// and duplicates are dropped.
func JoinAuthorizedKeys(primary string, extra []string) string {
	seen := map[string]bool{}
	var lines []string
	candidates := append([]string{primary}, extra...)
	candidates = append(candidates, os.Getenv(ControlPubKeyEnv))
	for _, key := range candidates {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		lines = append(lines, key)
	}
	return strings.Join(lines, "\n")
}

// ApplyAuthorizedKeys rewrites the running guest's authorized_keys over SSH
// with the instance key, the control key and extra. The config disk only
// seeds keys at first boot, so a later key change has to go through the guest.
func ApplyAuthorizedKeys(inst *Instance, extra []string) error {
	primary, err := ReadPublicKey(inst.PubKeyPath)
	if err != nil {
		return fmt.Errorf("read instance public key: %w", err)
	}
	keys := strings.Split(JoinAuthorizedKeys(primary, extra), "\n")
	quoted := make([]string, len(keys))
	for i, key := range keys {
		quoted[i] = shellQuote(key)
	}
	script := "mkdir -p ~/.ssh && printf '%s\\n' " + strings.Join(quoted, " ") +
		" > ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
	if _, err := SSHOutput(inst, script); err != nil {
		return fmt.Errorf("write authorized_keys: %w", err)
	}
	return nil
}
