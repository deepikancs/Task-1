package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"organisation-task-control-centre/backend/internal/models"
	"organisation-task-control-centre/backend/internal/repository"
)

var (
	isoDateRegex  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	isoMonthRegex = regexp.MustCompile(`^\d{4}-\d{2}$`)
)

// Handler serves HTTP API endpoints for the Organisation Task Control Centre.
type Handler struct {
	repo             repository.Store
	defaultDemoToday string
	demoMode         bool
}

// NewHandler creates a new HTTP API Handler.
func NewHandler(repo repository.Store, defaultDemoToday string, demoMode bool) *Handler {
	return &Handler{
		repo:             repo,
		defaultDemoToday: defaultDemoToday,
		demoMode:         demoMode,
	}
}

// RegisterRoutes attaches all API routes to the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", h.handleHealth)
	mux.HandleFunc("/api/metadata", h.handleMetadata)
	mux.HandleFunc("/api/tasks", h.handleTasks)
	mux.HandleFunc("/api/tasks/aggregate", h.handleAggregateTasks)
	mux.HandleFunc("/api/tasks/", h.handleTaskSubroutes)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	storageMode := "mysql"
	if h.demoMode {
		storageMode = "in_memory_demo"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"demoMode": h.demoMode,
		"storage":  storageMode,
		"mode":     "demo_permissions_no_auth",
	})
}

// GET /api/metadata
func (h *Handler) handleMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	meta, err := h.repo.GetMetadata(r.Context())
	if err != nil {
		log.Printf("[ERROR] GET /api/metadata failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load dashboard metadata")
		return
	}
	meta.DemoMode = h.demoMode

	writeJSON(w, http.StatusOK, meta)
}

// GET /api/tasks?mode=role&filter=All&person=All&date=2026-10-11&today=2026-10-11
func (h *Handler) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	mode := strings.TrimSpace(q.Get("mode"))
	if mode == "" {
		mode = "role"
	}
	if mode != "role" && mode != "type" {
		writeError(w, http.StatusBadRequest, "invalid mode parameter; expected 'role' or 'type'")
		return
	}

	selectedFilter := strings.TrimSpace(q.Get("filter"))
	if selectedFilter == "" {
		selectedFilter = "All"
	}

	selectedPerson := strings.TrimSpace(q.Get("person"))
	if selectedPerson == "" {
		selectedPerson = "All"
	}

	todayISO := strings.TrimSpace(q.Get("today"))
	if todayISO == "" {
		todayISO = h.defaultDemoToday
	}
	if !isoDateRegex.MatchString(todayISO) {
		writeError(w, http.StatusBadRequest, "invalid 'today' date format; expected YYYY-MM-DD")
		return
	}

	dateISO := strings.TrimSpace(q.Get("date"))
	if dateISO == "" {
		dateISO = todayISO
	}
	if !isoDateRegex.MatchString(dateISO) {
		writeError(w, http.StatusBadRequest, "invalid 'date' format; expected YYYY-MM-DD")
		return
	}

	tasks, err := h.repo.ListTasks(r.Context(), mode, selectedFilter, selectedPerson, dateISO, todayISO)
	if err != nil {
		log.Printf("[ERROR] GET /api/tasks failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load tasks")
		return
	}

	writeJSON(w, http.StatusOK, models.TasksResponse{Tasks: tasks})
}

