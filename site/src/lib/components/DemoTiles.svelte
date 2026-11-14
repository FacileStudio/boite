<script lang="ts">
  import { LANDING_HEADING } from "@facile/muse";
  import { Tile } from "@facile/mise-en-scene";
  import { t } from "$lib/i18n/index.svelte";
  import AgentDemo from "./AgentDemo.svelte";
  import RollbackDemo from "./RollbackDemo.svelte";
  import ToolsDemo from "./ToolsDemo.svelte";
</script>

<section id="demos" aria-labelledby="demos-title" class="flex scroll-mt-24 flex-col gap-8">
  <h2 id="demos-title" class={LANDING_HEADING}>{t("cards.title")}</h2>
  <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
    <Tile title={t("cards.agent.title")} body={t("cards.agent.body")}>
      <AgentDemo />
    </Tile>
    <Tile title={t("cards.rollback.title")} body={t("cards.rollback.body")}>
      <RollbackDemo />
    </Tile>
    <Tile class="md:col-span-2 lg:col-span-1" title={t("cards.tools.title")} body={t("cards.tools.body")}>
      <ToolsDemo />
    </Tile>
  </div>
</section>
