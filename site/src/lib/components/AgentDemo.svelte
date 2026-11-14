<script lang="ts">
  import { Blob } from "@facile/blob";
  import { Icon, icons } from "@facile/muse";
  import { Bar, Stage, Window, type Act } from "@facile/mise-en-scene";
  import { t } from "$lib/i18n/index.svelte";
  import "./agent-demo.css";

  type Step = "hidden" | "running" | "ok";

  const CALLS = ["git clone -b fix/login app", "bun install", "bun test"];

  const prompt = $derived(t("demo.agent.prompt"));
  const answer = $derived(t("demo.agent.answer"));

  let draft = $state("");
  let sent = $state(false);
  let steps = $state<Step[]>(["hidden", "hidden", "hidden"]);
  let done = $state(false);

  function reset() {
    draft = "";
    sent = false;
    steps = ["hidden", "hidden", "hidden"];
    done = false;
  }

  function poster() {
    sent = true;
    steps = ["ok", "ok", "ok"];
    done = true;
  }

  async function script(act: Act) {
    await act.wait(500);
    await act.type(prompt, (v) => (draft = v), 40);
    await act.wait(350);
    sent = true;
    draft = "";
    for (const i of CALLS.keys()) {
      await act.wait(350);
      steps[i] = "running";
      await act.wait(900);
      steps[i] = "ok";
    }
    await act.wait(400);
    done = true;
    await act.wait(3200);
    reset();
    await act.wait(700);
  }
</script>

<Stage width={380} height={400} label={t("demo.agent.label")} {script} {reset} {poster}>
  {#snippet children()}
    <Window title="Bulle" style="top:14px;left:14px;width:352px;height:372px">
      <div class="bd-thread">
        <div class="bd-past">
          <span class="bd-bubble-bar"></span>
          <Bar w="70%" />
          <Bar w="52%" />
          <span class="bd-bubble-bar bd-bubble-short"></span>
          <Bar w="64%" />
        </div>
        <div class="bd-q bd-slot" data-on={sent || undefined}><p>{prompt}</p></div>
        <section class="bd-group bd-slot" data-on={steps[0] !== "hidden" || undefined}>
          <header>
            <Blob name="Bulle" size={18} animate="always" class="size-[18px] shrink-0" />
            <b>Boîte</b>
          </header>
          <ul>
            {#each CALLS as call, i (i)}
              <li class="bd-slot" data-on={steps[i] !== "hidden" || undefined}>
                <span class="bd-mark">
                  {#if steps[i] === "ok"}<Icon icon={icons.check} size={14} />{:else}<span class="bd-spin"></span>{/if}
                </span>
                <b>bash</b><code>{call}</code>
              </li>
            {/each}
          </ul>
        </section>
        <p class="bd-answer bd-slot" data-on={done || undefined}>{answer}</p>
      </div>
      <div class="bd-composer">
        <span class:bd-placeholder={!draft}>{draft || t("demo.agent.composer")}</span>
        <span class="bd-send" class:bd-idle={!draft}></span>
      </div>
    </Window>
  {/snippet}
</Stage>
