<script lang="ts">
  import { DownloadSection, LANDING_HEADING, type CliCommand, type DownloadEntry } from "@facile/muse";
  import { t } from "$lib/i18n/index.svelte";
  import { BASE_IMAGE_SIZE, BASE_IMAGE_URL, INSTALL_CMD, RELEASES_URL, SITE_URL } from "$lib/site";

  let { page = false }: { page?: boolean } = $props();

  const downloads: DownloadEntry[] = $derived([
    { platform: "linux", href: BASE_IMAGE_URL, label: t("install.image"), arch: "x64", size: BASE_IMAGE_SIZE },
  ]);

  const commands: CliCommand[] = $derived([
    { label: t("install.script"), command: INSTALL_CMD },
    { label: "Facile", command: "facile install boite" },
    { label: t("install.baseImage"), command: `curl -LO ${SITE_URL}${BASE_IMAGE_URL}` },
  ]);
</script>

<section id="download" aria-labelledby="download-title" class="flex scroll-mt-24 flex-col gap-6">
  <div class="flex flex-col gap-3">
    <svelte:element this={page ? "h1" : "h2"} id="download-title" class={LANDING_HEADING}>
      {page ? t("nav.downloads") : t("install.title")}
    </svelte:element>
    <p class="max-w-2xl text-fc-md text-fc-fg-muted">{t("install.line")}</p>
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