// GET /api/tasks/aggregate?filter=All&monthly=false&date=2026-10-11&month=2026-10&today=2026-10-11
func (h *Handler) handleAggregateTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	selectedFilter := strings.TrimSpace(q.Get("filter"))
	if selectedFilter == "" {
		selectedFilter = "All"
	}

	isMonthly := strings.EqualFold(strings.TrimSpace(q.Get("monthly")), "true")

	todayISO := strings.TrimSpace(q.Get("today"))
	if todayISO == "" {
		todayISO = h.defaultDemoToday
	}
	if !isoDateRegex.MatchString(todayISO) {
		writeError(w, http.StatusBadRequest, "invalid 'today' date format; expected YYYY-MM-DD")
		return
	}

	dateISO := strings.TrimSpace(q.Get("date"))
	if dateISO == "" {
		dateISO = todayISO
	}
	if !isoDateRegex.MatchString(dateISO) {
		writeError(w, http.StatusBadRequest, "invalid 'date' format; expected YYYY-MM-DD")
		return
	}

	monthISO := strings.TrimSpace(q.Get("month"))
	if monthISO == "" && len(dateISO) >= 7 {
		monthISO = dateISO[:7]
	}
	if !isoMonthRegex.MatchString(monthISO) {
		writeError(w, http.StatusBadRequest, "invalid 'month' format; expected YYYY-MM")
		return
	}

	res, err := h.repo.GetAggregateTasks(r.Context(), selectedFilter, isMonthly, dateISO, monthISO, todayISO)
	if err != nil {
		log.Printf("[ERROR] GET /api/tasks/aggregate failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load aggregate task metrics")
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// Dispatches:
// - GET  /api/tasks/{uid}
// - GET  /api/tasks/{uid}/history
// - POST /api/tasks/{uid}/submit
// - POST /api/tasks/{uid}/verify
// - POST /api/tasks/{uid}/escalate
func (h *Handler) handleTaskSubroutes(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if trimmed == "" || trimmed == "aggregate" {
		http.NotFound(w, r)
		return
	}

	// Check suffix actions
	if strings.HasSuffix(trimmed, "/submit") {
		rawUID := strings.TrimSuffix(trimmed, "/submit")
		uid, err := url.PathUnescape(rawUID)
		if err != nil || strings.TrimSpace(uid) == "" {
			writeError(w, http.StatusBadRequest, "invalid task identifier")
			return
		}
		h.handleSubmitTask(w, r, uid)
		return
	}

	if strings.HasSuffix(trimmed, "/verify") {
		rawUID := strings.TrimSuffix(trimmed, "/verify")
		uid, err := url.PathUnescape(rawUID)
		if err != nil || strings.TrimSpace(uid) == "" {
			writeError(w, http.StatusBadRequest, "invalid task identifier")
			return
		}
		h.handleVerifyTask(w, r, uid)
		return
	}

	if strings.HasSuffix(trimmed, "/escalate") {
		rawUID := strings.TrimSuffix(trimmed, "/escalate")
		uid, err := url.PathUnescape(rawUID)
		if err != nil || strings.TrimSpace(uid) == "" {
			writeError(w, http.StatusBadRequest, "invalid task identifier")
			return
		}
		h.handleEscalateTask(w, r, uid)
		return
	}

	if strings.HasSuffix(trimmed, "/history") {
		rawUID := strings.TrimSuffix(trimmed, "/history")
		uid, err := url.PathUnescape(rawUID)
		if err != nil || strings.TrimSpace(uid) == "" {
			writeError(w, http.StatusBadRequest, "invalid task identifier")
			return
		}
		h.handleTaskHistory(w, r, uid)
		return
	}

	// Otherwise treat as GET /api/tasks/{uid}
	uid, err := url.PathUnescape(trimmed)
	if err != nil || strings.TrimSpace(uid) == "" {
		writeError(w, http.StatusBadRequest, "invalid task identifier")
		return
	}
	h.handleGetTaskDetail(w, r, uid)
}

// GET /api/tasks/{uid}?date=YYYY-MM-DD&today=YYYY-MM-DD
func (h *Handler) handleGetTaskDetail(w http.ResponseWriter, r *http.Request, uid string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	todayISO := strings.TrimSpace(q.Get("today"))
	if todayISO == "" {
		todayISO = h.defaultDemoToday
	}
	if !isoDateRegex.MatchString(todayISO) {
		writeError(w, http.StatusBadRequest, "invalid 'today' date format; expected YYYY-MM-DD")
		return
	}

	dateISO := strings.TrimSpace(q.Get("date"))
	if dateISO == "" {
		dateISO = todayISO
	}
	if !isoDateRegex.MatchString(dateISO) {
		writeError(w, http.StatusBadRequest, "invalid 'date' format; expected YYYY-MM-DD")
		return
	}

	task, err := h.repo.GetTaskByUID(r.Context(), uid, dateISO, todayISO)
	if errors.Is(err, repository.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, "task instance not found")
		return
	}
	if err != nil {
		log.Printf("[ERROR] GET /api/tasks/%s failed: %v", uid, err)
		writeError(w, http.StatusInternalServerError, "failed to load task detail")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// GET /api/tasks/{uid}/history?date=YYYY-MM-DD&today=YYYY-MM-DD
func (h *Handler) handleTaskHistory(w http.ResponseWriter, r *http.Request, uid string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	todayISO := strings.TrimSpace(q.Get("today"))
	if todayISO == "" {
		todayISO = h.defaultDemoToday
	}
	dateISO := strings.TrimSpace(q.Get("date"))
	if dateISO == "" {
		dateISO = todayISO
	}
	if !isoDateRegex.MatchString(dateISO) {
		writeError(w, http.StatusBadRequest, "invalid 'date' format; expected YYYY-MM-DD")
		return
	}

	task, err := h.repo.GetTaskByUID(r.Context(), uid, dateISO, todayISO)
	if errors.Is(err, repository.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, "task instance not found")
		return
	}
	if err != nil {
		log.Printf("[ERROR] GET /api/tasks/%s/history failed: %v", uid, err)
		writeError(w, http.StatusInternalServerError, "failed to load task activity history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"uid":     task.UID,
		"dateISO": task.DateISO,
		"history": task.History,
	})
}

// POST /api/tasks/{uid}/submit
func (h *Handler) handleSubmitTask(w http.ResponseWriter, r *http.Request, uid string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	todayISO := strings.TrimSpace(req.Today)
	if todayISO == "" {
		todayISO = h.defaultDemoToday
	}
	if !isoDateRegex.MatchString(todayISO) {
		writeError(w, http.StatusBadRequest, "invalid 'today' date format; expected YYYY-MM-DD")
		return
	}

	dateISO := strings.TrimSpace(req.Date)
	if dateISO == "" {
		dateISO = todayISO
	}
	if !isoDateRegex.MatchString(dateISO) {
		writeError(w, http.StatusBadRequest, "invalid 'date' format; expected YYYY-MM-DD")
		return
	}

	nextState, err := h.repo.SubmitTask(r.Context(), uid, dateISO, todayISO, req.Remarks)
	if errors.Is(err, repository.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, "task instance not found")
		return
	}
	if err != nil {
		log.Printf("[ERROR] POST /api/tasks/%s/submit failed: %v", uid, err)
		writeError(w, http.StatusInternalServerError, "failed to submit task")
		return
	}

	writeJSON(w, http.StatusOK, models.MutationResponse{
		Success: true,
		State:   nextState,
	})
}

