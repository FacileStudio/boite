<script lang="ts">
  import SiteShell from "$lib/components/SiteShell.svelte";
  import SiteHead from "$lib/components/SiteHead.svelte";
  import Downloads from "$lib/components/Downloads.svelte";
  import { t } from "$lib/i18n/index.svelte";
</script>

<SiteHead title={t("meta.downloads")} />

<SiteShell>
  <Downloads page />
</SiteShell>
