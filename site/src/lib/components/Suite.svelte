<script lang="ts">
  import { WorksWith, type WorksWithApp } from "@facile/muse";
  import { t } from "$lib/i18n/index.svelte";
  import { BULLE_URL, CASIER_URL } from "$lib/site";

  const apps: WorksWithApp[] = $derived([
    { name: "Bulle", icon: "solar:smile-circle-bold-duotone", line: t("with.bulle"), href: BULLE_URL },
    { name: "Casier", icon: "solar:key-minimalistic-square-bold-duotone", line: t("with.casier"), href: CASIER_URL },
  ]);
</script>

<WorksWith id="suite" class="scroll-mt-24" {apps} />
