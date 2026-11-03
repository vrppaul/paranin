// What the page's TypeScript is held to beyond the type checker: typescript-eslint's strict rules over the types,
// the layers of src/ (docs/FOUNDATION.md §5.3), and the rules for any code (.claude/rules/code.md) a linter can
// hold. Taken in part from claude-review's frontend/eslint.config.js: its layers, its names, its nested helpers and
// its lookups by id; what it holds of Svelte, Tailwind and its own library comes here with the parts that need it.
// typescript-eslint takes TypeScript below 6.1 alone, so web/package.json pins TypeScript at 6.0.
import eslint from "@eslint/js";
import tseslint from "typescript-eslint";

// What each part of src/ may build on. A part importing one it is not listed with fails here: nothing reaches up
// into what is built on it. A part comes into the list with its first module
const LAYERS = {
  // The scripts a page runs: the stylesheet it is drawn in, and the controllers over the server's markup
  entries: ["controllers", "ui"],
  // Behaviour over the markup the server draws, each over one part of it
  controllers: [],
};
// Every part of src/, those that hold no script among them: a stylesheet is reached by an import as a module is
const PARTS = [...Object.keys(LAYERS), "ui", "tokens"];

/** The imports a part may not use: every part it is not listed as building on, and any path climbing out of src/. */
function outOfReach(part) {
  const others = PARTS.filter((other) => other !== part && !LAYERS[part].includes(other));
  return [...others.map((other) => `../${other}/*`), "../../*"];
}

// A part is named one way, `../entries/site`: `./../entries/site`, `../controllers/../entries/site` and
// `.//../tokens` reach another part by a spelling no pattern of a part would refuse, so a path with an empty, `.`
// or `..` step past its first is refused everywhere
const PLAINLY_NAMED = {
  regex: String.raw`^\./\.\.|//|/\.\.?(/|$)`,
  message: "Name a module plainly, with no empty, . or .. step past the path's first",
};

const layerRules = Object.keys(LAYERS).map((part) => ({
  files: [`src/${part}/**`],
  rules: {
    "no-restricted-imports": [
      "error",
      {
        patterns: [
          { group: outOfReach(part), message: `${part}/ builds only on ${LAYERS[part].join(", ") || "nothing"}` },
          PLAINLY_NAMED,
        ],
      },
    ],
  },
}));

// The fewest letters a name may have: `row`, `end`. The rule counts a name's letters alone, so an abbreviation of
// three letters or more (`btn`, `cfg`) passes it, which scripts/coderules refuses in Go; it is the reviewer's to find
const WHOLE_WORD_LETTERS = 3;

// A helper is a function of its module, with a name and parameters of its own, not a closure in the body of another.
// A one-line lambda handed to a call is still fine; one with a body of statements is a helper wherever it is inside
// another function, and so is a function declared there, a function expression and a class
const NESTED_HELPER = "A helper with logic is a function of the module, not one inside another.";
const NO_NESTED_HELPERS = [
  {
    selector:
      ":function :matches(ArrowFunctionExpression[expression=false], FunctionExpression, FunctionDeclaration, ClassDeclaration, ClassExpression)",
    message: NESTED_HELPER,
  },
];

// An element is found by what it is, a data-part the server's markup gives it, never by an id, which is page-wide.
// A selector is a named constant, so every string that names an id as a selector is refused wherever it stands:
// `#name` at a selector's start or after a combinator, a comma or a bracket, and `[id…]`
const BY_PAGE_WIDE_ID = "Find an element by its data-part, never by a page-wide id.";
const NAMES_AN_ID = String.raw`/(^|[\s>+~,(])#[A-Za-z]|\[id\b/`;
const NO_PAGE_WIDE_ID_LOOKUPS = [
  { selector: "CallExpression[callee.property.name='getElementById']", message: BY_PAGE_WIDE_ID },
  { selector: `Literal[value=${NAMES_AN_ID}]`, message: BY_PAGE_WIDE_ID },
  { selector: `TemplateElement[value.raw=${NAMES_AN_ID}]`, message: BY_PAGE_WIDE_ID },
];

// A page's colour scheme is chosen by an attribute on its root, which ui.css turns into color-scheme: the build
// writes each token's light-dark() pair as variables the stylesheet's rules switch, and a scheme set as a style
// passes them by, so a script never names one, as a property, a key or in a style's text
const SCHEME_BY_ATTRIBUTE =
  "Choose the colour scheme by the root's data-theme, never as a style (src/controllers/theme.ts).";
// The property in any spelling, but not as the end of a media feature, `(prefers-color-scheme: dark)`, which asks
// the system's scheme and sets none
const NAMES_A_SCHEME = "/(^|[^-])color-?scheme/i";
const NO_STYLED_SCHEME = [
  { selector: "MemberExpression[property.name='colorScheme']", message: SCHEME_BY_ATTRIBUTE },
  { selector: `Literal[value=${NAMES_A_SCHEME}]`, message: SCHEME_BY_ATTRIBUTE },
  { selector: `TemplateElement[value.raw=${NAMES_A_SCHEME}]`, message: SCHEME_BY_ATTRIBUTE },
];

// No state at a module's level (.claude/rules/code.md): what changes is held by whoever is handed it, so a variable
// a module declares outside every function, exported or not, is a const. The rule reads the keyword alone: a const
// bound to a value that changes, a Map or an array filled later, passes it, and is the reviewer's to find
const STATE_HELD_BY_ITS_USER = "Hold what changes where it is handed, never in a variable of the module.";
const NO_MODULE_STATE = [
  { selector: "Program > VariableDeclaration[kind!='const']", message: STATE_HELD_BY_ITS_USER },
  {
    selector: "Program > ExportNamedDeclaration > VariableDeclaration[kind!='const']",
    message: STATE_HELD_BY_ITS_USER,
  },
];

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  {
    // The rules that read types read them from the page's own TypeScript program, tsconfig.json's
    files: ["**/*.ts"],
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    // The configurations are JavaScript no TypeScript program holds
    files: ["*.config.js"],
    ...tseslint.configs.disableTypeChecked,
  },
  ...layerRules,
  {
    files: ["src/**/*.ts", "tests/**/*.ts"],
    rules: {
      "id-length": ["error", { min: WHOLE_WORD_LETTERS, properties: "never" }],
      "no-restricted-syntax": [
        "error",
        ...NO_NESTED_HELPERS,
        ...NO_PAGE_WIDE_ID_LOOKUPS,
        ...NO_STYLED_SCHEME,
        ...NO_MODULE_STATE,
      ],
      // A handler is one expression here, a body of statements being a helper of its own, and what it hands back is
      // nobody's to read
      "@typescript-eslint/no-confusing-void-expression": ["error", { ignoreArrowShorthand: true }],
      // No number with a meaning in an expression: it is a named constant, with a comment saying what it is
      "@typescript-eslint/no-magic-numbers": ["error", { ignore: [0, 1, -1], ignoreArrayIndexes: true }],
    },
  },
);
