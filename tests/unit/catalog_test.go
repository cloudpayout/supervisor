package unit

import (
	"testing"

	"github.com/cloudpayout/supervisor/internal/casino"
)

func TestCatalogOneWin(t *testing.T) {
	got, err := casino.Lookup("onewin")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "onewin" || got.Page != "https://1win.com/" || got.Game != "aviator" {
		t.Fatalf("%+v", got)
	}
}

func TestCatalogRejectsNonIdentifiers(t *testing.T) {
	for _, id := range []string{"Onewin", "1win.com", "one-win", "one1"} {
		if _, err := casino.Lookup(id); err == nil {
			t.Fatalf("accepted %q", id)
		}
	}
}
