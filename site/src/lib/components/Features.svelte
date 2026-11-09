<script lang="ts">
  import { Badge, Card, Icon } from "@facile/muse";
</script>

<section id="features" aria-labelledby="features-title" class="flex flex-col gap-6">
  <div class="flex flex-col gap-2">
    <h2 id="features-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      Architecture &amp; Security
    </h2>
    <p class="max-w-2xl text-fc-md text-fc-fg-muted">
      Built for developers and autonomous systems who need fast disposable environments without host risks.
    </p>
  </div>

  <div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6 sm:p-8">
      <div class="flex size-10 items-center justify-center rounded-xl bg-fc-surface text-fc-accent">
        <Icon icon="solar:bolt-circle-bold-duotone" size={22} />
      </div>
      <h3 class="text-fc-md font-semibold text-fc-fg">Zero Cloud-Init</h3>
      <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
        Baked with <code>virt-builder</code> at image compile time. No first-boot cloud-init scripts, no apt-get waits, and no slow package installs. Boots directly into a ready shell in under 3 seconds.
      </p>
    </div>

    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6 sm:p-8">
      <div class="flex size-10 items-center justify-center rounded-xl bg-fc-surface text-fc-accent">
        <Icon icon="solar:history-bold-duotone" size={22} />
      </div>
      <h3 class="text-fc-md font-semibold text-fc-fg">Instant Snapshots &amp; Rollback</h3>
      <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
        Every sandbox writes to a thin qcow2 copy-on-write overlay. Checkpoint state with <code>boite snapshot</code> and restore with <code>boite rollback</code> in milliseconds to discard corrupted environments.
      </p>
    </div>

    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6 sm:p-8">
      <div class="flex size-10 items-center justify-center rounded-xl bg-fc-surface text-fc-accent">
        <Icon icon="solar:folder-with-files-bold-duotone" size={22} />
      </div>
      <h3 class="text-fc-md font-semibold text-fc-fg">SSH Workspace Sync</h3>
      <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
        Your current host directory mounts automatically into <code>/workspace</code> inside the VM via SSH. Edit code on your host editor; test and compile inside the sandbox.
      </p>
    </div>

    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6 sm:p-8">
      <div class="flex size-10 items-center justify-center rounded-xl bg-fc-surface text-fc-accent">
        <Icon icon="solar:cpu-bolt-bold-duotone" size={22} />
      </div>
      <h3 class="text-fc-md font-semibold text-fc-fg">Hardware KVM Isolation</h3>
      <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
        Kernel-level virtualization with QEMU and KVM hardware acceleration. Root access in the VM cannot touch your host file system, kernel memory, or docker daemon.
      </p>
    </div>

    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6 sm:p-8">
      <div class="flex size-10 items-center justify-center rounded-xl bg-fc-surface text-fc-accent">
        <Icon icon="solar:shield-check-bold-duotone" size={22} />
      </div>
      <h3 class="text-fc-md font-semibold text-fc-fg">Guest Egress Control</h3>
      <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
        Isolate network access per sandbox. Set <code>vm.net: offline</code> to block all outbound traffic, or restrict egress with slirp and passt to stop data exfiltration before running untrusted scripts.
      </p>
    </div>

    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-6 sm:p-8">
      <div class="flex size-10 items-center justify-center rounded-xl bg-fc-surface text-fc-accent">
        <Icon icon="solar:lock-keyhole-minimalistic-bold-duotone" size={22} />
      </div>
      <h3 class="text-fc-md font-semibold text-fc-fg">Session-Scoped Secrets</h3>
      <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
        Inject secrets securely from Casier into the guest environment. Tokens stay in memory with automatic TTL expiry and evaporate when the sandbox stops, leaving no traces on disk.
      </p>
    </div>
  </div>
</section>
