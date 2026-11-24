// What a browser has and jsdom lacks, which Svelte's reduced motion and pixel density read as their modules load: the
// system's answer to every media query, which matches none, and an observer of an element's size, which hears it as a
// test lays it out (tests/browser.ts). A test file whose modules read either imports this first, before them.
import { FakeSystem, LayoutObserver } from "./browser";

window.matchMedia = () => new FakeSystem(false) as unknown as MediaQueryList;
window.ResizeObserver = LayoutObserver as unknown as typeof ResizeObserver;
