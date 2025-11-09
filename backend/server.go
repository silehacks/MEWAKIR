package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Server provides HTTP handlers for the backend API.
type Server struct {
	store      *MeetingStore
	hub        *Hub
	mux        *http.ServeMux
	httpServer *http.Server
}

// NewServer builds a new Server instance.
func NewServer(store *MeetingStore, hub *Hub) *Server {
	s := &Server{
		store: store,
		hub:   hub,
		mux:   http.NewServeMux(),
	}
	s.routes()
	return s
}

// Start begins serving HTTP traffic.
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.withCORS(s.mux),
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/api/login", s.handleLogin)
	s.mux.HandleFunc("/api/meetings", s.handleMeetings)
	s.mux.HandleFunc("/api/meetings/join", s.handleJoinMeeting)
	s.mux.HandleFunc("/ws/signaling", s.handleWebsocket)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": name})
}

func (s *Server) handleMeetings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		meeting, err := s.store.CreateMeeting(r.Context())
		if err != nil {
			http.Error(w, "failed to create meeting", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"id":        meeting.ID,
			"password":  meeting.Password,
			"expiresAt": meeting.ExpiresAt,
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleJoinMeeting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		MeetingID string `json:"meetingId"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	meetingID := strings.TrimSpace(body.MeetingID)
	password := strings.TrimSpace(body.Password)
	if meetingID == "" || password == "" {
		http.Error(w, "meetingId and password are required", http.StatusBadRequest)
		return
	}
	meeting, err := s.store.ValidateMeeting(r.Context(), meetingID, password)
	if err != nil {
		if errors.Is(err, ErrMeetingNotFound) {
			http.Error(w, ErrMeetingNotFound.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "failed to validate meeting", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":        meeting.ID,
		"expiresAt": meeting.ExpiresAt,
	})
}

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.hub.ServeWS(w, r)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
