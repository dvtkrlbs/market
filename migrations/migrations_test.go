package migrations

import "testing"

func TestInitialMigrationIsEmbedded(t *testing.T) {
	for _, name := range []string{
		"20261007000000_init.up.sql",
		"20261007000000_init.down.sql",
	} {
		if _, err := FS.ReadFile(name); err != nil {
			t.Fatalf("read embedded migration %q: %v", name, err)
		}
	}
}
