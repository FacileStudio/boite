package qemu

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

// EnsureBaseImage returns the path to the Debian base image, downloading and
// checksum-verifying it into the cache when it is not already there.
func EnsureBaseImage() (string, error) {
	cacheDir := GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	basePath := GetBaseImagePath()
	if fi, err := os.Stat(basePath); err == nil && fi.Size() > 0 {
		return basePath, nil
	}

	return downloadBaseImage(basePath)
}

func downloadBaseImage(destPath string) (string, error) {
	fmt.Fprintf(os.Stderr, "Downloading Debian base image...\n")
	resp, err := http.Get(BaseImageURL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Downloaded %.1f MB\n", float64(written)/(1024*1024))

	if err := verifyChecksum(destPath); err != nil {
		os.Remove(destPath)
		return "", err
	}

	return destPath, nil
}

func verifyChecksum(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	sum := fmt.Sprintf("%x", h.Sum(nil))
	if sum != BaseImageSHA256 {
		return fmt.Errorf("checksum mismatch: got %s, expected %s", sum, BaseImageSHA256)
	}
	return nil
}

// CreateOverlay creates a qcow2 overlay disk backed by baseImage at
// overlayPath, sized to sizeGB when it is positive and left to qemu-img's
// default otherwise.
func CreateOverlay(baseImage, overlayPath string, sizeGB int) error {
	args := []string{"create", "-f", "qcow2", "-b", baseImage, "-F", "qcow2", overlayPath}
	if sizeGB > 0 {
		args = append(args, fmt.Sprintf("%dG", sizeGB))
	}
	cmd := exec.Command("qemu-img", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("qemu-img failed: %s (%w)", string(output), err)
	}
	return nil
}
