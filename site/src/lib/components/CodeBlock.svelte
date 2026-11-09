<script lang="ts">
  import Prism from "prismjs";
  import "prismjs/components/prism-typescript.js";
  import "prismjs/components/prism-bash.js";
  import "prismjs/components/prism-json.js";
  import "prismjs/components/prism-yaml.js";
  import "prismjs/components/prism-http.js";
  import CopyButton from "./CopyButton.svelte";

  let {
    code,
    label,
    lang = "bash",
    class: className = "",
  }: {
    code: string;
    label?: string;
    lang?: string;
    class?: string;
  } = $props();

  const grammar = $derived(Prism.languages[lang] ?? Prism.languages.bash);
  const highlighted = $derived(Prism.highlight(code, grammar, lang));
</script>

<div class="group relative overflow-hidden rounded-xl bg-fc-surface text-fc-fg {className}">
  {#if label}
    <div class="flex items-center justify-between bg-fc-field/60 px-4 py-2 font-mono text-fc-xs text-fc-fg-muted">
      <span>{label}</span>
      <CopyButton text={code} label="Copy code" />
    </div>
  {:else}
    <div class="absolute right-2 top-2 z-10">
      <CopyButton text={code} label="Copy code" />
    </div>
  {/if}
  <pre class="overflow-x-auto p-4 font-mono text-fc-xs leading-relaxed sm:text-fc-sm"><code>{@html highlighted}</code></pre>
</div>
