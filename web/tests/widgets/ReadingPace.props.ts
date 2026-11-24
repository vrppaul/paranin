// The tests' widget's props (ReadingPace.svelte): the schema its component's type is read from, and which the build
// turns into the JSON Schema the Go fixtures' build holds the fixture post's widgets to (internal/assets/testdata)
import { picklist, strictObject } from "valibot";

/** The props of the tests' widget: how the post's reader reads. */
export const schema = strictObject({
  pace: picklist(["slow", "fast"]),
});
