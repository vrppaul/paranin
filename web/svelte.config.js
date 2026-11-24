// How every Svelte component of the page is compiled, read by the build's plugin, svelte-check and eslint alike: in
// runes mode, Svelte 5's own, so a component written in the older form is refused rather than compiled another way.
// Its style is left to the plugin, which writes it to a file the build emits beside its script, never into the page,
// whose policy refuses an inline style (docs/WEB.md, "Security headers").
export default {
  compilerOptions: {
    runes: true,
  },
};
