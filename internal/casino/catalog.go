package casino

import "fmt"

// Casino is one registered casino. FR-008 names onewin and aviator.
// FR-009 requires a lowercase ASCII identifier. No command starts the game.
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
