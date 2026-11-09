<script lang="ts">
  import { Badge } from "@facile/muse";

  const tools = [
    { name: "Go 1.26", role: "Compiler & tools" },
    { name: "Rust & Cargo", role: "Toolchain & clippy" },
    { name: "Bun & Node 24", role: "Runtimes & npm" },
    { name: "Playwright & Chromium", role: "Headless in /opt/playwright" },
    { name: "Python 3", role: "Virtualenv & pip" },
    { name: "mise", role: "Polyglot manager" },
    { name: "tmux, fzf & jq", role: "CLI power tools" },
    { name: "zsh & Starship", role: "Interactive shell" },
  ];
</script>

<section id="toolchain" aria-labelledby="toolchain-title" class="flex flex-col gap-6">
  <div class="flex flex-col gap-1.5">
    <h2 id="toolchain-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      Baked toolchains
    </h2>
    <p class="max-w-xl text-fc-md text-fc-fg-muted">
      Pre-installed in the base disk image. Zero setup latency.
    </p>
  </div>

  <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
    {#each tools as tool (tool.name)}
      <div class="flex flex-col gap-1 rounded-xl bg-fc-component p-4">
        <span class="font-semibold text-fc-fg">{tool.name}</span>
        <span class="text-fc-xs text-fc-fg-muted">{tool.role}</span>
      </div>
    {/each}
  </div>
</section>
