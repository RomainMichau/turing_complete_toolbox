// Command turing_complete_toolbox serves a small web toolbox used while
// playing Turing Complete.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/RomainMichau/turing_complete_toolbox/internal/tools"
)

//go:embed web
var webFS embed.FS

type runRequest struct {
	Input string `json:"input"`
}

type runResponse struct {
	Fields []tools.Field `json:"fields"`
	Error  string        `json:"error,omitempty"`
}

func handleList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, tools.All())
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/tools/")
	tool := tools.Get(id)
	if tool == nil {
		writeJSON(w, http.StatusNotFound, runResponse{Error: "unknown tool: " + id})
		return
	}

	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, runResponse{Error: "invalid request body"})
		return
	}

	fields, err := tool.Run(req.Input)
	if err != nil {
		writeJSON(w, http.StatusOK, runResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, runResponse{Fields: fields})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

// openBrowser opens url in the user's default browser, best effort.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("could not open a browser (%v), go to %s yourself", err, url)
		return
	}
	go cmd.Wait() // reap the child instead of leaving a zombie
}

// browseURL turns the listener address into something a browser accepts.
func browseURL(addr net.Addr) string {
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "http://" + addr.String()
	}
	if host == "" || host == "::" || host == "0.0.0.0" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%s", net.JoinHostPort(host, port))
}

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	open := flag.Bool("open", true, "open the toolbox in a browser on startup")
	flag.Parse()

	static, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("embed web assets: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tools", handleList)
	mux.HandleFunc("POST /api/tools/{id}", handleRun)
	mux.Handle("GET /", http.FileServer(http.FS(static)))

	// Listen before serving so the browser is only launched once the port
	// is actually accepting connections.
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen on %s: %v", *addr, err)
	}

	url := browseURL(ln.Addr())
	log.Printf("turing complete toolbox listening on %s", url)
	if *open {
		openBrowser(url)
	}

	if err := http.Serve(ln, mux); err != nil {
		log.Fatal(err)
	}
}
