<script lang="ts">
  import { Badge } from "@facile/muse";
  import CodeBlock from "./CodeBlock.svelte";

  const cliCommands = `# Create a new sandbox VM with current directory mapped to /workspace
boite create my-project

# Launch the VM and open an interactive zsh shell
boite run my-project

# Run commands non-interactively
boite exec my-project bun test

# Copy changes from /workspace back to host directory
boite sync my-project

# Checkpoint the qcow2 overlay before running risky operations
boite snapshot my-project pre-install

# Roll back the overlay if dependencies break
boite rollback my-project pre-install

# List and cleanup
boite list
boite stop my-project
boite rm my-project`;

  const workerApi = `POST   /boxes                  # Create sandbox with CPU/RAM/Net specs
GET    /boxes                  # List managed virtual machines
GET    /boxes/{id}             # Inspect sandbox status and IP
POST   /boxes/{id}/start       # Boot stopped sandbox
POST   /boxes/{id}/stop        # Gracefully power off
POST   /boxes/{id}/rm          # Destroy sandbox and delete disk overlay
PUT    /boxes/{id}/net         # Toggle network mode (offline | local | open)
PUT    /boxes/{id}/ssh-key     # Inject authorized public key
POST   /boxes/{id}/snapshot    # Take named qcow2 snapshot
POST   /boxes/{id}/rollback    # Revert to named snapshot
GET    /boxes/{id}/snapshots   # List available snapshots`;
</script>

<section id="commands" aria-labelledby="commands-title" class="flex flex-col gap-6">
  <div class="flex flex-col gap-2">
    <h2 id="commands-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      CLI &amp; Worker Control Plane
    </h2>
    <p class="max-w-2xl text-fc-md text-fc-fg-muted">
      Drive sandboxes interactively from the terminal or automate them over the HTTP daemon.
    </p>
  </div>

  <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-fc-md font-semibold text-fc-fg">Terminal CLI</span>
          <Badge tone="accent">boite</Badge>
        </div>
        <span class="text-fc-xs text-fc-fg-muted">Interactive Development</span>
      </div>
      <p class="text-fc-sm text-fc-fg-muted">
        Manage local VM lifecycle with single-command workspace mounting and rollback.
      </p>
      <CodeBlock code={cliCommands} label="bash" />
    </div>

    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-fc-md font-semibold text-fc-fg">Host Daemon API</span>
          <Badge tone="neutral">boite-worker</Badge>
        </div>
        <span class="text-fc-xs text-fc-fg-muted">Unix Socket / HTTP</span>
      </div>
      <p class="text-fc-sm text-fc-fg-muted">
        Control plane driven by Bulle, CI pipelines, and autonomous agent orchestrators.
      </p>
      <CodeBlock code={workerApi} label="http" />
    </div>
  </div>
</section>
