package qemu

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	modeSetuid        = 0o4000
	modeSetgid        = 0o2000
	modeWorldWritable = 0o0002
)

// syncOutDenyList holds the guest-authored paths that never reach the host
// tree on sync-out: executable hooks and env or task files a compromised
// guest could use to run code on the host. Entries are anchored to the
// workspace root: "dir/*" denies everything inside dir at any depth but keeps
// the directory entry itself; other entries are shell globs where '*' stops
// at path separators. Alongside the paths, regular files carrying setuid,
// setgid or world-writable modes are denied too. sync-in is never filtered,
// and 'boite sync --all' disables the whole guard.
var syncOutDenyList = []string{
	".git/hooks/*",
	".envrc",
	".env",
	".env.*",
	".vscode/tasks.json",
	".husky/*",
}

// syncOutPump streams the ssh producer's archive into the untar consumer's
// stdin, through the deny-list filter unless allowAll disables it. It runs
// while both processes execute, so a full pipe never blocks the producer.
func syncOutPump(consumerIn io.WriteCloser, producerOut io.Reader, allowAll bool) error {
	if allowAll {
		return copyArchive(consumerIn, producerOut)
	}
	return filterSyncOut(consumerIn, producerOut, os.Stderr)
}

// filterSyncOut copies the guest tar archive to consumerIn, dropping entries
// the sync-out deny list rejects and re-emitting everything else unchanged,
// so the host tar keeps doing the actual extraction. The consumer's stdin is
// closed when the archive ends or the first error surfaces, so untar always
// sees a terminated stream.
func filterSyncOut(consumerIn io.WriteCloser, producerOut io.Reader, warn io.Writer) error {
	tr := tar.NewReader(producerOut)
	tw := tar.NewWriter(consumerIn)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return closeSyncOut(consumerIn, err)
		}
		if reason := syncOutDenyReason(hdr); reason != "" {
			fmt.Fprintf(warn, "warning: sync-out skipped %s (%s)\n", cleanSyncOutName(hdr.Name), reason)
			continue
		}
		if err := emitSyncOutEntry(tw, tr, hdr); err != nil {
			return closeSyncOut(consumerIn, err)
		}
	}
	if err := tw.Close(); err != nil {
		consumerIn.Close()
		return err
	}
	return consumerIn.Close()
}

// emitSyncOutEntry re-emits one allowed tar entry, header then body, so the
// host tar extracts it exactly as the guest produced it.
func emitSyncOutEntry(tw *tar.Writer, tr *tar.Reader, hdr *tar.Header) error {
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := io.Copy(tw, tr)
	return err
}

// closeSyncOut closes the consumer's stdin and returns the error the caller
// was already handling.
func closeSyncOut(consumerIn io.WriteCloser, err error) error {
	consumerIn.Close()
	return err
}

// syncOutDenyReason reports the first reason a tar entry must not reach the
// host tree, or "" when the entry passes. The mode check covers regular
// files only: symlinks always carry world-writable modes and directories
// transfer no executable content.
func syncOutDenyReason(hdr *tar.Header) string {
	name := cleanSyncOutName(hdr.Name)
	for _, pattern := range syncOutDenyList {
		if matchSyncOutDeny(pattern, name) {
			return fmt.Sprintf("matches deny list entry %s", pattern)
		}
	}
	if hdr.Typeflag != tar.TypeReg {
		return ""
	}
	switch {
	case hdr.Mode&modeSetuid != 0:
		return "setuid bit set"
	case hdr.Mode&modeSetgid != 0:
		return "setgid bit set"
	case hdr.Mode&modeWorldWritable != 0:
		return "world-writable"
	}
	return ""
}

// matchSyncOutDeny reports whether the cleaned relative name is denied by one
// deny-list pattern: "dir/*" covers every entry under dir, at any depth, but
// never the directory entry itself; other patterns are anchored globs where
// '*' stops at path separators.
func matchSyncOutDeny(pattern, name string) bool {
	if dir, ok := strings.CutSuffix(pattern, "/*"); ok {
		return strings.HasPrefix(name, dir+"/")
	}
	matched, err := filepath.Match(pattern, name)
	return err == nil && matched
}

// cleanSyncOutName strips the "./" root prefix and the trailing slash of
// directory entries, so deny patterns match plain relative paths.
func cleanSyncOutName(name string) string {
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimPrefix(name, "/")
	return strings.TrimSuffix(name, "/")
}
