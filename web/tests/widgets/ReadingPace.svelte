<!--
  The tests' widget, which the fixture post shows (internal/content/testdata): a widget as the islands mount one, its
  one prop drawn in the window's body, from its edge, styled by the tokens alone, typed by the props' schema beside it
  as the post writes them, which is what it gets. It stands with the tests, so the page's own build holds no widget
  until a post of the site shows one.
-->
<script lang="ts">
  import type { InferInput } from "valibot";
  import type { schema } from "./ReadingPace.props";

  const { pace }: InferInput<typeof schema> = $props();
</script>

<div class="pace" data-part="reading-pace">{pace}</div>

<style>
  .pace {
    color: var(--text);
  }
</style>
