// What every stylesheet of the page is held to beyond stylelint's standard rules. A colour lives in the design
// tokens alone (.claude/rules/code.md): src/tokens/tokens.css, which scripts/tokens writes from design/tokens.yaml,
// so no other stylesheet writes one, by its value, its name, a system's colour or a function that makes one; it
// names a token.

// CSS's system colours, the browser's own, which are no tokens either: a property that takes a colour, every one
// named for a colour (stop-color, -webkit-tap-highlight-color) and every custom property among them, since one may
// carry a colour to another, refuses a value naming one, in any case, as a word of its own; so a custom property's
// name in a value (var(--field-width)) or a word joined to another by a hyphen passes, and a property that takes
// none (grid-area: mark) is never read
const COLOUR_PROPERTIES =
  /color|^(--|background|border|outline|fill|stroke|text-decoration|column-rule|box-shadow|text-shadow|text-emphasis)/;
const SYSTEM_COLOURS =
  /(?<![\w-])(AccentColor|AccentColorText|ActiveText|ButtonBorder|ButtonFace|ButtonText|Canvas|CanvasText|Field|FieldText|GrayText|Highlight|HighlightText|LinkText|Mark|MarkText|SelectedItem|SelectedItemText|VisitedText)(?![\w-])/i;

const COLOUR_FUNCTIONS = [
  "rgb",
  "rgba",
  "hsl",
  "hsla",
  "hwb",
  "lab",
  "lch",
  "oklab",
  "oklch",
  "color",
  "color-mix",
  "light-dark",
];

export default {
  extends: ["stylelint-config-standard"],
  // Written by scripts/tokens, whose hook holds it to its source
  ignoreFiles: ["src/tokens/tokens.css"],
  rules: {
    "color-no-hex": true,
    "color-named": "never",
    "function-disallowed-list": COLOUR_FUNCTIONS,
    "declaration-property-value-disallowed-list": {
      [COLOUR_PROPERTIES.toString()]: [SYSTEM_COLOURS],
    },
  },
};
