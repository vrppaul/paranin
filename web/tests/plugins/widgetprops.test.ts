// @vitest-environment node
// The build's writing of the widgets' props' schemas (plugins/widgetprops.ts), run in Node as the build runs it: over
// the tests' widget, whose file the Go fixtures' build holds as this writes it, so the load's tests hold the fixture
// post to the schema the build writes; and the schemas the build refuses, by which a typo in a post would pass.
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { expect, test } from "vitest";
import {
  looseObject,
  number,
  object,
  objectWithRest,
  optional,
  picklist,
  string,
  strictObject,
  type GenericSchema,
} from "valibot";
import { jsonSchemaOf, propsSchemaText, widgetPropsSchemas } from "../../plugins/widgetprops.ts";

// The Go fixtures' build's file of the widgets' props' schemas, from this file, which `vitest -u` writes
const FIXTURE_SCHEMAS = "../../../internal/assets/testdata/build/props.schema.json";
// The tests' widgets' folder, and a folder that holds none
const TEST_WIDGETS = fileURLToPath(new URL("../widgets", import.meta.url));
const NO_FOLDER = fileURLToPath(new URL("../no-widgets", import.meta.url));
// A widget's component a test writes, what it holds, its schema's file, and a schema's file that exports no schema
const BARE_COMPONENT = "Bare.svelte";
const BARE_MARKUP = "<p>bare</p>\n";
const BARE_SCHEMA = "Bare.props.ts";
const NO_SCHEMA_EXPORTED = "export const other = 1;\n";
// The pace the tests' widget takes, and a count a refused schema takes by default
const PACE = picklist(["slow", "fast"]);
const DEFAULT_COUNT = 3;

// How the folder a test writes a widget into is named, in the system's folder of temporary files
const SCRATCH_PREFIX = path.join(tmpdir(), "widgets-");

test("the build writes each widget's props' schema, as the Go fixtures' build holds the tests' widget's", async () => {
  await expect(propsSchemaText(await widgetPropsSchemas(TEST_WIDGETS))).toMatchFileSnapshot(FIXTURE_SCHEMAS);
});

test("a build of no widget writes a file of no schema, which the load still reads", async () => {
  expect(propsSchemaText(await widgetPropsSchemas(NO_FOLDER))).toBe("{}\n");
});

test("a widget whose component has no schema beside it fails the build", async ({ onTestFinished }) => {
  const folder = await mkdtemp(SCRATCH_PREFIX);
  onTestFinished(() => rm(folder, { recursive: true }));
  await writeFile(path.join(folder, BARE_COMPONENT), BARE_MARKUP);
  await expect(widgetPropsSchemas(folder)).rejects.toThrow(BARE_SCHEMA);
});

test("a widget whose schema's file exports no schema fails the build", async ({ onTestFinished }) => {
  const folder = await mkdtemp(SCRATCH_PREFIX);
  onTestFinished(() => rm(folder, { recursive: true }));
  await writeFile(path.join(folder, BARE_COMPONENT), BARE_MARKUP);
  await writeFile(path.join(folder, BARE_SCHEMA), NO_SCHEMA_EXPORTED);
  await expect(widgetPropsSchemas(folder)).rejects.toThrow(BARE_SCHEMA);
});

test.each<{ name: string; schema: GenericSchema }>([
  { name: "props that are no object", schema: string() },
  { name: "props open to a key they do not name", schema: object({ pace: PACE }) },
  { name: "props loose to a key they do not name", schema: looseObject({ pace: PACE }) },
  {
    name: "an object of the props open to a key it does not name",
    schema: strictObject({ pace: object({ pace: PACE }) }),
  },
  {
    name: "an object of the props taking any key its rest's schema takes",
    schema: strictObject({ inner: objectWithRest({ pace: string() }, string()) }),
  },
  { name: "a prop taken by default", schema: strictObject({ count: optional(number(), DEFAULT_COUNT) }) },
])("a schema of $name fails the build", ({ schema }) => {
  expect(() => jsonSchemaOf(schema)).toThrow();
});