// POST /api/tasks/{uid}/verify
func (h *Handler) handleVerifyTask(w http.ResponseWriter, r *http.Request, uid string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.VerifyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action != "approve" && action != "reject" {
		writeError(w, http.StatusBadRequest, "invalid 'action'; expected 'approve' or 'reject'")
		return
	}

	todayISO := strings.TrimSpace(req.Today)
	if todayISO == "" {
		todayISO = h.defaultDemoToday
	}
	if !isoDateRegex.MatchString(todayISO) {
		writeError(w, http.StatusBadRequest, "invalid 'today' date format; expected YYYY-MM-DD")
		return
	}

	dateISO := strings.TrimSpace(req.Date)
	if dateISO == "" {
		dateISO = todayISO
	}
	if !isoDateRegex.MatchString(dateISO) {
		writeError(w, http.StatusBadRequest, "invalid 'date' format; expected YYYY-MM-DD")
		return
	}

	nextState, err := h.repo.VerifyTask(r.Context(), uid, dateISO, todayISO, action, req.Remarks)
	if errors.Is(err, repository.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, "task instance not found")
		return
	}
	if err != nil {
		log.Printf("[ERROR] POST /api/tasks/%s/verify failed: %v", uid, err)
		writeError(w, http.StatusInternalServerError, "failed to verify task")
		return
	}

	writeJSON(w, http.StatusOK, models.MutationResponse{
		Success: true,
		State:   nextState,
	})
}

// POST /api/tasks/{uid}/escalate
func (h *Handler) handleEscalateTask(w http.ResponseWriter, r *http.Request, uid string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.EscalateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	todayISO := strings.TrimSpace(req.Today)
	if todayISO == "" {
		todayISO = h.defaultDemoToday
	}
	if !isoDateRegex.MatchString(todayISO) {
		writeError(w, http.StatusBadRequest, "invalid 'today' date format; expected YYYY-MM-DD")
		return
	}

	dateISO := strings.TrimSpace(req.Date)
	if dateISO == "" {
		dateISO = todayISO
	}
	if !isoDateRegex.MatchString(dateISO) {
		writeError(w, http.StatusBadRequest, "invalid 'date' format; expected YYYY-MM-DD")
		return
	}

	err := h.repo.EscalateTask(r.Context(), uid, dateISO, todayISO, req.Recipient, req.Reason)
	if errors.Is(err, repository.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, "task instance not found")
		return
	}
	if err != nil {
		log.Printf("[ERROR] POST /api/tasks/%s/escalate failed: %v", uid, err)
		writeError(w, http.StatusInternalServerError, "failed to record escalation")
		return
	}

	writeJSON(w, http.StatusOK, models.MutationResponse{
		Success: true,
		Message: "Demo escalation recorded in task history",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("[ERROR] JSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
