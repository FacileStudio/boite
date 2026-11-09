<script lang="ts">
  import CopyButton from "./CopyButton.svelte";

  let {
    code,
    label,
    class: className = "",
  }: {
    code: string;
    label?: string;
    class?: string;
  } = $props();
</script>

<div class="group relative overflow-hidden rounded-xl bg-fc-surface text-fc-fg {className}">
  {#if label}
    <div class="flex items-center justify-between bg-fc-field/60 px-4 py-2 text-fc-xs text-fc-fg-muted font-mono">
      <span>{label}</span>
      <CopyButton text={code} label="Copy code" />
    </div>
  {:else}
    <div class="absolute top-2 right-2 z-10">
      <CopyButton text={code} label="Copy code" />
    </div>
  {/if}
  <pre class="overflow-x-auto p-4 font-mono text-fc-xs leading-relaxed sm:text-fc-sm"><code>{code}</code></pre>
</div>
