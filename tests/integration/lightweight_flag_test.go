package integration

import (
	"testing"
)

func TestNoLightweightHeadlessLifecycle(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	off := startWith(t, bin, env, "start", "--headless", "--no-lightweight")
	on := startHeadless(t, bin, env)
	t.Cleanup(func() {
		stopID(t, bin, env, off)
		stopID(t, bin, env, on)
	})
	doc := readDoc(t, regPath)
	var sawOff, sawOn bool
	for _, inst := range doc.Instances {
		if inst.ID == off {
			sawOff = true
			if inst.Lightweight == nil || *inst.Lightweight {
				t.Fatalf("off lightweight %+v", inst.Lightweight)
			}
		}
		if inst.ID == on {
			sawOn = true
			if inst.Lightweight == nil || !*inst.Lightweight {
				t.Fatalf("on lightweight %+v", inst.Lightweight)
			}
		}
	}
	if !sawOff || !sawOn {
		t.Fatal("missing rows")
	}
	stopID(t, bin, env, off)
	stopID(t, bin, env, on)
}

func TestHeadedLightweightAndMemoryLimit(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	plain := startWith(t, bin, env, "start")
	off := startWith(t, bin, env, "start", "--no-lightweight")
	limited := startWith(t, bin, env, "start", "--memory-limit", "512M")
	t.Cleanup(func() {
		stopID(t, bin, env, plain)
		stopID(t, bin, env, off)
		stopID(t, bin, env, limited)
	})
	doc := readDoc(t, regPath)
	for _, inst := range doc.Instances {
		switch inst.ID {
		case plain:
			if inst.Headless != nil && *inst.Headless {
				t.Fatal("plain headed flag")
			}
			if inst.Lightweight == nil || !*inst.Lightweight {
				t.Fatal("headed default should be lightweight true")
			}
		case off:
			if inst.Lightweight == nil || *inst.Lightweight {
				t.Fatal("headed opt-out")
			}
		case limited:
			if inst.MemoryLimitBytes == 0 {
				t.Fatal("missing limit")
			}
		}
	}
}
