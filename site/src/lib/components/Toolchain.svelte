<script lang="ts">
  import { Badge } from "@facile/muse";

  const tools = [
    {
      name: "Go 1.26",
      category: "Compiler",
      desc: "Native Go runtime, compiler, and testing tools for backend services.",
    },
    {
      name: "Rust & Cargo",
      category: "Systems",
      desc: "Full rustc and cargo toolchain ready for bare-metal and systems programming.",
    },
    {
      name: "Bun & Node 24",
      category: "JavaScript",
      desc: "High-performance JS runtime, package manager, and full modern Node ecosystem.",
    },
    {
      name: "Playwright & Chromium",
      category: "Automation",
      desc: "Pre-baked browser binaries in /opt/playwright for headless agent tasks and web testing.",
    },
    {
      name: "Python 3",
      category: "Scripting",
      desc: "Modern Python runtime with venv, pip, and standard data-handling libraries.",
    },
    {
      name: "mise",
      category: "Runtimes",
      desc: "Polyglot environment manager to switch language versions per directory.",
    },
    {
      name: "CLI Power Suite",
      category: "Shell",
      desc: "fzf, jq, tmux, and Starship prompt with zsh for fast keyboard ergonomics.",
    },
    {
      name: "boite User",
      category: "Security",
      desc: "Non-root developer user with passwordless sudo and automated SSH key injection.",
    },
  ];
</script>

<section id="toolchain" aria-labelledby="toolchain-title" class="flex flex-col gap-6">
  <div class="flex flex-col gap-2">
    <h2 id="toolchain-title" class="text-fc-xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
      Baked-In Toolchains
    </h2>
    <p class="max-w-2xl text-fc-md text-fc-fg-muted">
      Everything needed for development and AI agents is pre-installed in the base image. Zero setup delay.
    </p>
  </div>

  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
    {#each tools as tool (tool.name)}
      <div class="flex flex-col gap-3 rounded-2xl bg-fc-component p-5 sm:p-6">
        <div class="flex items-center justify-between">
          <span class="text-fc-md font-semibold text-fc-fg">{tool.name}</span>
          <Badge tone="neutral">{tool.category}</Badge>
        </div>
        <p class="text-fc-xs leading-relaxed text-fc-fg-muted sm:text-fc-sm">
          {tool.desc}
        </p>
      </div>
    {/each}
  </div>
</section>
