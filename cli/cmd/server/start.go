package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// ServerCmd controls the MangaHub server lifecycle.
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the MangaHub server",
	Long:  "Start, stop, and monitor the MangaHub server.",
}

// startCmd boots a lightweight HTTP server that serves the required endpoints.
var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the API server",
	Long:    "Start a lightweight MangaHub HTTP API server for local development.",
	Example: "mangahub server start",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		addr := fmt.Sprintf(":%d", port)

		mux := http.NewServeMux()
		registerRoutes(mux)

		srv := &http.Server{Addr: addr, Handler: mux}
		go func() {
			_ = srv.ListenAndServe()
		}()

		output.PrintSuccess(cmd, fmt.Sprintf("MangaHub API server started on %s", addr))

		<-cmd.Context().Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	},
}

func init() {
	ServerCmd.AddCommand(startCmd)
	startCmd.Flags().Int("port", 8080, "Port to run the server on")
}

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"Registration successful"}`))
	})

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Username string `json:"username"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		token := fmt.Sprintf("token-%s", strings.TrimSpace(payload.Username))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"token":"%s"}`, token)))
	})

	mux.HandleFunc("/manga/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		query := strings.ToLower(r.URL.Query().Get("q"))
		results := []api.Manga{
			{ID: "one-piece", Title: "One Piece", Status: "ongoing"},
			{ID: "naruto", Title: "Naruto", Status: "completed"},
			{ID: "bleach", Title: "Bleach", Status: "completed"},
		}
		filtered := results[:0]
		for _, m := range results {
			if query == "" || strings.Contains(strings.ToLower(m.Title), query) || strings.Contains(strings.ToLower(m.ID), query) {
				filtered = append(filtered, m)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filtered)
	})

	mux.HandleFunc("/library/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"Manga added to library"}`))
	})

	mux.HandleFunc("/progress/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"Progress updated"}`))
	})
}
