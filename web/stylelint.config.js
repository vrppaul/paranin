// What every stylesheet of the page is held to beyond stylelint's standard rules. A colour lives in the design
// tokens alone (.claude/rules/code.md): src/tokens/tokens.css, which scripts/tokens writes from design/tokens.yaml,
// so no other stylesheet writes one, by its value, its name, a system's colour or a function that makes one; it
// names a token. So does a size, a time and a weight: no other stylesheet writes a unit, nor a weight, a line, an
// opacity or a layer of its own.

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
    // The font shorthand with values sets the face's every longhand again, the page's ligatures off among them, and
    // so do font-variant and all: a text style sets its size, line and weight, and a control takes the page's face
    // whole, by inherit alone. A property's name is read in any case, as CSS reads it
    // A weight, a line's height, an opacity and a layer, numbers with no unit, are tokens too, named by var()
    "declaration-property-value-allowed-list": {
      "/^font(-variant)?$/i": ["inherit"],
      "/^(font-weight|line-height|opacity|z-index)$/i": ["/^var\\(--[\\w-]+\\)$/", "inherit"],
    },
    "property-disallowed-list": ["/^all$/i"],
    // No unit at all: a size or a time is a token of design/tokens.yaml, which tokens.css writes and every other
    // stylesheet names, and a value written here in any unit would be a second source of the design. A unit joins
    // the list with its first user, a cell's ch or a grid's fr, in the chunk that writes it
    "declaration-property-unit-allowed-list": {
      // a drawing is at most as wide as its column: the block letters. The rule takes the first key that matches a
      // property, so each property's own key stands above the catch-all

      "/^max-width$/": ["%"],
      "/.*/": [],
    },
    // A media query holds no width at all yet, and a container query is refused whole, since neither has a user:
    // the first media query brings the page's two widths into media-feature-name-value-allowed-list
    "media-feature-name-unit-allowed-list": {
      "/.*/": [],
    },
    "at-rule-disallowed-list": ["container"],
    // Never a length fixed in pixels or another of CSS's absolute units, in a declaration or a media query, whatever
    // the lists above come to open: a size is in rem, the text size its reader set, or in a cell's ch, so the whole
    // page grows with it
    "unit-disallowed-list": ["px", "pt", "pc", "in", "cm", "mm", "q"],
  },
};
