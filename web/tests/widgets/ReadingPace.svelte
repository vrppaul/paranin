<!--
  The tests' widget, which the fixture post shows (internal/content/testdata): a widget as the islands mount one, its
  one prop drawn in the window's body, from its edge, styled by the tokens alone. It stands with the tests, so the
  page's own build holds no widget until a post of the site shows one.
-->
<script lang="ts">
  interface Props {
    /** How the post's reader reads, as the post's block writes it */
    pace: string;
  }

  const { pace }: Props = $props();
</script>

<div class="pace" data-part="reading-pace">{pace}</div>

<style>
  .pace {
    color: var(--text);
  }
</style>
