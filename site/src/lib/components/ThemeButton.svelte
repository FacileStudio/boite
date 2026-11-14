<script lang="ts">
  import { browser } from "$app/environment";
  import { Icon, IconButton, icons } from "@facile/muse";
  import { t } from "$lib/i18n/index.svelte";

  let theme = $state<"light" | "dark">(
    browser && document.documentElement.classList.contains("light") ? "light" : "dark",
  );

  function toggle() {
    theme = theme === "dark" ? "light" : "dark";
    document.documentElement.classList.toggle("dark", theme === "dark");
    document.documentElement.classList.toggle("light", theme === "light");
    try {
      localStorage.setItem("boite-theme", theme);
    } catch {}
  }
</script>

<IconButton
  variant="ghost"
  class="size-9 bg-fc-surface/70 text-fc-fg hover:bg-fc-surface focus-visible:outline-fc-ring"
  aria-label={theme === "dark" ? t("theme.toLight") : t("theme.toDark")}
  onclick={toggle}
>
  <Icon icon={theme === "dark" ? icons.sun : icons.moon} size={16} />
</IconButton>
