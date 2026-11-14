<script lang="ts">
  import { Blob } from "@facile/blob";
  import { Icon, icons } from "@facile/muse";
  import { fade } from "svelte/transition";
  import { Bar, Stage, Stream, Window, grow, type Act } from "@facile/mise-en-scene";

  type Step = "hidden" | "running" | "ok";

  const PROMPT = "Run the tests on fix/login.";
  const ANSWER = "All 48 tests pass on fix/login. The code ran in the VM, never on your machine.";
  const CALLS = [
    { args: "git clone -b fix/login app", end: "ok" },
    { args: "bun install", end: "ok" },
    { args: "bun test", end: "ok" },
  ] as const;
  const WORDS = ANSWER.split(" ").length;

  let draft = $state("");
  let sent = $state(false);
  let steps = $state<Step[]>(["hidden", "hidden", "hidden"]);
  let words = $state(0);

  function reset() {
    draft = "";
    sent = false;
    steps = ["hidden", "hidden", "hidden"];
    words = 0;
  }

  function poster() {
    sent = true;
    steps = CALLS.map((c) => c.end);
    words = WORDS;
  }

  async function script(act: Act) {
    await act.wait(800);
    await act.type(PROMPT, (v) => (draft = v), 40);
    await act.wait(350);
    sent = true;
    draft = "";
    for (const [i, call] of CALLS.entries()) {
      await act.wait(350);
      steps[i] = "running";
      await act.wait(1000);
      steps[i] = call.end;
    }
    await act.wait(400);
    await act.count(WORDS, (v) => (words = v), 70);
    await act.wait(2200);
    reset();
    await act.wait(600);
  }
</script>

<Stage
  width={720}
  height={460}
  compact={{ below: 600, width: 358, height: 500 }}
  label="Bulle runs a test suite inside a Boite VM and reports the result."
  {script}
  {reset}
  {poster}
>
  {#snippet children({ compact })}
    <Window
      title="Bulle"
      style={compact ? "top:12px;left:12px;width:334px;height:468px" : "top:20px;left:48px;width:624px;height:412px"}
    >
      <div class="bd-thread">
        <div class="bd-past">
          <span class="bd-bubble-bar"></span>
          <Bar w="78%" />
          <Bar w="54%" />
        </div>
        {#if sent}
          <div class="bd-q" in:grow out:fade={{ duration: 200, delay: 180 }}><p>{PROMPT}</p></div>
          <div class="bd-agent" in:grow out:fade={{ duration: 200, delay: 120 }}>
            <Blob name="Bulle" size={24} animate="always" class="size-6 shrink-0" />
            <span>Bulle</span>
          </div>
        {/if}
        {#if steps[0] !== "hidden"}
          <section class="bd-group" in:grow out:fade={{ duration: 200, delay: 120 }}>
            <header><Icon icon="solar:box-minimalistic-bold-duotone" size={16} /><b>Boite VM</b></header>
            <ul>
              {#each CALLS as call, i (i)}
                {#if steps[i] !== "hidden"}
                  <li in:grow={{ duration: 350 }} data-state={steps[i]}>
                    <span class="bd-mark">
                      {#if steps[i] === "ok"}<Icon icon={icons.check} size={16} />{:else}<span class="bd-spin"></span>{/if}
                    </span>
                    <b>bash</b><code>{call.args}</code>
                  </li>
                {/if}
              {/each}
            </ul>
          </section>
        {/if}
        {#if words > 0}
          <div in:grow={{ duration: 300 }} out:fade={{ duration: 200 }}>
            <Stream class="bd-answer" text={ANSWER} count={words} />
          </div>
        {/if}
      </div>
      <div class="bd-composer">
        <span class:bd-placeholder={!draft}>{draft || "Message Bulle"}</span>
        <span class="bd-send" class:bd-idle={!draft}></span>
      </div>
    </Window>
  {/snippet}
</Stage>

<style>
  .bd-thread {
    display: flex;
    flex: 1;
    flex-direction: column;
    justify-content: flex-end;
    min-height: 0;
    overflow: hidden;
    padding: 0 20px 12px;
    font-size: 14px;
    color: var(--color-fc-fg);
    mask-image: linear-gradient(to bottom, transparent, black 48px);
  }
  .bd-thread > :global(*) {
    flex: none;
  }
  .bd-past {
    display: grid;
    gap: 8px;
    padding-bottom: 8px;
  }
  .bd-bubble-bar {
    justify-self: end;
    width: 40%;
    height: 32px;
    border-radius: 12px;
    background: var(--color-fc-surface);
  }
  .bd-q {
    display: flex;
    justify-content: flex-end;
    padding-top: 16px;
  }
  .bd-q p {
    max-width: 85%;
    margin: 0;
    padding: 8px 14px;
    border-radius: 12px;
    background: var(--color-fc-surface);
  }
  .bd-agent {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-top: 14px;
    font-size: 12px;
    font-weight: 700;
  }
  .bd-group {
    margin-top: 10px;
    overflow: hidden;
    border-radius: 12px;
    background: color-mix(in oklch, var(--color-fc-fg) 5%, transparent);
  }
  .bd-group header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    color: var(--color-fc-fg-muted);
  }
  .bd-group header b {
    color: var(--color-fc-fg);
  }
  .bd-group ul {
    margin: 0;
    padding: 0 0 4px;
    list-style: none;
  }
  .bd-group li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 7px 12px;
    white-space: nowrap;
  }
  .bd-group code {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: var(--font-fc-mono);
    font-size: 12px;
    color: var(--color-fc-fg-muted);
  }
  .bd-mark {
    display: flex;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    color: var(--color-fc-success);
  }
  .bd-spin {
    width: 12px;
    height: 12px;
    border-radius: 999px;
    background: conic-gradient(var(--color-fc-fg-muted), transparent);
    animation: bd-turn 0.9s linear infinite;
  }
  @keyframes bd-turn {
    to {
      transform: rotate(1turn);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .bd-spin {
      animation: none;
    }
  }
  :global(.bd-answer) {
    margin: 0;
    padding-top: 10px;
    line-height: 1.6;
  }
  .bd-composer {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 20px 20px;
    padding: 8px 8px 8px 20px;
    border-radius: 31px;
    background: var(--color-fc-field);
    font-size: 15px;
  }
  .bd-composer > span:first-child {
    flex: 1;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .bd-placeholder {
    color: var(--color-fc-fg-muted);
  }
  .bd-send {
    width: 40px;
    height: 40px;
    border-radius: 999px;
    background: var(--color-fc-accent);
    transition: opacity 0.15s ease;
  }
  .bd-idle {
    opacity: 0.4;
  }
</style>
