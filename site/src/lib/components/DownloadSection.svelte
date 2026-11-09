<script lang="ts">
  import { Badge, Button, icons } from "@facile/muse";
  import { BASE_IMAGE_NAME, BASE_IMAGE_OS, BASE_IMAGE_SIZE, BASE_IMAGE_URL } from "$lib/site";
  import CodeBlock from "./CodeBlock.svelte";

  const directCurl = `curl -LO https://boite.facile.studio${BASE_IMAGE_URL}`;
</script>

<section id="download" aria-labelledby="download-title" class="flex flex-col gap-6 rounded-2xl bg-fc-component p-6 sm:p-10">
  <div class="flex flex-col gap-1.5">
    <div class="flex items-center gap-2">
      <Badge tone="accent">Base image</Badge>
      <Badge tone="neutral">{BASE_IMAGE_OS}</Badge>
      <Badge tone="info">{BASE_IMAGE_SIZE}</Badge>
    </div>
    <h2 id="download-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      Base image
    </h2>
    <p class="max-w-xl text-fc-md text-fc-fg-muted">
      Debian 13 base disk with baked compilers. The CLI downloads and verifies it automatically on first run.
    </p>
  </div>

  <div class="grid grid-cols-1 gap-6 md:grid-cols-2 md:items-center">
    <div class="flex flex-col gap-3">
      <div class="flex items-center justify-between rounded-xl bg-fc-surface p-4 font-mono text-fc-sm">
        <span class="text-fc-fg">{BASE_IMAGE_NAME}</span>
        <span class="text-fc-fg-muted">{BASE_IMAGE_SIZE} · ZSTD</span>
      </div>
      <CodeBlock code={directCurl} label="curl" lang="bash" />
    </div>

    <div class="flex flex-col items-start gap-2 md:items-end">
      <Button
        variant="primary"
        size="lg"
        href={BASE_IMAGE_URL}
        download="boite.qcow2"
        icon={icons.download}
      >
        Download base image ({BASE_IMAGE_SIZE})
      </Button>
      <span class="text-fc-xs text-fc-fg-muted">
        SHA-256 verified by CLI on download
      </span>
    </div>
  </div>
</section>
