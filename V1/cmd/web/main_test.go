package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTickShowsDissolvedSaltThenResetClearsIt(t *testing.T) {
	handler := newSim().routes()

	start := getFrame(t, handler, http.MethodGet, "/api/state")
	if start.Tick != 0 {
		t.Fatalf("tick = %d, want 0", start.Tick)
	}
	if moles(start, "dissolved_salt") != 0 {
		t.Fatalf("dissolved salt present before any tick")
	}

	after := getFrame(t, handler, http.MethodPost, "/api/tick")
	if after.Tick != 1 {
		t.Fatalf("tick = %d, want 1", after.Tick)
	}
	if moles(after, "dissolved_salt") < 0.5 {
		t.Fatalf("dissolved salt = %.3f, want the salt to start dissolving", moles(after, "dissolved_salt"))
	}

	page := getBody(t, handler, http.MethodGet, "/")
	if !strings.Contains(page, `id="root"`) {
		t.Fatal("page did not mount the client shell")
	}
	script := getBody(t, handler, http.MethodGet, assetSrc(t, page))
	if !strings.Contains(script, "/api/tick") {
		t.Fatal("client bundle did not call the tick API")
	}

	reset := getFrame(t, handler, http.MethodPost, "/api/reset")
	if reset.Tick != 0 || moles(reset, "salt") < 0.99 {
		t.Fatalf("reset tick=%d salt=%.3f", reset.Tick, moles(reset, "salt"))
	}
}

func getFrame(t *testing.T, handler http.Handler, method, path string) frame {
	t.Helper()
	var got frame
	if err := json.Unmarshal([]byte(getBody(t, handler, method, path)), &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func assetSrc(t *testing.T, page string) string {
	t.Helper()
	const marker = `src="`
	i := strings.Index(page, marker)
	if i < 0 {
		t.Fatal("page did not reference a script")
	}
	rest := page[i+len(marker):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		t.Fatal("script src was not quoted")
	}
	return rest[:j]
}

func getBody(t *testing.T, handler http.Handler, method, path string) string {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s %s -> %d", method, path, rec.Code)
	}
	return rec.Body.String()
}

func moles(f frame, species string) float64 {
	var total float64
	for _, p := range f.Parcels {
		if string(p.Species) == species {
			total += p.Moles
		}
	}
	return total
}
