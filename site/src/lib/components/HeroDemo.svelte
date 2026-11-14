<script lang="ts">
  import { Caret, Stage, Terminal, type Act } from "@facile/mise-en-scene";
  import { t } from "$lib/i18n/index.svelte";

  type Kind = "cmd" | "out" | "ok" | "head" | "row" | "gap";
  type Line = { kind: Kind; text: string; value?: string };

  const CREATE: Line[] = [
    { kind: "out", text: "Creating sandbox 'demo'..." },
    { kind: "gap", text: "" },
    { kind: "head", text: "Sandbox Ready" },
    { kind: "row", text: "VM:       ", value: "demo" },
    { kind: "row", text: "Distro:   ", value: "debian" },
    { kind: "row", text: "Workspace:", value: "/workspace (~/code/app)" },
    { kind: "row", text: "Connect:  ", value: "boite run demo" },
    { kind: "gap", text: "" },
  ];

  const SCRIPT: { cmd: string; out: Line[] }[] = [
    { cmd: "boite create demo", out: CREATE },
    { cmd: "boite snapshot demo clean", out: [{ kind: "ok", text: "✓ Snapshot 'clean' recorded for 'demo'" }] },
    { cmd: "boite exec demo sudo rm -rf /usr", out: [] },
    { cmd: "boite rollback demo clean", out: [{ kind: "ok", text: "✓ 'demo' rolled back to snapshot 'clean'" }] },
  ];

  const FIRST: Line[] = [{ kind: "cmd", text: SCRIPT[0].cmd }, ...SCRIPT[0].out];

  let lines = $state<Line[]>([...FIRST]);

  function reset() {
    lines = [...FIRST];
  }

  function poster() {
    lines = SCRIPT.flatMap((step) => [{ kind: "cmd" as const, text: step.cmd }, ...step.out]);
  }

  async function script(act: Act) {
    await act.wait(900);
    for (const step of SCRIPT.slice(1)) {
      lines.push({ kind: "cmd", text: "" });
      const at = lines.length - 1;
      await act.type(step.cmd, (v) => (lines[at].text = v), 55);
      await act.wait(400);
      for (const line of step.out) {
        lines.push(line);
        await act.wait(line.kind === "row" ? 90 : 260);
      }
      await act.wait(900);
    }
    await act.wait(2400);
    reset();
    await act.wait(900);
  }
</script>

<Stage
  width={600}
  height={420}
  compact={{ below: 520, width: 358, height: 400 }}
  label={t("hero.demoLabel")}
  {script}
  {reset}
  {poster}
>
  {#snippet children({ compact })}
    <Terminal style={compact ? "top:12px;left:12px;right:12px;bottom:28px" : "top:20px;left:24px;right:24px;bottom:36px"}>
      <div class="hd-head"><span class="hd-dot"></span>host <span class="hd-dim">qemu + kvm</span></div>
      {#each lines as line, i (i)}
        <div class="hd-line">
          {#if line.kind === "cmd"}
            <span class="hd-dim">$</span> {line.text}{#if i === lines.length - 1}<Caret />{/if}
          {:else if line.kind === "head"}
            <span class="hd-title">{line.text}</span>
          {:else if line.kind === "row"}
            <span class="hd-label">{line.text}</span>  <span class="hd-dim">{line.value}</span>
          {:else if line.kind === "ok"}
            <span class="hd-ok">{line.text}</span>
          {:else}
            <span class="hd-dim">{line.text}</span>
          {/if}
        </div>
      {/each}
      {#if lines.at(-1)?.kind !== "cmd"}
        <div class="hd-line"><span class="hd-dim">$</span> <Caret /></div>
      {/if}
    </Terminal>
  {/snippet}
</Stage>

<style>
  .hd-head {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }
  .hd-dot {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: var(--color-fc-success);
  }
  .hd-line {
    min-height: 20px;
    white-space: pre;
  }
  .hd-dim {
    color: var(--color-fc-on-inverse-muted);
  }
  .hd-label {
    color: color-mix(in oklab, var(--color-fc-accent), var(--color-fc-on-inverse) 40%);
  }
  .hd-title {
    font-weight: 700;
  }
  .hd-ok {
    color: color-mix(in oklab, var(--color-fc-success), var(--color-fc-on-inverse) 35%);
  }
</style>
