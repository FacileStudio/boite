<script lang="ts">
  import SiteShell from "$lib/components/SiteShell.svelte";
  import Hero from "$lib/components/Hero.svelte";
  import BulleSection from "$lib/components/BulleSection.svelte";
  import Features from "$lib/components/Features.svelte";
  import Toolchain from "$lib/components/Toolchain.svelte";
  import Commands from "$lib/components/Commands.svelte";
  import DownloadSection from "$lib/components/DownloadSection.svelte";
</script>


<SiteShell>
  <Hero />
  <BulleSection />
  <Features />
  <Toolchain />
  <Commands />
  <DownloadSection />
</SiteShell>
