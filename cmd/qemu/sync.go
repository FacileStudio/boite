package qemu

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const workspaceDir = "/workspace"

// SyncWorkspaceIn replaces the sandbox's /workspace tree with a snapshot of
// srcDir, streamed as a tar archive over the instance's SSH transport. The
// archive extracts into a staging directory on the guest and is swapped in
// only after extraction succeeds, so a failed or dropped stream leaves the
// previous workspace intact. The host tree never leaks paths or permissions
// across the boundary: the guest only ever touches its own staging and
// workspace directories.
func SyncWorkspaceIn(inst *Instance, srcDir string) error {
	tarCmd := exec.Command("tar", "-C", srcDir, "-cf", "-", ".")
	tarCmd.Stderr = os.Stderr
	remote := fmt.Sprintf(
		"sudo sed -i '/[$][(]hostname[)]/d' /etc/hosts; grep -qs \" $(hostname)\" /etc/hosts || echo \"127.0.0.1 $(hostname)\" | sudo tee -a /etc/hosts > /dev/null; "+
			"sudo rm -rf %[1]s.tmp %[1]s.old && sudo mkdir -p %[1]s.tmp && sudo chown $(id -u):$(id -g) %[1]s.tmp && tar -xf - -C %[1]s.tmp && read confirm && [ \"$confirm\" = OK ] && sudo mv %[1]s %[1]s.old && sudo mv %[1]s.tmp %[1]s && sudo rm -rf %[1]s.old",
		workspaceDir)
	sshCmd := exec.Command("ssh", BuildSSHArgs(inst, []string{remote})...)
	sshCmd.Stderr = os.Stderr
	return streamIn(tarCmd, sshCmd)
}

// SyncWorkspaceOut copies /workspace from the sandbox back into destDir,
// overwriting matching files. Ownership from the guest (user boite) is not
// preserved so the copies stay owned by the host user.
func SyncWorkspaceOut(inst *Instance, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}
	remote := fmt.Sprintf("tar -cf - -C %s .", workspaceDir)
	sshCmd := exec.Command("ssh", BuildSSHArgs(inst, []string{remote})...)
	sshCmd.Stderr = os.Stderr
	untarCmd := exec.Command("tar", "-xf", "-", "-C", destDir, "--no-same-owner")
	return streamOut(sshCmd, untarCmd, destDir)
}

// streamIn pipes the host tar producer's archive into the ssh consumer while
// both run, then sends an OK sentinel only if tar exited cleanly: the remote
// chain swaps the staged tree in only after reading it, so a failed or
// truncated stream never replaces the previous workspace. Errors report the
// first failure: tar, then ssh, then the pipe.
func streamIn(producer, consumer *exec.Cmd) error {
	producerOut, err := producer.StdoutPipe()
	if err != nil {
		return fmt.Errorf("tar workspace stdout: %w", err)
	}
	consumerIn, err := consumer.StdinPipe()
	if err != nil {
		return fmt.Errorf("workspace sync stdin: %w", err)
	}
	if err := producer.Start(); err != nil {
		return fmt.Errorf("tar workspace: %w", err)
	}
	if err := consumer.Start(); err != nil {
		producer.Process.Kill()
		producer.Wait()
		return fmt.Errorf("workspace sync over ssh: %w", err)
	}
	producerErr := pumpWithSentinel(consumerIn, producerOut, producer)
	consumerErr := consumer.Wait()
	return streamInError(producerErr, consumerErr)
}

// pumpWithSentinel streams producer's stdout into consumer's stdin, waits for
// the producer, and only on a clean exit sends the OK sentinel the remote
// chain requires before swapping the staged tree in. The returned error is
// the producer's exit status, or the copy error when tar exited cleanly.
func pumpWithSentinel(consumerIn io.WriteCloser, producerOut io.Reader, producer *exec.Cmd) error {
	_, copyErr := io.Copy(consumerIn, producerOut)
	producerErr := producer.Wait()
	if producerErr != nil {
		consumerIn.Close()
		return producerErr
	}
	consumerIn.Write([]byte("OK\n"))
	consumerIn.Close()
	if copyErr != nil {
		return fmt.Errorf("stream tar workspace: %w", copyErr)
	}
	return nil
}

// streamInError picks the first failure for reporting: tar, then ssh.
func streamInError(producerErr, consumerErr error) error {
	switch {
	case producerErr != nil:
		return fmt.Errorf("tar workspace: %w", producerErr)
	case consumerErr != nil:
		return fmt.Errorf("workspace sync over ssh: %w", consumerErr)
	}
	return nil
}

// streamOut pipes the ssh producer's archive into the host untar consumer
// while both run, and reports the first failure: ssh, then untar, then the
// pipe. Untar's stderr is captured so a failure includes what tar said.
func streamOut(producer, consumer *exec.Cmd, destDir string) error {
	var untarErr bytes.Buffer
	consumer.Stderr = &untarErr
	copyErr, producerErr, consumerErr := pipeStream(producer, consumer)
	switch {
	case producerErr != nil:
		return fmt.Errorf("tar workspace in VM: %w", producerErr)
	case consumerErr != nil:
		return fmt.Errorf("untar to %s: %w: %s", destDir, consumerErr, untarErr.String())
	case copyErr != nil:
		return fmt.Errorf("stream tar workspace in VM: %w", copyErr)
	}
	return nil
}

// pipeStream starts producer and consumer, streams producer's stdout into
// consumer's stdin, then waits for the copy and both processes. The copy runs
// concurrently with the waits so a full pipe never blocks the producer.
func pipeStream(producer, consumer *exec.Cmd) (copyErr, producerErr, consumerErr error) {
	producerOut, err := producer.StdoutPipe()
	if err != nil {
		producerErr = err
		return copyErr, producerErr, consumerErr
	}
	consumerIn, err := consumer.StdinPipe()
	if err != nil {
		consumerErr = err
		return copyErr, producerErr, consumerErr
	}
	if err := producer.Start(); err != nil {
		producerErr = err
		return copyErr, producerErr, consumerErr
	}
	if err := consumer.Start(); err != nil {
		producer.Process.Kill()
		producer.Wait()
		consumerErr = err
		return copyErr, producerErr, consumerErr
	}
	copied := make(chan error, 1)
	go func() {
		copied <- copyArchive(consumerIn, producerOut)
	}()
	consumerErr = consumer.Wait()
	producerErr = producer.Wait()
	copyErr = <-copied
	return copyErr, producerErr, consumerErr
}

// copyArchive drains producer's stdout into consumer's stdin until EOF, then
// closes the consumer's stdin so it sees the end of the stream.
func copyArchive(consumerIn io.WriteCloser, producerOut io.Reader) error {
	_, copyErr := io.Copy(consumerIn, producerOut)
	closeErr := consumerIn.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
