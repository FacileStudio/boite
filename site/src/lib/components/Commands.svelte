<script lang="ts">
  import { Badge } from "@facile/muse";
  import CodeBlock from "./CodeBlock.svelte";

  const cliCommands = `boite create my-project
boite run my-project
boite exec my-project bun test
boite sync my-project
boite snapshot my-project clean
boite rollback my-project clean
boite stop my-project
boite rm my-project`;

  const workerApi = `POST /boxes
GET  /boxes
GET  /boxes/{id}
POST /boxes/{id}/start
POST /boxes/{id}/stop
POST /boxes/{id}/snapshot
POST /boxes/{id}/rollback
PUT  /boxes/{id}/net
POST /boxes/{id}/rm`;
</script>

<section id="commands" aria-labelledby="commands-title" class="flex flex-col gap-6">
  <div class="flex flex-col gap-1.5">
    <h2 id="commands-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      CLI &amp; daemon API
    </h2>
    <p class="max-w-xl text-fc-md text-fc-fg-muted">
      Drive from the terminal or automate over HTTP.
    </p>
  </div>

  <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-5 sm:p-6">
      <div class="flex items-center justify-between">
        <span class="text-fc-md font-semibold text-fc-fg">Developer CLI</span>
        <Badge tone="accent">boite</Badge>
      </div>
      <CodeBlock code={cliCommands} label="bash" lang="bash" />
    </div>

    <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-5 sm:p-6">
      <div class="flex items-center justify-between">
        <span class="text-fc-md font-semibold text-fc-fg">Host daemon</span>
        <Badge tone="neutral">boite-worker</Badge>
      </div>
      <CodeBlock code={workerApi} label="http" lang="http" />
    </div>
  </div>
</section>
