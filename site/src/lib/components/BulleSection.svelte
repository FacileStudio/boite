<script lang="ts">
  import { Blob, gaze } from "@facile/blob";
  import { Badge, Icon } from "@facile/muse";
  import { BULLE_URL, CASIER_URL } from "$lib/site";
  import CodeBlock from "./CodeBlock.svelte";

  const eyes = gaze({ travel: 5, target: "pointer" });

  const workerCode = `const box = await fetch("http://worker/boxes", {
  method: "POST",
  headers: {
    Authorization: \`Bearer \${process.env.BOITE_WORKER_TOKEN}\`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    name: "bulle-agent-task",
    cpus: 4,
    memory: "4G",
    net: "offline",
    env: {
      CASIER_TOKEN: sessionToken,
      PLAYWRIGHT_BROWSERS_PATH: "/opt/playwright",
    },
  }),
}).then((r) => r.json());`;
</script>

<section
  id="bulle"
  aria-labelledby="bulle-title"
  class="flex flex-col gap-6 rounded-2xl bg-fc-component p-6 sm:p-10"
>
  <div class="flex flex-col gap-3">
    <a
      href={BULLE_URL}
      target="_blank"
      rel="noreferrer"
      class="group inline-flex w-fit items-center gap-3 rounded-full bg-fc-surface px-3 py-1.5 text-fc-sm text-fc-fg transition-colors hover:bg-fc-field"
      aria-label="Bulle Autonomous Agents"
    >
      <Blob {@attach eyes} name="Bulle" size={28} animate="always" class="shrink-0" />
      <span class="font-medium">Bulle</span>
      <span class="text-fc-xs text-fc-fg-muted">bulle.facile.studio</span>
      <Icon
        icon="solar:arrow-right-up-linear"
        size={14}
        class="text-fc-fg-muted transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-fc-fg"
      />
    </a>

    <div class="flex flex-col gap-1.5">
      <h2 id="bulle-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
        AI agent execution with hardware containment
      </h2>
      <p class="max-w-2xl text-fc-md text-fc-fg-muted">
        Bulle runs untrusted AI code and headless browsers in Boite micro-VMs. Real Linux kernel, memory boundary, ephemeral tokens.
      </p>
    </div>
  </div>

  <div class="grid grid-cols-1 gap-6 lg:grid-cols-12 lg:items-start">
    <div class="flex flex-col gap-3 lg:col-span-5">
      <div class="flex flex-col gap-1.5 rounded-xl bg-fc-surface p-4">
        <div class="flex items-center gap-2">
          <Badge tone="accent">Hardware containment</Badge>
        </div>
        <p class="text-fc-sm text-fc-fg-muted">
          Each agent task boots its own Linux kernel. Container escapes cannot touch the host.
        </p>
      </div>

      <div class="flex flex-col gap-1.5 rounded-xl bg-fc-surface p-4">
        <div class="flex items-center gap-2">
          <Badge tone="success">Ephemeral secrets</Badge>
        </div>
        <p class="text-fc-sm text-fc-fg-muted">
          <a href={CASIER_URL} target="_blank" rel="noreferrer" class="underline hover:text-fc-fg">Casier</a> tokens live only in guest memory with TTL expiry, destroyed on exit.
        </p>
      </div>

      <div class="flex flex-col gap-1.5 rounded-xl bg-fc-surface p-4">
        <div class="flex items-center gap-2">
          <Badge tone="info">Playwright &amp; network filter</Badge>
        </div>
        <p class="text-fc-sm text-fc-fg-muted">
          Chromium in <code>/opt/playwright</code> with slirp and passt egress controls to prevent data exfiltration.
        </p>
      </div>
    </div>

    <div class="flex flex-col gap-2 lg:col-span-7">
      <CodeBlock code={workerCode} label="boite-worker API dispatch" lang="typescript" />
    </div>
  </div>
</section>
