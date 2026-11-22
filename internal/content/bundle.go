package content

// LoadBundle reads and schema-checks the CUE package in dir like [Load], but returns the decoded
// bundle without building a registry. Tools and tests use it to adjust content before [New].
func LoadBundle(dir string) (Bundle, error) { return loadBundle(dir) }

// New validates a bundle, compiles its formulas and text, and builds the registry.
func New(bundle Bundle) (*Registry, error) { return newRegistry(bundle) }
