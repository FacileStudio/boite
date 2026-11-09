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
  <title>Boite: QEMU virtual machines for code execution and AI agents</title>
  <meta
    name="description"
    content="Disposable QEMU virtual machines for code execution and AI agents. KVM isolation, overlay rollbacks, and zero host risk."
  />
  <meta property="og:title" content="Boite: QEMU virtual machines for code execution and AI agents" />
  <meta
    property="og:description"
    content="Disposable QEMU virtual machines for code execution and AI agents. KVM isolation, overlay rollbacks, and zero host risk."
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
