<script lang="ts">
  import { Icon } from "@facile/muse";

  const features = [
    {
      title: "A real VM boundary",
      icon: "solar:cpu-bolt-bold-duotone",
      desc: "Each sandbox boots its own kernel under KVM. Root in the guest stays in the guest.",
    },
    {
      title: "Snapshot and roll back",
      icon: "solar:history-bold-duotone",
      desc: "Save a clean state with boite snapshot, then undo any damage with boite rollback.",
    },
    {
      title: "Toolchains baked in",
      icon: "solar:bolt-circle-bold-duotone",
      desc: "Go, Rust, Bun, Node, Python and Playwright ship in the Debian 13 base image.",
    },
    {
      title: "Offline on demand",
      icon: "solar:shield-check-bold-duotone",
      desc: "Set vm.net to offline and the guest reaches nothing. SSH still works.",
    },
  ];
</script>

<section id="features" aria-labelledby="features-title" class="flex flex-col gap-6">
  <h2 id="features-title" class="text-fc-xl font-bold tracking-tight text-fc-fg sm:text-3xl">Why a VM</h2>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
    {#each features as f (f.title)}
      <div class="flex flex-col gap-2 rounded-fc-lg bg-fc-component p-5 sm:p-6">
        <div class="flex size-10 items-center justify-center rounded-fc-md bg-fc-surface text-fc-accent">
          <Icon icon={f.icon} size={20} />
        </div>
        <h3 class="text-fc-md font-bold text-fc-fg">{f.title}</h3>
        <p class="text-fc-sm text-fc-fg-muted">{f.desc}</p>
      </div>
    {/each}
  </div>
</section>
