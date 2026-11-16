package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// Cookie is one browser cookie from a session file. Value stays in this process.
// It is not written to the registry, the command line, the worker log, or an error (FR-005).
type Cookie struct {
	Name        string
	Value       string
	Domain      string
	Path        string
	Expires     *time.Time
	Secure      bool
	HTTPOnly    bool
	SameSite    string
	Partitioned bool
}

type rawCookie struct {
	Name        string  `json:"name"`
	Value       string  `json:"value"`
	Domain      *string `json:"domain"`
	Path        *string `json:"path"`
	Expires     *string `json:"expires"`
	MaxAge      *int    `json:"maxAge"`
	Secure      *bool   `json:"secure"`
	HTTPOnly    *bool   `json:"httpOnly"`
	SameSite    *string `json:"sameSite"`
	Partitioned *bool   `json:"partitioned"`
}

// Parse reads a session file. The document is not copied into the project (FR-001).
// expires wins over maxAge. maxAge is seconds added to now when expires is absent (FR-004).
func Parse(path string, now time.Time) ([]Cookie, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session file: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var doc struct {
		Cookies []json.RawMessage `json:"cookies"`
	}
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("session file: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("session file: extra data")
	}
	if len(doc.Cookies) < 1 {
		return nil, fmt.Errorf("session file: cookies must contain at least one cookie")
	}
	out := make([]Cookie, 0, len(doc.Cookies))
	for i, rawCookie := range doc.Cookies {
		c, err := decodeCookie(rawCookie, now)
		if err != nil {
			return nil, fmt.Errorf("session file: cookie %d: %w", i, err)
		}
		out = append(out, c)
	}
	return out, nil
}

func decodeCookie(raw json.RawMessage, now time.Time) (Cookie, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var in rawCookie
	if err := dec.Decode(&in); err != nil {
		return Cookie{}, err
	}
	if in.Name == "" {
		return Cookie{}, fmt.Errorf("name is empty")
	}
	if in.Value == "" {
		return Cookie{}, fmt.Errorf("value is empty")
	}
	c := Cookie{Name: in.Name, Value: in.Value}
	if in.Domain != nil {
		c.Domain = *in.Domain
	}
	if in.Path != nil {
		c.Path = *in.Path
	}
	if in.Secure != nil {
		c.Secure = *in.Secure
	}
	if in.HTTPOnly != nil {
		c.HTTPOnly = *in.HTTPOnly
	}
	if in.Partitioned != nil {
		c.Partitioned = *in.Partitioned
	}
	if in.SameSite != nil {
		switch *in.SameSite {
		case "Strict", "Lax", "None":
			c.SameSite = *in.SameSite
		default:
			return Cookie{}, fmt.Errorf("sameSite must be Strict, Lax, or None")
		}
	}
	if in.Expires != nil {
		exp, err := time.Parse(time.RFC3339, *in.Expires)
		if err != nil {
			return Cookie{}, fmt.Errorf("expires must be RFC3339")
		}
		exp = exp.UTC()
		c.Expires = &exp
		return c, nil
	}
	if in.MaxAge != nil {
		exp := now.Add(time.Duration(*in.MaxAge) * time.Second).UTC()
		c.Expires = &exp
	}
	return c, nil
}
