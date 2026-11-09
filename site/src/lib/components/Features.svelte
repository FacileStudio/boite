<script lang="ts">
  import { Icon } from "@facile/muse";

  const features = [
    {
      title: "Zero cloud-init",
      icon: "solar:bolt-circle-bold-duotone",
      desc: "Baked with virt-builder. Boots straight into a ready shell in under 3 seconds.",
    },
    {
      title: "Instant rollback",
      icon: "solar:history-bold-duotone",
      desc: "Thin qcow2 copy-on-write overlay. Revert damaged state in milliseconds.",
    },
    {
      title: "SSH workspace sync",
      icon: "solar:folder-with-files-bold-duotone",
      desc: "Host directory mounted into /workspace on launch. Edit locally, run isolated.",
    },
    {
      title: "KVM hardware isolation",
      icon: "solar:cpu-bolt-bold-duotone",
      desc: "Real virtual machine boundary. Root in the guest cannot reach the host.",
    },
    {
      title: "Egress controls",
      icon: "solar:shield-check-bold-duotone",
      desc: "Set vm.net: offline to kill outbound traffic, or restrict egress via passt.",
    },
    {
      title: "Session-scoped secrets",
      icon: "solar:lock-keyhole-minimalistic-bold-duotone",
      desc: "Inject Casier keys into guest memory with automatic TTL expiry.",
    },
  ];
</script>

<section id="features" aria-labelledby="features-title" class="flex flex-col gap-6">
  <div class="flex flex-col gap-1.5">
    <h2 id="features-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      Architecture &amp; security
    </h2>
    <p class="max-w-xl text-fc-md text-fc-fg-muted">
      Disposable environments without host risks.
    </p>
  </div>

  <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
    {#each features as f (f.title)}
      <div class="flex flex-col gap-2 rounded-2xl bg-fc-component p-5 sm:p-6">
        <div class="flex size-9 items-center justify-center rounded-xl bg-fc-surface text-fc-accent">
          <Icon icon={f.icon} size={20} />
        </div>
        <h3 class="text-fc-md font-semibold text-fc-fg">{f.title}</h3>
        <p class="text-fc-sm text-fc-fg-muted">
          {f.desc}
        </p>
      </div>
    {/each}
  </div>
</section>
