package casino

import "fmt"

// Casino is one registered casino. Identifiers are lowercase ASCII (constitution V).
// specs/020-open-brazino-page: brazino opens the registered home page with no game.
// onewin starts with a session and enters the registered game aviator (004/007).
type Casino struct {
	ID   string
	Page string
	Game string
}

// Lookup returns the casino for an identifier. Hosts and mixed-case names are refused.
func Lookup(id string) (Casino, error) {
	switch id {
	case "onewin":
		return Casino{ID: "onewin", Page: "https://1win.com/", Game: "aviator"}, nil
	case "brazino":
		return Casino{ID: "brazino", Page: "https://www.brazino777.mx/", Game: ""}, nil
	default:
		return Casino{}, fmt.Errorf("usage: casino %q is not supported", id)
	}
}

// StartURL is the page the worker opens after cookies are set (020 FR-002, 004 play URL).
// When Game is empty, returns registered Page (home page). Otherwise returns the game play URL.
func StartURL(casinoID, locale string) (string, error) {
	c, err := Lookup(casinoID)
	if err != nil {
		return "", err
	}
	if c.Game == "" {
		return c.Page, nil
	}
	return PlayURL(locale), nil
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