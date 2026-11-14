<script lang="ts">
  import { Icon, icons } from "@facile/muse";
  import { Caret, Stage, Window, type Act } from "@facile/mise-en-scene";
  import { t } from "$lib/i18n/index.svelte";
  import "./tile-demo.css";

  type Key = "demo.rollback.saved" | "demo.rollback.broken" | "demo.rollback.back";
  type Step = { cmd: string; out: string; note: Key; bad?: boolean };

  const STEPS: Step[] = [
    { cmd: "boite snapshot demo clean", out: "✓ Snapshot 'clean' recorded for 'demo'", note: "demo.rollback.saved" },
    { cmd: "boite exec demo sudo rm -rf /usr", out: "", note: "demo.rollback.broken", bad: true },
    { cmd: "boite rollback demo clean", out: "✓ 'demo' rolled back to snapshot 'clean'", note: "demo.rollback.back" },
  ];

  let typed = $state<string[]>([STEPS[0].cmd, "", ""]);
  let shown = $state(1);
  let typing = $state(-1);

  const broken = $derived(shown === 2);

  function reset() {
    typed = [STEPS[0].cmd, "", ""];
    shown = 1;
    typing = -1;
  }

  function poster() {
    typed = STEPS.map((s) => s.cmd);
    shown = 3;
    typing = -1;
  }

  async function script(act: Act) {
    await act.wait(1200);
    for (const i of [1, 2]) {
      typing = i;
      await act.type(STEPS[i].cmd, (v) => (typed[i] = v), 45);
      await act.wait(300);
      typing = -1;
      shown = i + 1;
      await act.wait(1600);
    }
    await act.wait(2400);
    reset();
    await act.wait(900);
  }
</script>

<Stage width={380} height={400} label={t("demo.rollback.label")} {script} {reset} {poster}>
  {#snippet children()}
    <Window title="demo" style="top:14px;left:14px;width:352px;height:372px">
      {#snippet trailing()}<span class="td-state" data-bad={broken || undefined}></span>{/snippet}
      <ol class="td-steps">
        {#each STEPS as step, i (i)}
          <li class="td-step" data-on={typed[i] ? true : undefined} data-done={i < shown || undefined} data-bad={step.bad || undefined}>
            <code class="td-cmd"><span class="td-dim">$</span> {typed[i]}{#if typing === i}<Caret />{/if}</code>
            <span class="td-out">{step.out}</span>
            <span class="td-note">
              <Icon icon={step.bad ? icons.error : icons.check} size={14} />
              {t(step.note)}
            </span>
          </li>
        {/each}
      </ol>
    </Window>
  {/snippet}
</Stage>
