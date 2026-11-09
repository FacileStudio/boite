<script lang="ts">
  import { Icon, icons } from "@facile/muse";

  let { text, label = "Copy to clipboard" }: { text: string; label?: string } = $props();

  let copied = $state(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(text);
      copied = true;
      setTimeout(() => (copied = false), 1800);
    } catch {}
  }
</script>

<button
  type="button"
  onclick={copy}
  aria-label={label}
  class="inline-flex size-8 shrink-0 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-fg-muted transition-colors hover:bg-fc-field hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
>
  <Icon icon={copied ? icons.check : icons.copy} size={15} />
  <span class="sr-only" role="status">{copied ? "Copied" : ""}</span>
</button>
