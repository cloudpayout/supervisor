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
	if casino.PlayURL("") != "https://1win.com/casino/play/v_spribe:aviator" {
		t.Fatalf("default play url %s", casino.PlayURL(""))
	}
	if casino.PlayURL("es-MX") != "https://1win.com/es-MX/casino/play/v_spribe:aviator" {
		t.Fatalf("locale play url %s", casino.PlayURL("es-MX"))
	}
}

func TestCatalogRejectsNonIdentifiers(t *testing.T) {
	for _, id := range []string{"Onewin", "1win.com", "one-win", "one1"} {
		if _, err := casino.Lookup(id); err == nil {
			t.Fatalf("accepted %q", id)
		}
	}
}
