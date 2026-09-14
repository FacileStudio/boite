package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd/qemu"
	"github.com/spf13/cobra"
)

// newSnapshotCmd builds the "snapshot" subcommand.
func newSnapshotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snapshot <name> <tag>",
		Short: "Snapshot a stopped sandbox's overlay disk",
		Long: `Record an internal qcow2 snapshot tagged <tag> on the overlay disk of a
stopped sandbox VM. Cheap retries: break things inside the VM, then roll back
to this snapshot with 'boite rollback' instead of recreating the sandbox.`,
		Args: cobra.ExactArgs(2),
		Run:  runSnapshotCreate,
	}
}

func runSnapshotCreate(cmd *cobra.Command, args []string) {
	name, tag := args[0], args[1]
	if err := qemu.SnapshotCreate(name, tag); err != nil {
		printError(fmt.Sprintf("Error: failed to snapshot '%s': %v", name, err))
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("Snapshot '%s' recorded for '%s'", tag, name))
}

// newRollbackCmd builds the "rollback" subcommand.
func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <name> <tag>",
		Short: "Roll a stopped sandbox's overlay back to a snapshot",
		Long: `Apply snapshot <tag> to the overlay disk of a stopped sandbox VM.
Everything written to the overlay since the snapshot was taken is discarded.
The VM must be stopped.`,
		Args: cobra.ExactArgs(2),
		Run:  runSnapshotRollback,
	}
}

func runSnapshotRollback(cmd *cobra.Command, args []string) {
	name, tag := args[0], args[1]
	if err := qemu.SnapshotRollback(name, tag); err != nil {
		printError(fmt.Sprintf("Error: failed to roll back '%s': %v", name, err))
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("'%s' rolled back to snapshot '%s'", name, tag))
}

// newSnapshotsCmd builds the "snapshots" subcommand.
func newSnapshotsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snapshots <name>",
		Short: "List a sandbox's overlay snapshots",
		Long:  `List the internal qcow2 snapshots recorded on a stopped sandbox's overlay disk.`,
		Args:  cobra.ExactArgs(1),
		Run:   runSnapshotList,
	}
}

func runSnapshotList(cmd *cobra.Command, args []string) {
	out, err := qemu.SnapshotList(args[0])
	if err != nil {
		printError(fmt.Sprintf("Error: failed to list snapshots for '%s': %v", args[0], err))
		os.Exit(1)
	}
	fmt.Println(out)
}
