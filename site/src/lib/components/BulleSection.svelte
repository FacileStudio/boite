<script lang="ts">
  import { Blob, gaze } from "@facile/blob";
  import { Icon } from "@facile/muse";
  import { BULLE_URL } from "$lib/site";
  import BulleDemo from "./BulleDemo.svelte";

  const eyes = gaze({ travel: 5, target: "pointer" });
</script>

<section id="bulle" aria-labelledby="bulle-title" class="grid grid-cols-1 items-center gap-8 lg:grid-cols-12 lg:gap-12">
  <div class="flex flex-col gap-4 lg:col-span-5">
    <a
      href={BULLE_URL}
      target="_blank"
      rel="noreferrer"
      class="group inline-flex min-h-11 w-fit items-center gap-3 rounded-fc-pill bg-fc-component px-3 py-1.5 text-fc-sm text-fc-fg transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
    >
      <Blob {@attach eyes} name="Bulle" size={28} animate="always" class="shrink-0" />
      <span class="font-bold">Bulle</span>
      <span class="text-fc-xs text-fc-fg-muted">bulle.facile.studio</span>
      <Icon
        icon="solar:arrow-right-up-linear"
        size={14}
        class="text-fc-fg-muted transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:text-fc-fg"
      />
    </a>
    <h2 id="bulle-title" class="text-fc-xl font-bold tracking-tight text-fc-fg sm:text-3xl">Where Bulle runs code</h2>
    <p class="max-w-md text-fc-md text-fc-fg-muted">
      Bulle agents run commands in a Boite VM with its own kernel, never on your machine.
    </p>
  </div>
  <div class="w-full min-w-0 max-w-[720px] lg:col-span-7">
    <BulleDemo />
  </div>
</section>
