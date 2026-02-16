package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/example/golang-project/internal/model"
	"github.com/example/golang-project/internal/service"
)

// ChurchMemberHandler wires HTTP requests to the ChurchMemberService.
type ChurchMemberHandler struct {
	svc *service.ChurchMemberService
}

// NewChurchMemberHandler creates a new handler with the given service.
func NewChurchMemberHandler(svc *service.ChurchMemberService) *ChurchMemberHandler {
	return &ChurchMemberHandler{svc: svc}
}

// CreateMemberHandler handles POST /members
// @Summary Create a new church member
// @Description Create a new church member with name, email, and optional biography
// @Tags members
// @Accept json
// @Produce json
// @Param member body model.ChurchMember true "Church member data"
// @Success 201 {object} map[string]int64 "Member created"
// @Failure 400 {string} string "Invalid request body or validation error"
// @Failure 500 {string} string "Internal server error"
// @Router /members [post]
func (h *ChurchMemberHandler) CreateMemberHandler(w http.ResponseWriter, r *http.Request) {
	var in model.ChurchMember
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	id, err := h.svc.CreateMember(r.Context(), &in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// GetMemberHandler handles GET /members/{id}
// @Summary Get church member by ID
// @Description Retrieve a church member by their ID
// @Tags members
// @Produce json
// @Param id path int64 true "Member ID"
// @Success 200 {object} model.ChurchMember "Member data"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Member not found"
// @Failure 500 {string} string "Internal server error"
// @Router /members/{id} [get]
func (h *ChurchMemberHandler) GetMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	m, err := h.svc.GetMember(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if m == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// UpdateMemberHandler handles PUT /members/{id}
// @Summary Update a church member
// @Description Update church member information including name, email, phone, address, and biography
// @Tags members
// @Accept json
// @Param id path int64 true "Member ID"
// @Param member body model.ChurchMember true "Updated member data"
// @Success 204 {string} string "No content"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Member not found"
// @Failure 500 {string} string "Internal server error"
// @Router /members/{id} [put]
func (h *ChurchMemberHandler) UpdateMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var in model.ChurchMember
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	in.ID = id
	if err := h.svc.UpdateMember(r.Context(), &in); err != nil {
		if err.Error() == "member not found" {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteMemberHandler handles DELETE /members/{id}
// @Summary Delete a church member
// @Description Delete church member by ID
// @Tags members
// @Param id path int64 true "Member ID"
// @Success 204 {string} string "No content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 500 {string} string "Internal server error"
// @Router /members/{id} [delete]
func (h *ChurchMemberHandler) DeleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteMember(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListMembersHandler handles GET /members
// @Summary List all church members
// @Description Retrieve all church members from the database, ordered by join date (newest first)
// @Tags members
// @Produce json
// @Success 200 {array} model.ChurchMember "List of members"
// @Failure 500 {string} string "Internal server error"
// @Router /members [get]
func (h *ChurchMemberHandler) ListMembersHandler(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListMembers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if list == nil {
		list = []*model.ChurchMember{}
	}
	json.NewEncoder(w).Encode(list)
}

// ListMembersByDateHandler handles GET /members/joined?start=2024-01-01&end=2024-12-31
// @Summary List church members by joined date range
// @Description Retrieve church members joined within a specific date range
// @Tags members
// @Produce json
// @Param start query string true "Start date (YYYY-MM-DD)"
// @Param end query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} model.ChurchMember "List of members"
// @Failure 400 {string} string "Invalid date format"
// @Failure 500 {string} string "Internal server error"
// @Router /members/joined [get]
func (h *ChurchMemberHandler) ListMembersByDateHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, "start and end date parameters are required", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, "invalid start date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		http.Error(w, "invalid end date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Set end date to end of day
	endDate = endDate.Add(24 * time.Hour)

	list, err := h.svc.ListMembersByJoinedDate(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if list == nil {
		list = []*model.ChurchMember{}
	}
	json.NewEncoder(w).Encode(list)
}
