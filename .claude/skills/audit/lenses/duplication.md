# Lens: duplication and abstraction

Start with the tools, then read:

- `golangci-lint run --enable-only dupl,goconst <your packages>` lists copied blocks and repeated literals.
- Then look for what tools miss:
  - the same loop over components in several places;
  - parallel `switch` statements over the same set of kinds;
  - near-identical functions that differ only in a type (a generic fits);
  - repeated error-wrapping or option-building code;
  - test helpers copied between test files.

When you propose an abstraction:

- **Respect the layers** (AGENTS.md rule 2). A shared helper lives where every user may import it, usually a foundation package. Feature packages never import each other.
- **Reuse what exists first.** Check whether a helper already lives somewhere, and name it.
- **Only when it pays.** It must remove code and concepts overall, not just move them. Two similar three-line snippets stay as they are. Say how many lines and places it removes.
- **Content duplication counts too.** Repeated CUE structures that a definition or default would express once.

Report each finding with every place it appears (`file:line`), the proposed shared form and where it lives, and the size.
