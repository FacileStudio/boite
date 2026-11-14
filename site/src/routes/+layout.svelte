<script lang="ts">
  import "../app.css";
  import type { Snippet } from "svelte";
  import { provideLocale } from "@facile/muse";
  import { registerSiteIcons } from "$lib/icons";
  import { bindLocale, localeSource } from "$lib/i18n/index.svelte";

  let { children }: { children: Snippet } = $props();

  const i18n = provideLocale(localeSource);
  bindLocale(() => i18n.locale);
  registerSiteIcons();
</script>

{@render children()}
