<script lang="ts">
  import { Badge, Button, Icon, icons } from "@facile/muse";
  import { BASE_IMAGE_OS, BASE_IMAGE_SIZE, BASE_IMAGE_URL, CREATE_CMD, INSTALL_CMD, REPO_URL } from "$lib/site";
  import CopyButton from "./CopyButton.svelte";

  const installScript = `${INSTALL_CMD}\n${CREATE_CMD}`;
</script>

<section class="grid grid-cols-1 gap-8 rounded-2xl bg-fc-component p-6 sm:p-10 lg:grid-cols-12 lg:items-center">
  <div class="flex flex-col gap-5 lg:col-span-7">
    <div class="flex flex-wrap items-center gap-2">
      <Badge tone="accent">QEMU &amp; KVM</Badge>
      <Badge tone="neutral">{BASE_IMAGE_OS}</Badge>
      <Badge tone="info">Zero cloud-init</Badge>
      <Badge tone="success">Instant rollbacks</Badge>
    </div>

    <div class="flex flex-col gap-2">
      <h1 class="text-fc-2xl font-semibold tracking-tight text-fc-fg sm:text-4xl md:text-5xl">
        Boite
      </h1>
      <p class="text-fc-lg font-medium text-fc-fg sm:text-fc-xl">
        Bare-metal development sandboxes in QEMU
      </p>
    </div>

    <p class="text-fc-md leading-relaxed text-fc-fg-muted sm:text-fc-lg">
      Spun up in seconds from a baked base image with full hardware isolation, instant snapshot rollbacks, and zero host contamination.
    </p>

    <div class="flex flex-col gap-2">
      <span class="text-fc-xs font-medium text-fc-fg-muted">Quick Install</span>
      <div class="flex items-center justify-between rounded-xl bg-fc-surface p-3 sm:p-4">
        <pre class="overflow-x-auto font-mono text-fc-xs text-fc-fg sm:text-fc-sm"><code>{installScript}</code></pre>
        <CopyButton text={installScript} label="Copy install commands" />
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-3 pt-2">
      <Button
        variant="primary"
        size="lg"
        href={BASE_IMAGE_URL}
        download="boite.qcow2"
        icon={icons.download}
      >
        Download base image ({BASE_IMAGE_SIZE})
      </Button>

      <Button
        variant="secondary"
        size="lg"
        href={REPO_URL}
        target="_blank"
        rel="noreferrer"
        icon="solar:code-square-linear"
      >
        GitHub
      </Button>
    </div>
  </div>

  <div class="flex flex-col gap-4 lg:col-span-5">
    <div class="flex flex-col gap-3 rounded-2xl bg-fc-surface p-5 sm:p-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="size-2.5 rounded-full bg-fc-fg-muted/40"></span>
          <span class="size-2.5 rounded-full bg-fc-fg-muted/40"></span>
          <span class="size-2.5 rounded-full bg-fc-fg-muted/40"></span>
          <span class="font-mono text-fc-xs text-fc-fg-muted">boite session</span>
        </div>
        <span class="rounded-fc-pill bg-fc-field px-2 py-0.5 font-mono text-fc-xs text-fc-success">
          isolated
        </span>
      </div>

      <div class="flex flex-col gap-3 font-mono text-fc-xs">
        <div class="rounded-xl bg-fc-field p-3">
          <div class="text-fc-fg-muted">$ boite create test-env</div>
          <div class="text-fc-fg font-medium">Created overlay boite-test-env.qcow2 in 0.12s</div>
        </div>

        <div class="rounded-xl bg-fc-field p-3">
          <div class="text-fc-fg-muted">$ boite run test-env</div>
          <div class="text-fc-fg">Booted QEMU+KVM (2 CPU, 2 GB RAM) in 2.1s</div>
          <div class="text-fc-info">Mounted host directory at /workspace</div>
          <div class="text-fc-success font-medium">boite@test-env:~$ _</div>
        </div>

        <div class="rounded-xl bg-fc-field p-3">
          <div class="text-fc-fg-muted">$ boite snapshot test-env clean</div>
          <div class="text-fc-fg">Snapshot 'clean' created on overlay disk</div>
          <div class="text-fc-fg-muted mt-1">$ boite rollback test-env clean</div>
          <div class="text-fc-success">Overlay rolled back to 'clean' in 38ms</div>
        </div>
      </div>
    </div>
  </div>
</section>
