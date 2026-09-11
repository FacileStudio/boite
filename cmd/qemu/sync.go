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
// remote side clears only /workspace first, so the host tree never leaks paths
// or permissions across the boundary: the guest only ever touches its own
// mounted-empty directory.
func SyncWorkspaceIn(inst *Instance, srcDir string) error {
	tarCmd := exec.Command("tar", "-C", srcDir, "-cf", "-", ".")
	tarCmd.Stderr = os.Stderr
	remote := fmt.Sprintf("mkdir -p %s && find %s -mindepth 1 -delete && tar -xf - -C %s",
		workspaceDir, workspaceDir, workspaceDir)
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
// both run, and reports the first failure: tar, then ssh, then the pipe.
func streamIn(producer, consumer *exec.Cmd) error {
	copyErr, producerErr, consumerErr := pipeStream(producer, consumer)
	switch {
	case producerErr != nil:
		return fmt.Errorf("tar workspace: %w", producerErr)
	case consumerErr != nil:
		return fmt.Errorf("workspace sync over ssh: %w", consumerErr)
	case copyErr != nil:
		return fmt.Errorf("stream tar workspace: %w", copyErr)
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
		producerErr = err
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
