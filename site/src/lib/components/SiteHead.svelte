<script lang="ts">
  import { t } from "$lib/i18n/index.svelte";

  let { title }: { title: string } = $props();
</script>

<svelte:head>
  <title>{title}</title>
  <meta name="description" content={t("meta.description")} />
  <meta property="og:title" content={title} />
  <meta property="og:description" content={t("meta.description")} />
  <meta name="twitter:title" content={title} />
  <meta name="twitter:description" content={t("meta.description")} />
</svelte:head>
