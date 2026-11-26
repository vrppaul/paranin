// A project's clips (DESIGN.md, Clip), over the markup internal/face/html draws: the browser's own player, with its
// still until a press plays it, which needs no script. Here the player's controls go, and the way to play it, which the
// server draws hidden, shows under the still; a press on the still or on that way gives the controls back and plays the
// clip in its place, its sound on.

// The parts as the server draws them: a clip, its player and the way to play it
const CLIP = '[data-part="clip"]';
const RECORDING = '[data-part="clip-recording"]';
const PLAY = '[data-part="clip-play"]';
// The tab index of an element the script may focus that the keys go round without
const FOCUSED_BY_SCRIPT = -1;

/** Starts each clip of a page: its player's controls taken away and the way to play it shown, until a press plays it. */
export function startClips(page: Document): void {
  for (const clip of page.querySelectorAll(CLIP)) {
    const recording = clip.querySelector<HTMLVideoElement>(RECORDING);
    const play = clip.querySelector<HTMLButtonElement>(PLAY);
    if (recording === null || play === null) {
      continue;
    }
    recording.controls = false;
    play.hidden = false;
    play.addEventListener("click", () => playClip(recording, play));
    recording.addEventListener("click", () => playClip(recording, play), { once: true });
  }
}

/**
 * Plays a clip in its place, its controls given back and its sound on, the way to play it gone and the focus on the
 * player, so a reader who pressed it from the keyboard goes on from there.
 */
function playClip(recording: HTMLVideoElement, play: HTMLButtonElement): void {
  if (recording.controls) {
    return;
  }
  recording.controls = true;
  recording.muted = false;
  play.hidden = true;
  // a player takes the focus in no browser of its own accord, Safari's among them, though its controls do
  recording.tabIndex = FOCUSED_BY_SCRIPT;
  recording.focus();
  void recording.play();
}
