<script lang="ts">
  import "../app.css";
  import type { Snippet } from "svelte";
  import { registerSiteIcons } from "$lib/icons";

  let { children }: { children: Snippet } = $props();
  registerSiteIcons();
</script>

{@render children()}
