// One rule for a widget's props (docs/FOUNDATION.md §5.4, Layout of a post, item 3). A widget's schema stands beside
// its component, in valibot, as <Name>.props.ts exporting `schema`, which the component reads its props' type from.
// The build runs each schema through Vite's module runner, turns it into JSON Schema and writes every one, by its
// widget's name, into one file at the build's root beside the manifest, which internal/assets reads to hold every
// post's props to as a release loads. The manifest names no such file, so the site never serves it.
import { glob } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { toJsonSchema, type JsonSchema, type OverrideSchemaContext } from "@valibot/to-json-schema";
import type { GenericSchema } from "valibot";
import { runnerImport, type Plugin, type Rollup } from "vite";

// The file the schemas are written into at the build's root, by the name internal/assets reads it with
const PROPS_SCHEMA_FILE = "props.schema.json";
// The site's widgets' folder, where the site's entry finds each widget's code by a glob of its own
const WIDGETS_FOLDER = fileURLToPath(new URL("../src/widgets", import.meta.url));
// A widget's component's file, and its schema's beside it, by their endings
const COMPONENT_FILES = "*.svelte";
const COMPONENT_ENDING = ".svelte";
const SCHEMA_ENDING = ".props.ts";
// How the file is written: indented as Vite writes the manifest, and ended with a line's end
const INDENT = 2;
const LINE_END = "\n";
// The draft of JSON Schema written, the current one, which the Go load's validator reads off its $schema
const TARGET = "draft-2020-12";
// valibot's one kind of object that refuses a key it does not name, the one the build takes: every other lets a typo
// in a post's props reach the widget unseen. Every kind of object holds its keys' schemas as its entries
const STRICT_OBJECT = "strict_object";
const OBJECT_ENTRIES = "entries";

/**
 * A widget of a folder: its component's name, the files its schema is read from, its own and every module it imports,
 * and the schema of its props as JSON Schema.
 */
interface WidgetSchema {
  name: string;
  files: string[];
  schema: JsonSchema;
}

/** What a widget's schema's file exports, as the build reads it. */
interface SchemaModule {
  schema?: GenericSchema;
}

/**
 * Every widget a folder holds, by its component, with its props' schema as JSON Schema, in its names' order; none where
 * there is no folder. Fails on a component with no schema's file beside it, or a schema the build refuses.
 */
export async function widgetPropsSchemas(folder: string): Promise<WidgetSchema[]> {
  const components = await Array.fromAsync(glob(COMPONENT_FILES, { cwd: folder }));
  return Promise.all(
    components.sort().map((component) => widgetSchemaOf(folder, path.basename(component, COMPONENT_ENDING))),
  );
}

/** A widget's props' schema, read from its file beside its component, as JSON Schema. */
async function widgetSchemaOf(folder: string, name: string): Promise<WidgetSchema> {
  const file = path.join(folder, name + SCHEMA_ENDING);
  const { module, dependencies } = await runnerImport<SchemaModule>(file);
  if (module.schema === undefined) {
    throw new Error(`${file} exports no schema of the widget's props`);
  }
  try {
    return { name, files: [file, ...dependencies], schema: jsonSchemaOf(module.schema) };
  } catch (cause) {
    throw new Error(`${file} holds a schema the build refuses`, { cause });
  }
}

/**
 * A widget's props' schema as JSON Schema, refused where a post's typo would pass it or the props would reach the
 * component other than as the schema types them: props that are no strict object, an object in them that is none, a
 * prop taken by default (refusedPart), and what JSON Schema cannot say, toJsonSchema's own refusals.
 */
export function jsonSchemaOf(schema: GenericSchema): JsonSchema {
  if (schema.type !== STRICT_OBJECT) {
    throw new Error(`a widget's props are a ${STRICT_OBJECT}, never a ${schema.type}`);
  }
  return toJsonSchema(schema, { target: TARGET, overrideSchema: refusedPart });
}

/**
 * Refuses a part of a widget's props' schema the build does not take, and leaves every other to toJsonSchema: an
 * object of any kind but the strict, and a default, which JSON Schema writes as a note no reader applies, so the
 * schema would say what the page never does: the page hands the component its props as the post writes them.
 */
function refusedPart({ valibotSchema }: OverrideSchemaContext): undefined {
  if (OBJECT_ENTRIES in valibotSchema && valibotSchema.type !== STRICT_OBJECT) {
    throw new Error(`an object of a widget's props is a ${STRICT_OBJECT}, never a ${valibotSchema.type}`);
  }
  if ("default" in valibotSchema && valibotSchema.default !== undefined) {
    throw new Error("a widget's prop takes no default in its schema, which no reader of the props applies");
  }
  return undefined;
}

/** The file of the widgets' props' schemas, as the build writes it: each widget's schema by its name. */
export function propsSchemaText(widgets: readonly WidgetSchema[]): string {
  const schemas = Object.fromEntries(widgets.map(({ name, schema }) => [name, schema]));
  return JSON.stringify(schemas, null, INDENT) + LINE_END;
}

/**
 * Writes the site's widgets' props' schemas into the build, and watches every file each is read from, which no chunk
 * imports, since a component reads its schema's type alone, so a change to one builds again.
 */
async function emitPropsSchemas(this: Rollup.PluginContext): Promise<void> {
  const widgets = await widgetPropsSchemas(WIDGETS_FOLDER);
  for (const file of widgets.flatMap(({ files }) => files)) {
    this.addWatchFile(file);
  }
  this.emitFile({ type: "asset", fileName: PROPS_SCHEMA_FILE, source: propsSchemaText(widgets) });
}

/** The plugin that writes the widgets' props' schemas, in a build alone: a test builds no file. */
export const WIDGET_PROPS: Plugin = { name: "paranin:widget-props", apply: "build", buildStart: emitPropsSchemas };
