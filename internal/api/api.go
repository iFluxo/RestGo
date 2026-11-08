package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"restgo/internal/store"
)

type handler struct {
	store *store.Store
}

// New returns a complete http.Handler: router plus middleware.
func New(st *store.Store) http.Handler {
	h := &handler{store: st}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /items", h.listItems)
	mux.HandleFunc("POST /items", h.createItem)
	mux.HandleFunc("GET /items/{id}", h.getItem)
	mux.HandleFunc("PUT /items/{id}", h.updateItem)
	mux.HandleFunc("DELETE /items/{id}", h.deleteItem)

	return recoverer(logging(mux))
}

func (h *handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) listItems(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.store.List())
}

func (h *handler) getItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, ok := h.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *handler) createItem(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}

	item := h.store.Create(in)
	w.Header().Set("Location", fmt.Sprintf("/items/%d", item.ID))
	writeJSON(w, http.StatusCreated, item)
}

func (h *handler) updateItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	in, ok := decodeInput(w, r)
	if !ok {
		return
	}

	item, ok := h.store.Update(id, in)
	if !ok {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !h.store.Delete(id) {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// decodeInput reads and validates the body, writing an error response on failure.
func decodeInput(w http.ResponseWriter, r *http.Request) (store.ItemInput, bool) {
	var in store.ItemInput

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&in); err != nil {
		if errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "body must not be empty")
		} else {
			writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		}
		return in, false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "body must contain a single JSON object")
		return in, false
	}

	if err := validate(&in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return in, false
	}
	return in, true
}

func validate(in *store.ItemInput) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return errors.New("name is required")
	}
	if len(in.Name) > 100 {
		return errors.New("name must be at most 100 characters")
	}
	if in.Price < 0 {
		return errors.New("price must not be negative")
	}
	return nil
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id must be a positive integer")
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
