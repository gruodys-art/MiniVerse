// Command web serves the demo world and the browser client.
//
// The client is the React app in web/. World state stays in this process.
// Build the client into dist before running or testing:
//
//	npm --prefix web install
//	npm --prefix web run build
//	go run ./cmd/web
//
// For UI work, keep this server up and run `npm --prefix web run dev`.
// Vite serves the client and proxies /api to this process.
package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"miniverse/chem"
)

//go:generate npm --prefix ../../web run build

//go:embed all:dist
var page embed.FS

type frame struct {
	Tick    int               `json:"tick"`
	Mass    float64           `json:"mass"`
	Parcels []chem.ParcelView `json:"parcels"`
}

type sim struct {
	mu    sync.Mutex
	world *chem.World
	tick  int
}

func newSim() *sim {
	return &sim{world: chem.DemoWorld()}
}

func (s *sim) snapshot() frame {
	return frame{
		Tick:    s.tick,
		Mass:    s.world.Mass(),
		Parcels: s.world.Snapshot(),
	}
}

func (s *sim) routes() http.Handler {
	sub, err := fs.Sub(page, "dist")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("POST /api/tick", s.handleTick)
	mux.HandleFunc("POST /api/reset", s.handleReset)
	mux.Handle("GET /", http.FileServer(http.FS(sub)))
	return mux
}

func (s *sim) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, s.snapshot())
}

func (s *sim) handleTick(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.world.Tick(1)
	s.tick++
	writeJSON(w, s.snapshot())
}

func (s *sim) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.world = chem.DemoWorld()
	s.tick = 0
	writeJSON(w, s.snapshot())
}

func writeJSON(w http.ResponseWriter, body frame) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Print(err)
	}
}

func main() {
	addr := "127.0.0.1:8080"
	log.Printf("open http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, newSim().routes()))
}
