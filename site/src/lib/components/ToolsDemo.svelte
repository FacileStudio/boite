<script lang="ts">
  import { Icon, icons } from "@facile/muse";
  import { Stage, Window, type Act } from "@facile/mise-en-scene";
  import { t } from "$lib/i18n/index.svelte";
  import "./tile-demo.css";

  const TOOLS = [
    { name: "Go", bin: "go" },
    { name: "Rust", bin: "cargo" },
    { name: "Node", bin: "node" },
    { name: "Python", bin: "python3" },
    { name: "Bun", bin: "bun" },
  ];

  let ready = $state(0);

  function reset() {
    ready = 0;
  }

  function poster() {
    ready = TOOLS.length;
  }

  async function script(act: Act) {
    await act.wait(900);
    for (const i of TOOLS.keys()) {
      ready = i + 1;
      await act.wait(550);
    }
    await act.wait(3000);
    reset();
    await act.wait(900);
  }
</script>

<Stage width={380} height={400} label={t("demo.tools.label")} {script} {reset} {poster}>
  {#snippet children()}
    <Window title="demo" style="top:14px;left:14px;width:352px;height:372px">
      <div class="td-tools">
        <code class="td-cmd"><span class="td-dim">$</span> boite create demo</code>
        <ul>
          {#each TOOLS as tool, i (tool.name)}
            <li data-done={i < ready || undefined}>
              <span class="td-name">{tool.name}<code>{tool.bin}</code></span>
              <span class="td-ready"><Icon icon={icons.check} size={14} />{t("demo.tools.ready")}</span>
            </li>
          {/each}
        </ul>
      </div>
    </Window>
  {/snippet}
</Stage>
