<script lang="ts">
  import { Badge, Button, icons } from "@facile/muse";
  import { BASE_IMAGE_OS, BASE_IMAGE_SIZE, BASE_IMAGE_URL, CREATE_CMD, INSTALL_CMD, REPO_URL } from "$lib/site";
  import CodeBlock from "./CodeBlock.svelte";

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

    <div class="flex flex-col gap-1.5">
      <h1 class="text-fc-2xl font-semibold tracking-tight text-fc-fg sm:text-4xl md:text-5xl">
        Boite
      </h1>
      <p class="text-fc-lg font-medium text-fc-fg sm:text-fc-xl">
        QEMU virtual machines for code execution and AI agents
      </p>
    </div>

    <p class="text-fc-md text-fc-fg-muted">
      Boots in seconds from a pre-built base image. KVM hardware isolation, overlay rollbacks, zero host risk.
    </p>

    <div class="flex flex-col gap-2">
      <CodeBlock code={installScript} label="Quick install" lang="bash" />
    </div>

    <div class="flex flex-wrap items-center gap-3 pt-1">
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
          <span class="font-mono text-fc-xs text-fc-fg-muted">boite demo</span>
        </div>
        <span class="rounded-fc-pill bg-fc-field px-2 py-0.5 font-mono text-fc-xs text-fc-success">
          isolated
        </span>
      </div>

      <div class="flex flex-col gap-2.5 font-mono text-fc-xs">
        <div class="rounded-xl bg-fc-field p-3">
          <div class="text-fc-fg-muted">$ boite create demo</div>
          <div class="text-fc-fg font-medium">✓ overlay created in 0.12s</div>
        </div>

        <div class="rounded-xl bg-fc-field p-3">
          <div class="text-fc-fg-muted">$ boite run demo</div>
          <div class="text-fc-fg">✓ booted QEMU+KVM in 2.1s · /workspace mounted</div>
          <div class="text-fc-success font-medium">boite@demo:~$ _</div>
        </div>

        <div class="rounded-xl bg-fc-field p-3">
          <div class="text-fc-fg-muted">$ boite rollback demo clean</div>
          <div class="text-fc-success font-medium">✓ overlay restored in 38ms</div>
        </div>
      </div>
    </div>
  </div>
</section>
