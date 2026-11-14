<script lang="ts">
  import { Button, icons } from "@facile/muse";
  import { REPO_URL } from "$lib/site";
</script>

<section aria-labelledby="cta-title" class="flex flex-col items-start gap-5 rounded-fc-lg bg-fc-component p-6 sm:items-center sm:p-12 sm:text-center">
  <h2 id="cta-title" class="text-fc-xl font-bold tracking-tight text-fc-fg sm:text-3xl">Break things somewhere else</h2>
  <p class="max-w-md text-fc-md text-fc-fg-muted">One command to install, one to create a sandbox.</p>
  <div class="flex flex-wrap gap-3">
    <Button variant="primary" size="lg" href="#download" icon={icons.download}>Install</Button>
    <Button variant="secondary" size="lg" href={REPO_URL} target="_blank" rel="noreferrer" icon="solar:code-square-linear">
      GitHub
    </Button>
  </div>
</section>
