# Lens: the content guide

`docs/content-guide.md` teaches the owner, and agents, how to make content for the engine in CUE. It covers what the schema can't say on its own:

- how `content/` is organized;
- IDs and namespaces;
- text IDs;
- how rules and formulas work;
- short recipes, such as adding a creature or a form;
- how to check the work: `go tool task content`, playing, and the harness.

The field-by-field reference isn't in the guide. It's the schema, which `go tool task schema` generates from the Go content structs, so their doc comments are the reference.

**If the guide doesn't exist,** say so. `/audit` files a task to write it.

**If it exists,** check it against the content and code as they are now:

- **Every file, ID, field, command and rule it names still exists,** and does what the guide says.
- **Examples point at real files in `content/`,** rather than copying CUE into the guide, so the content check keeps them valid. Flag any copied CUE.
- **Its terms match `docs/glossary.md`.**
- **It covers every kind of content the schema defines.** Flag kinds added since it was written.
- **It stays short.** It teaches the patterns and links to the files, and doesn't restate the schema.

Report each gap with the guide's section and what to change.
