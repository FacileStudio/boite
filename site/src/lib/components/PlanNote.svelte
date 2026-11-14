<script lang="ts">
  import { Button, LANDING_HEADING } from "@facile/muse";
  import { t } from "$lib/i18n/index.svelte";
  import { BULLE_URL, SUITE_PRICING_URL } from "$lib/site";
</script>

<section id="pricing" aria-labelledby="pricing-title" class="flex scroll-mt-24 flex-col items-start gap-5 rounded-fc-lg bg-fc-component p-6 sm:p-10">
  <h2 id="pricing-title" class={LANDING_HEADING}>{t("plan.title")}</h2>
  <p class="max-w-2xl text-fc-lg text-fc-fg">{t("plan.note")}</p>
  <p class="text-fc-md text-fc-fg-muted">{t("plan.suite")}</p>
  <div class="flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
    <Button variant="primary" size="lg" href={BULLE_URL} icon="solar:smile-circle-bold-duotone">{t("plan.bulle")}</Button>
    <Button variant="secondary" size="lg" href={SUITE_PRICING_URL}>{t("plan.all")}</Button>
  </div>
</section>
