<script lang="ts">
  import type { Snippet } from "svelte";
  import { Button, LandingFooter, LandingHeader, Page, icons } from "@facile/muse";
  import { LOGO } from "$lib/icons";
  import { rememberLocale, t } from "$lib/i18n/index.svelte";
  import { BULLE_URL, CASIER_URL, INSTALL_SCRIPT_URL, REPO_URL } from "$lib/site";
  import ThemeButton from "./ThemeButton.svelte";

  let { children }: { children: Snippet } = $props();

  const links = $derived([
    { label: t("nav.demos"), href: "/#demos" },
    { label: t("nav.pricing"), href: "/#pricing" },
    { label: t("nav.install"), href: "/#download" },
  ]);

  const columns = $derived([
    {
      title: "Boîte",
      links: [...links, { label: t("nav.downloads"), href: "/downloads" }],
    },
    {
      title: t("footer.suite"),
      links: [
        { label: "Bulle", href: BULLE_URL, external: true },
        { label: "Casier", href: CASIER_URL, external: true },
        { label: "facile.studio", href: "https://facile.studio", external: true },
      ],
    },
    {
      title: t("footer.source"),
      links: [
        { label: t("footer.repo"), href: REPO_URL, external: true },
        { label: t("footer.script"), href: INSTALL_SCRIPT_URL, external: true },
        { label: t("footer.license"), href: `${REPO_URL}/blob/main/LICENSE`, external: true },
      ],
    },
  ]);
</script>

<div class="flex min-h-dvh w-full flex-col bg-fc-page text-fc-fg">
  <LandingHeader appName="Boîte" icon={LOGO} homeHref="/" {links} skipHref="#main">
    {#snippet actions()}
      <ThemeButton />
      <Button size="sm" class="h-9" href="/#download" icon={icons.download}>
        {t("nav.install")}
      </Button>
    {/snippet}
  </LandingHeader>

  <main id="main" class="flex-1">
    <Page width="xl" gap="section">
      {@render children()}
    </Page>
  </main>

  <LandingFooter
    appName="Boîte"
    icon={LOGO}
    homeHref="/"
    tagline={t("footer.tagline")}
    {columns}
    legal={{ onLocaleChange: rememberLocale }}
  />
</div>
