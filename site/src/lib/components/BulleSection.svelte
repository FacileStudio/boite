<script lang="ts">
  import { Blob, gaze } from "@facile/blob";
  import { Badge, Button, Card, Icon } from "@facile/muse";
  import { BULLE_URL, CASIER_URL } from "$lib/site";
  import CodeBlock from "./CodeBlock.svelte";

  const eyes = gaze({ travel: 5, target: "pointer" });

  const workerCode = `import { request } from "node:http";

async function runAgentSandbox(taskId: string, sessionToken: string) {
  const box = await fetch("http://localhost/boxes", {
    method: "POST",
    headers: {
      Authorization: \`Bearer \${process.env.BOITE_WORKER_TOKEN}\`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      name: \`bulle-\${taskId}\`,
      cpus: 4,
      memory: "4G",
      net: "slirp",
      env: {
        CASIER_SESSION_TOKEN: sessionToken,
        PLAYWRIGHT_BROWSERS_PATH: "/opt/playwright",
      },
    }),
  }).then((r) => r.json());

  return box;
}`;
</script>

<section
  id="bulle"
  aria-labelledby="bulle-title"
  class="flex flex-col gap-8 rounded-2xl bg-fc-component p-6 sm:p-10"
>
  <div class="flex flex-col gap-4">
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

    <div class="flex flex-col gap-2">
      <h2 id="bulle-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
        Autonomous agent execution with hardware containment
      </h2>
      <p class="max-w-3xl text-fc-md leading-relaxed text-fc-fg-muted sm:text-fc-lg">
        Bulle delegates all untrusted AI code execution and web browsing to Boite micro-virtual machines.
        Agents compile Rust, run Go suites, and automate Playwright browsers inside genuine QEMU+KVM sandboxes with zero risk to the host cluster.
      </p>
    </div>
  </div>

  <div class="grid grid-cols-1 gap-6 lg:grid-cols-12 lg:items-start">
    <div class="flex flex-col gap-4 lg:col-span-6">
      <div class="flex flex-col gap-4 rounded-xl bg-fc-surface p-5">
        <div class="flex items-center gap-2">
          <Badge tone="accent">KVM Isolation</Badge>
          <span class="text-fc-xs text-fc-fg-muted">Kernel &amp; memory boundary</span>
        </div>
        <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
          Container breakouts cannot escape hardware virtualization. Each Bulle task runs inside its own virtualized Linux kernel, preventing cross-tenant interference and host compromise.
        </p>
      </div>

      <div class="flex flex-col gap-4 rounded-xl bg-fc-surface p-5">
        <div class="flex items-center gap-2">
          <Badge tone="success">Casier Integration</Badge>
          <span class="text-fc-xs text-fc-fg-muted">Scoped ephemeral tokens</span>
        </div>
        <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
          Secrets from <a href={CASIER_URL} target="_blank" rel="noreferrer" class="underline hover:text-fc-fg">Casier</a> are injected into guest memory for the lifetime of the session. Tokens expire automatically on sandbox destruction without writing to disk.
        </p>
      </div>

      <div class="flex flex-col gap-4 rounded-xl bg-fc-surface p-5">
        <div class="flex items-center gap-2">
          <Badge tone="info">Playwright &amp; Slirp Egress</Badge>
          <span class="text-fc-xs text-fc-fg-muted">Headless Chromium + egress filters</span>
        </div>
        <p class="text-fc-sm leading-relaxed text-fc-fg-muted">
          Pre-installed Chromium and Playwright run out of <code>/opt/playwright</code> with no setup latency. Slirp networking limits outbound destinations and prevents prompt-injection exfiltration.
        </p>
      </div>
    </div>

    <div class="flex flex-col gap-3 lg:col-span-6">
      <div class="flex items-center justify-between">
        <span class="text-fc-xs font-medium text-fc-fg-muted">boite-worker API dispatch</span>
        <Badge tone="neutral">HTTP Control Plane</Badge>
      </div>
      <CodeBlock code={workerCode} label="bulle-sandbox.ts" />
    </div>
  </div>
</section>
