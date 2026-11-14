<script lang="ts">
  import type { Snippet } from "svelte";
  import { Button, LandingFooter, LandingHeader, Page, icons } from "@facile/muse";
  import { LOGO } from "$lib/icons";
  import { BULLE_URL, CASIER_URL, REPO_URL } from "$lib/site";
  import ThemeButton from "./ThemeButton.svelte";

  let { children }: { children: Snippet } = $props();

  const links = [
    { label: "Bulle", href: "#bulle" },
    { label: "Features", href: "#features" },
    { label: "Install", href: "#download" },
  ];

  const columns = [
    {
      title: "Boite",
      links: [
        { label: "Bulle", href: "#bulle" },
        { label: "Features", href: "#features" },
        { label: "Install", href: "#download" },
      ],
    },
    {
      title: "Facile suite",
      links: [
        { label: "Bulle", href: BULLE_URL, external: true },
        { label: "Casier", href: CASIER_URL, external: true },
        { label: "facile.studio", href: "https://facile.studio", external: true },
      ],
    },
    {
      title: "Open source",
      links: [
        { label: "GitHub repository", href: REPO_URL, external: true },
        { label: "Install script", href: "https://raw.githubusercontent.com/FacileStudio/boite/main/install.sh", external: true },
        { label: "GPL-3.0 license", href: `${REPO_URL}/blob/main/LICENSE`, external: true },
      ],
    },
  ];
</script>

<div class="flex min-h-dvh w-full flex-col bg-fc-page text-fc-fg">
  <LandingHeader appName="Boite" icon={LOGO} homeHref="/" {links} skipHref="#main">
    {#snippet actions()}
      <ThemeButton />
      <Button size="sm" class="h-9" href="#download" icon={icons.download}>
        Install
      </Button>
    {/snippet}
  </LandingHeader>

  <main id="main" class="flex-1">
    <Page width="xl" gap="section">
      {@render children()}
    </Page>
  </main>

  <LandingFooter
    appName="Boite"
    icon={LOGO}
    homeHref="/"
    tagline="Disposable QEMU virtual machines for your code and your AI agents."
    {columns}
  />
</div>
