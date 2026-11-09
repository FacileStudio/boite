<script lang="ts">
  import { Badge, Button, Icon, icons } from "@facile/muse";
  import { BASE_IMAGE_NAME, BASE_IMAGE_OS, BASE_IMAGE_SIZE, BASE_IMAGE_URL } from "$lib/site";
  import CopyButton from "./CopyButton.svelte";

  const directCurl = `curl -LO https://boite.facile.studio${BASE_IMAGE_URL}`;
</script>

<section id="download" aria-labelledby="download-title" class="flex flex-col gap-6 rounded-2xl bg-fc-component p-6 sm:p-10">
  <div class="flex flex-col gap-2">
    <div class="flex items-center gap-2">
      <Badge tone="accent">Base Image</Badge>
      <Badge tone="neutral">{BASE_IMAGE_OS}</Badge>
      <Badge tone="info">{BASE_IMAGE_SIZE}</Badge>
    </div>
    <h2 id="download-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      Download Base Image
    </h2>
    <p class="max-w-2xl text-fc-md text-fc-fg-muted">
      The official Debian 13 base disk image with baked-in compilers and Playwright binaries.
      The <code>boite</code> CLI downloads and verifies this automatically on first run.
    </p>
  </div>

  <div class="grid grid-cols-1 gap-6 md:grid-cols-2 md:items-center">
    <div class="flex flex-col gap-4">
      <div class="flex flex-col gap-2 rounded-xl bg-fc-surface p-4">
        <span class="text-fc-xs text-fc-fg-muted">File Specs</span>
        <div class="flex items-center justify-between text-fc-sm">
          <span class="text-fc-fg font-mono">{BASE_IMAGE_NAME}</span>
          <span class="text-fc-fg-muted font-mono">{BASE_IMAGE_SIZE}</span>
        </div>
        <div class="flex items-center justify-between text-fc-xs text-fc-fg-muted">
          <span>Format: QEMU qcow2</span>
          <span>Compression: ZSTD</span>
        </div>
      </div>

      <div class="flex flex-col gap-2">
        <span class="text-fc-xs text-fc-fg-muted">Direct Download URL</span>
        <div class="flex items-center justify-between rounded-xl bg-fc-surface p-3 font-mono text-fc-xs">
          <span class="truncate text-fc-fg">{directCurl}</span>
          <CopyButton text={directCurl} label="Copy curl command" />
        </div>
      </div>
    </div>

    <div class="flex flex-col items-start gap-4 md:items-end">
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
        SHA-256 checksum verified by CLI on download
      </span>
    </div>
  </div>
</section>
