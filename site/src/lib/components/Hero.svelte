<script lang="ts">
  import { Button, icons } from "@facile/muse";
  import { t } from "$lib/i18n/index.svelte";
  import { BULLE_URL } from "$lib/site";
  import HeroDemo from "./HeroDemo.svelte";
</script>

<section class="grid grid-cols-1 items-center gap-8 pt-4 lg:grid-cols-12 lg:gap-12">
  <div class="flex flex-col gap-5 lg:col-span-5">
    <h1 class="font-fc-title text-4xl font-bold tracking-tight text-balance text-fc-fg sm:text-5xl">{t("hero.title")}</h1>
    <p class="max-w-md text-fc-lg text-fc-fg-muted">{t("hero.subtitle")}</p>
    <div class="flex flex-col gap-3 pt-1 sm:flex-row sm:flex-wrap sm:items-center">
      <Button variant="primary" size="lg" href="#download" icon={icons.download} class="w-full sm:w-auto">
        {t("hero.install")}
      </Button>
      <Button variant="secondary" size="lg" href={BULLE_URL} icon="solar:smile-circle-bold-duotone" class="w-full sm:w-auto">
        {t("hero.bulle")}
      </Button>
    </div>
    <p class="max-w-md text-fc-sm text-fc-fg-muted">{t("hero.note")}</p>
  </div>
  <div class="w-full min-w-0 max-w-[600px] lg:col-span-7">
    <HeroDemo />
  </div>
</section>
