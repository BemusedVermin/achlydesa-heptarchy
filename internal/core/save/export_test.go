package save

import "io"

// ReadWith reads with a migration chain of the test's own, so a test can upgrade from a version
// that production no longer has.
func ReadWith(r io.Reader, current int, chain map[int]Migration) (Snapshot, error) {
	return read(r, current, chain)
}
