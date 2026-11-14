<script lang="ts">
  import { Button, icons } from "@facile/muse";
  import { REPO_URL } from "$lib/site";
  import HeroDemo from "./HeroDemo.svelte";
</script>

<section class="grid grid-cols-1 items-center gap-8 pt-4 lg:grid-cols-12 lg:gap-12">
  <div class="flex flex-col gap-5 lg:col-span-5">
    <h1 class="text-4xl font-bold tracking-tight text-fc-fg sm:text-5xl">Boite</h1>
    <p class="max-w-md text-fc-lg text-fc-fg-muted">
      Disposable QEMU virtual machines for your code and your AI agents.
    </p>
    <div class="flex flex-wrap items-center gap-3 pt-1">
      <Button variant="primary" size="lg" href="#download" icon={icons.download}>Install</Button>
      <Button variant="secondary" size="lg" href={REPO_URL} target="_blank" rel="noreferrer" icon="solar:code-square-linear">
        GitHub
      </Button>
    </div>
  </div>
  <div class="w-full min-w-0 max-w-[600px] lg:col-span-7">
    <HeroDemo />
  </div>
</section>
