<script lang="ts">
  import { ForDevelopers, type DeveloperRow } from "@facile/muse";
  import { t } from "$lib/i18n/index.svelte";
  import { REPO_URL } from "$lib/site";

  const rows: DeveloperRow[] = $derived([
    { kind: "cli", text: t("dev.cli"), command: "boite --help" },
    { kind: "source", text: t("dev.source"), href: REPO_URL },
  ]);
</script>

<ForDevelopers id="developers" {rows} description={t("dev.line")} />
