<script lang="ts">
  import SiteShell from "$lib/components/SiteShell.svelte";
  import Hero from "$lib/components/Hero.svelte";
  import BulleSection from "$lib/components/BulleSection.svelte";
  import Features from "$lib/components/Features.svelte";
  import Toolchain from "$lib/components/Toolchain.svelte";
  import Commands from "$lib/components/Commands.svelte";
  import DownloadSection from "$lib/components/DownloadSection.svelte";
</script>

<svelte:head>
  <title>Boite — Bare-metal development sandboxes in QEMU</title>
  <meta
    name="description"
    content="Spun up in seconds from a baked base image with full hardware isolation, instant snapshot rollbacks, and zero host contamination."
  />
  <meta property="og:title" content="Boite — Bare-metal development sandboxes in QEMU" />
  <meta
    property="og:description"
    content="Spun up in seconds from a baked base image with full hardware isolation, instant snapshot rollbacks, and zero host contamination."
  />
  <meta property="og:url" content="https://boite.facile.studio" />
  <meta name="twitter:card" content="summary_large_image" />
</svelte:head>

<SiteShell>
  <Hero />
  <BulleSection />
  <Features />
  <Toolchain />
  <Commands />
  <DownloadSection />
</SiteShell>
