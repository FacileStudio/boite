<script lang="ts">
  import SiteShell from "$lib/components/SiteShell.svelte";
  import Hero from "$lib/components/Hero.svelte";
  import BulleSection from "$lib/components/BulleSection.svelte";
  import Features from "$lib/components/Features.svelte";
  import Downloads from "$lib/components/Downloads.svelte";
  import ClosingCta from "$lib/components/ClosingCta.svelte";
</script>

<svelte:head>
  <title>Boîte · Disposable VM sandboxes for development</title>
</svelte:head>

<SiteShell>
  <Hero />
  <BulleSection />
  <Features />
  <Downloads />
  <ClosingCta />
</SiteShell>
