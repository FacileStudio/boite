<script lang="ts">
  import { DownloadSection, type CliCommand, type DownloadEntry } from "@facile/muse";
  import { BASE_IMAGE_SIZE, BASE_IMAGE_URL, INSTALL_CMD, RELEASES_URL, SITE_URL } from "$lib/site";

  let { page = false }: { page?: boolean } = $props();

  const downloads: DownloadEntry[] = [
    { platform: "linux", href: BASE_IMAGE_URL, label: "Base image, Debian 13 qcow2", arch: "x64", size: BASE_IMAGE_SIZE },
  ];

  const commands: CliCommand[] = [
    { label: "Script", command: INSTALL_CMD },
    { label: "Facile", command: "facile install boite" },
    { label: "Base image", command: `curl -LO ${SITE_URL}${BASE_IMAGE_URL}` },
  ];
</script>

<section id="download" aria-labelledby="download-title" class="flex scroll-mt-24 flex-col gap-6">
  <div class="flex flex-col gap-2">
    <svelte:element
      this={page ? "h1" : "h2"}
      id="download-title"
      class="text-fc-xl font-bold tracking-tight text-fc-fg sm:text-3xl"
    >
      {page ? "Downloads" : "Install"}
    </svelte:element>
    <p class="max-w-xl text-fc-md text-fc-fg-muted">
      Linux with KVM. The CLI fetches the base image and checks its SHA-256 on first run.
    </p>
  </div>
  <DownloadSection
    {downloads}
    {commands}
    mode={page ? "all" : "device"}
    allHref="/downloads"
    releasesHref={RELEASES_URL}
    checksumsHref={`${RELEASES_URL}/latest/download/checksums.txt`}
    headingLevel={page ? 2 : 3}
  />
</section>
