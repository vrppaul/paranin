// A project's clip, run over the project's page the server draws: the HTML face's golden page
// (internal/face/html/testdata), which that face's own tests hold to what it writes. jsdom plays nothing, so a test
// watches the player's play be asked for.
import { afterEach, expect, test, vi, type MockInstance } from "vitest";
import { startClips } from "../../src/controllers/clip";
import projectPage from "../../../internal/face/html/testdata/p/tidy-tool.html?raw";

// The clip's parts as the server draws them: the player and the way to play it
const RECORDING = '[data-part="clip-recording"]';
const PLAY = '[data-part="clip-play"]';

/** The document's element a selector finds, which the golden page draws. */
function found(selector: string): HTMLElement {
  const element = document.querySelector<HTMLElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

/** The clip's player, which the golden page draws. */
function player(): HTMLVideoElement {
  const recording = found(RECORDING);
  if (!(recording instanceof HTMLVideoElement)) {
    throw new Error(`the page's ${RECORDING} is no player`);
  }
  return recording;
}

/** The golden page with its clips started, its player's sound off as a browser starts it; the player's play, watched. */
function started(): MockInstance<() => Promise<void>> {
  const parsed = new DOMParser().parseFromString(projectPage, "text/html");
  document.documentElement.innerHTML = parsed.documentElement.innerHTML;
  const played = vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
  player().muted = true;
  startClips(document);
  return played;
}

/** What a reader can tell of the clip: whether the player shows its own controls, the way to play it, and its sound. */
function seen(): { controls: boolean; play: boolean; muted: boolean } {
  return { controls: player().controls, play: !found(PLAY).hidden, muted: player().muted };
}

afterEach(() => {
  vi.restoreAllMocks();
});

test("the script takes the player's own controls away and shows the way to play it under the still", () => {
  const played = started();
  expect({ ...seen(), played: played.mock.calls.length }).toEqual({
    controls: false,
    play: true,
    muted: true,
    played: 0,
  });
});

test.each([
  { name: "the way to play it", pressed: PLAY },
  { name: "the still", pressed: RECORDING },
])("a press on $name plays the clip in its place with its controls and its sound", ({ pressed }) => {
  const played = started();
  found(pressed).click();
  expect({ ...seen(), played: played.mock.calls.length }).toEqual({
    controls: true,
    play: false,
    muted: false,
    played: 1,
  });
});
