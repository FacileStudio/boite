<script lang="ts">
  import { FeatureSplit, type FeatureItem } from "@facile/muse";
  import { t } from "$lib/i18n/index.svelte";

  const features: FeatureItem[] = $derived([
    { icon: "solar:cpu-bolt-bold-duotone", title: t("features.vm.title"), text: t("features.vm.text") },
    { icon: "solar:folder-with-files-bold-duotone", title: t("features.workspace.title"), text: t("features.workspace.text") },
    { icon: "solar:shield-check-bold-duotone", title: t("features.offline.title"), text: t("features.offline.text") },
    { icon: "solar:layers-minimalistic-bold-duotone", title: t("features.distro.title"), text: t("features.distro.text") },
  ]);
</script>

<FeatureSplit id="features" class="scroll-mt-24" title={t("features.title")} description={t("features.line")} {features} />
