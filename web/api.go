package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"pharmacycounter/model"
)

type BoardState struct {
	Pending   []model.PharmacyTicket `json:"pending"`
	Called    []model.PharmacyTicket `json:"called"`
	Completed []model.PharmacyTicket `json:"completed"`
}

func JSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func BoardFromTickets(tickets []model.PharmacyTicket) BoardState {
	board := BoardState{Pending: make([]model.PharmacyTicket, 0), Called: make([]model.PharmacyTicket, 0), Completed: make([]model.PharmacyTicket, 0)}
	for _, ticket := range tickets {
		switch ticket.Status {
		case model.StatusPending:
			board.Pending = append(board.Pending, ticket)
		case model.StatusCalled:
			board.Called = append(board.Called, ticket)
		case model.StatusCompleted:
			board.Completed = append(board.Completed, ticket)
		}
	}
	return board
}

func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		JSONResponse(w, http.StatusOK, map[string]string{"status": "ready", "service": "pharmacy-counter"})
	})
}

func TicketHandler(tickets func() []model.PharmacyTicket) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		query := strings.TrimSpace(r.URL.Query().Get("status"))
		items := tickets()
		if query != "" {
			filtered := make([]model.PharmacyTicket, 0, len(items))
			for _, item := range items {
				if item.Status == query {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
		JSONResponse(w, http.StatusOK, BoardFromTickets(items))
	})
}

func MethodGuard(next http.Handler, method string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func PlainText(value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(value))
	})
}
