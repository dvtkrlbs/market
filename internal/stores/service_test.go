package stores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/dvtkrlbs/market/internal/database"
	"github.com/dvtkrlbs/market/internal/ent"
	"github.com/dvtkrlbs/market/internal/ent/storeadmin"
	"github.com/dvtkrlbs/market/migrations"
)

func TestNameValidation(t *testing.T) {
	s := New(nil)
	for _, name := range []string{"", " \t\n "} {
		if _, err := s.Create(context.Background(), 1, name); !errors.Is(err, ErrNameRequired) {
			t.Fatalf("Create(%q): %v", name, err)
		}
		if _, err := s.Update(context.Background(), 1, 1, Update{Name: &name}); !errors.Is(err, ErrNameRequired) {
			t.Fatalf("Update(%q): %v", name, err)
		}
	}
	if _, err := s.Update(context.Background(), 1, 1, Update{}); !errors.Is(err, ErrEmptyUpdate) {
		t.Fatalf("empty update: %v", err)
	}
}

// TEST_DATABASE_URL must point at a PostgreSQL database where the test may
// create a temporary schema. No existing tables or data are modified.
func TestStoreCRUDIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := fmt.Sprintf("stores_test_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	scopedDB, err := sql.Open("pgx", parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer scopedDB.Close()
	for _, migration := range []string{"20261007000000_init.up.sql", "20261007000001_auth.up.sql", "20261010000000_stores.up.sql"} {
		contents, err := migrations.FS.ReadFile(migration)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := scopedDB.ExecContext(ctx, string(contents)); err != nil {
			t.Fatal(err)
		}
	}
	client, err := database.Open(parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	first := client.User.Create().SetEmail("first@example.com").SetPasswordHash("test").SaveX(ctx)
	second := client.User.Create().SetEmail("second@example.com").SetPasswordHash("test").SaveX(ctx)
	s := New(client)
	created, err := s.Create(ctx, first.ID, "  Global shop  ")
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "Global shop" || created.Approved {
		t.Fatalf("created = %+v", created)
	}
	other, err := s.Create(ctx, second.ID, "Other shop")
	if err != nil {
		t.Fatal(err)
	}
	checkList := func(userID int, mode string, ids ...int) {
		t.Helper()
		items, err := s.List(ctx, userID, mode)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != len(ids) {
			t.Fatalf("List(%d, %q) = %v, want IDs %v", userID, mode, items, ids)
		}
		for i, id := range ids {
			if items[i].ID != id {
				t.Fatalf("List IDs = %v, want %v", items, ids)
			}
		}
	}
	checkList(first.ID, "shop")
	checkList(first.ID, "manage", created.ID)
	checkList(second.ID, "manage", other.ID)
	checkList(first.ID, "admin", created.ID)
	checkList(second.ID, "admin", other.ID)
	if _, err := s.List(ctx, first.ID, "invalid"); !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("invalid mode: %v", err)
	}
	if _, err := s.Get(ctx, second.ID, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unapproved store visible: %v", err)
	}
	if _, err := s.Get(ctx, first.ID, created.ID); err != nil {
		t.Fatal(err)
	}
	approved := true
	if _, err := s.Update(ctx, second.ID, created.ID, Update{Approved: &approved}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("non-admin update: %v", err)
	}
	if err := s.Delete(ctx, second.ID, created.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("non-admin delete: %v", err)
	}
	name := "Renamed shop"
	updated, err := s.Update(ctx, first.ID, created.ID, Update{Name: &name, Approved: &approved})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != name || !updated.Approved {
		t.Fatalf("updated = %+v", updated)
	}
	checkList(first.ID, "shop", created.ID)
	checkList(second.ID, "", created.ID)
	if _, err := s.Get(ctx, second.ID, created.ID); err != nil {
		t.Fatal(err)
	}
	// More than one user can administer the same global store.
	client.StoreAdmin.Create().SetUserID(second.ID).SetStoreID(created.ID).SaveX(ctx)
	checkList(second.ID, "manage", created.ID, other.ID)
	if _, err := client.StoreAdmin.Create().SetUserID(second.ID).SetStoreID(created.ID).Save(ctx); !ent.IsConstraintError(err) {
		t.Fatalf("duplicate membership: %v", err)
	}
	approved = false
	if _, err := s.Update(ctx, second.ID, created.ID, Update{Approved: &approved}); err != nil {
		t.Fatal(err)
	}
	checkList(first.ID, "shop")
	approved = true
	if _, err := s.Update(ctx, second.ID, created.ID, Update{Approved: &approved}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, second.ID, created.ID); err != nil {
		t.Fatal(err)
	}
	if count := client.StoreAdmin.Query().Where(storeadmin.StoreIDEQ(created.ID)).CountX(ctx); count != 2 {
		t.Fatalf("memberships not retained: %d", count)
	}
	if item := client.Store.GetX(ctx, created.ID); item.DeletedAt == nil {
		t.Fatal("store was not soft-deleted")
	}
	checkList(first.ID, "manage")
	checkList(second.ID, "manage", other.ID)
	checkList(second.ID, "shop")
	checkList(first.ID, "shop")
	checkList(first.ID, "")
	checkList(first.ID, "admin", created.ID)
	checkList(second.ID, "admin", created.ID, other.ID)
	if _, err := s.Get(ctx, first.ID, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted store: %v", err)
	}
	if _, err := s.Update(ctx, first.ID, created.ID, Update{Name: &name}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("update missing: %v", err)
	}
	if err := s.Delete(ctx, first.ID, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete missing: %v", err)
	}
	// Membership failure rolls back store creation as well.
	before := client.Store.Query().CountX(ctx)
	if _, err := s.Create(ctx, second.ID+first.ID+1000, "Orphan"); err == nil {
		t.Fatal("expected FK failure")
	}
	if after := client.Store.Query().CountX(ctx); before != after {
		t.Fatalf("orphan store created: before %d, after %d", before, after)
	}
	// Check the matching down migration too.
	down, err := migrations.FS.ReadFile("20261010000000_stores.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scopedDB.ExecContext(ctx, string(down)); err != nil {
		t.Fatal(err)
	}
}
