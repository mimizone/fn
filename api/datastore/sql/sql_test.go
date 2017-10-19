package sql

import (
	"net/url"
	"os"
	"testing"

	"github.com/fnproject/fn/api/datastore/internal/datastoretest"
	"github.com/fnproject/fn/api/datastore/internal/datastoreutil"
	"github.com/fnproject/fn/api/models"
)

// since New with fresh dbs skips all migrations:
// * open a fresh db on latest version
// * run all down migrations
// * run all up migrations
// [ then run tests against that db ]
func newWithMigrations(url *url.URL) (models.Datastore, error) {
	ds, err := New(url)
	if err != nil {
		return nil, err
	}

	m, err := migrator(url.String())
	if err != nil {
		return nil, err
	}

	err = m.Down()
	if err != nil {
		return nil, err
	}

	// go through New, to ensure our Up logic works in there...
	ds, err = New(url)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func TestDatastore(t *testing.T) {
	defer os.RemoveAll("sqlite_test_dir")
	u, err := url.Parse("sqlite3://sqlite_test_dir")
	if err != nil {
		t.Fatal(err)
	}
	f := func(t *testing.T) models.Datastore {
		os.RemoveAll("sqlite_test_dir")
		ds, err := New(u)
		if err != nil {
			t.Fatal(err)
		}
		// we don't want to test the validator, really
		return datastoreutil.NewValidator(ds)
	}
	datastoretest.Test(t, f)

	f = func(t *testing.T) models.Datastore {
		os.RemoveAll("sqlite_test_dir")
		ds, err := newWithMigrations(u)
		if err != nil {
			t.Fatal(err)
		}

		// we don't want to test the validator, really
		return datastoreutil.NewValidator(ds)
	}

	// test migrations work
	datastoretest.Test(t, f)

	// if being run from test script (CI) poke around for pg and mysql containers
	// to run tests against them too

	if pg := os.Getenv("POSTGRES_URL"); pg != "" {
		u, err := url.Parse(pg)
		if err != nil {
			t.Fatal(err)
		}

		f := func(t *testing.T) models.Datastore {
			ds, err := New(u)
			if err != nil {
				t.Fatal(err)
			}
			return datastoreutil.NewValidator(ds)
		}

		// test fresh w/o migrations
		datastoretest.Test(t, f)

		f = func(t *testing.T) models.Datastore {
			ds, err := newWithMigrations(u)
			if err != nil {
				t.Fatal(err)
			}
			return ds
		}

		// test that migrations work & things work with them
		datastoretest.Test(t, f)
	}
}
