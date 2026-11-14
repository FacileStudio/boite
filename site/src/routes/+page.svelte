<script lang="ts">
  import SiteShell from "$lib/components/SiteShell.svelte";
  import SiteHead from "$lib/components/SiteHead.svelte";
  import Hero from "$lib/components/Hero.svelte";
  import DemoTiles from "$lib/components/DemoTiles.svelte";
  import Features from "$lib/components/Features.svelte";
  import Suite from "$lib/components/Suite.svelte";
  import PlanNote from "$lib/components/PlanNote.svelte";
  import Downloads from "$lib/components/Downloads.svelte";
  import Developers from "$lib/components/Developers.svelte";
  import { t } from "$lib/i18n/index.svelte";
</script>

<SiteHead title={t("meta.title")} />

<SiteShell>
  <Hero />
  <DemoTiles />
  <Features />
  <Suite />
  <PlanNote />
  <Downloads />
  <Developers />
</SiteShell>
