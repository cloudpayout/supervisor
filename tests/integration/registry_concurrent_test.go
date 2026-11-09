package integration

import (
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParallelStartRegistry(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	const n = 2
	var wg sync.WaitGroup
	ids := make([]string, n)
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			cmd := exec.Command(bin, "start", "--headless")
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			errs[i] = err
			ids[i] = strings.TrimSpace(string(out))
		}(i)
	}
	wg.Wait()
	t.Cleanup(func() {
		for _, id := range ids {
			if id != "" {
				stopID(t, bin, env, id)
			}
		}
	})
	for i := range errs {
		if errs[i] != nil {
			t.Fatalf("start %d: %v %s", i, errs[i], ids[i])
		}
	}
	if ids[0] == "" || ids[0] == ids[1] {
		t.Fatalf("ids %+v", ids)
	}
	deadline := time.Now().Add(40 * time.Second)
	for time.Now().Before(deadline) {
		doc := readDoc(t, regPath)
		seen := map[string]bool{}
		for _, inst := range doc.Instances {
			seen[inst.ID] = true
		}
		if seen[ids[0]] && seen[ids[1]] {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatal("parallel ids missing from registry")
}
