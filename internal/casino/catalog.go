package casino

import "fmt"

// Casino is one registered casino. Identifiers are lowercase ASCII (constitution V).
// onewin starts with a session enter the registered game aviator.
type Casino struct {
	ID   string
	Page string
	Game string
}

// Lookup returns the casino for an identifier. Onewin and 1win.com are refused.
func Lookup(id string) (Casino, error) {
	if id == "onewin" {
		return Casino{ID: "onewin", Page: "https://1win.com/", Game: "aviator"}, nil
	}
	return Casino{}, fmt.Errorf("usage: casino %q is not supported", id)
}

// PlayURL is the Aviator play address for onewin.
// An empty locale uses the English unprefixed path.
// A non-empty locale is one path segment, for example es-MX.
func PlayURL(locale string) string {
	const game = "casino/play/v_spribe:aviator"
	if locale == "" {
		return "https://1win.com/" + game
	}
	return "https://1win.com/" + locale + "/" + game
}
