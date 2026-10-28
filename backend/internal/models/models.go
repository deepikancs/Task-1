package models

import "strings"

// OrganisationSummary holds top-level KPI counters for the left panel.
type OrganisationSummary struct {
	ActiveTasks int `json:"activeTasks"`
	Overdue     int `json:"overdue"`
	RepeatMiss  int `json:"repeatMiss"`
}

// RoleMetadata represents a role and its assigned owners and verifier.
type RoleMetadata struct {
	Name     string   `json:"name"`
	Count    int      `json:"count"`
	People   []string `json:"people"`
	Verifier string   `json:"verifier"`
}

// TaskTypeMetadata represents a task type filter item and its template count.
type TaskTypeMetadata struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// MetadataResponse matches the JSON shape expected by fetchDashboardMetadata() in React.
type MetadataResponse struct {
	DemoMode            bool                `json:"demoMode"`
	OrganisationSummary OrganisationSummary `json:"organisationSummary"`
	Roles               []RoleMetadata      `json:"roles"`
	PeopleByRole        map[string][]string `json:"peopleByRole"`
	VerifierByRole      map[string]string   `json:"verifierByRole"`
	TaskTypes           []TaskTypeMetadata  `json:"taskTypes"`
}

// FlowStep represents one of the 4 stages in the horizontal ribbon and detail workflow.
type FlowStep struct {
	Status string `json:"status"`
	No     string `json:"no"`
	Title  string `json:"title"`
	Person string `json:"person"`
}

// HistoryItem represents a chronological entry in the Historical Activity timeline.
type HistoryItem struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Meta   string `json:"meta"`
	Colour string `json:"colour"`
}

// TaskInstanceDTO matches the exact task object shape expected by the React frontend.
type TaskInstanceDTO struct {
	InstanceID          int           `json:"instanceId"`
	ID                  string        `json:"id"`
	UID                 string        `json:"uid"`
	Name                string        `json:"name"`
	Role                string        `json:"role"`
	Category            string        `json:"category"`
	Frequency           string        `json:"frequency"`
	Time                string        `json:"time"`
	Misses              int           `json:"misses"`
	Owner               string        `json:"owner"`
	Verifier            string        `json:"verifier"`
	State               string        `json:"state"`
	Effective           string        `json:"effective"`
	DateISO             string        `json:"dateISO"`
	SubmissionNotes     string        `json:"submissionNotes,omitempty"`
	VerificationRemarks string        `json:"verificationRemarks,omitempty"`
	Flow                []FlowStep    `json:"flow"`
	History             []HistoryItem `json:"history"`
}

// TasksResponse wraps the list of task instances for GET /api/tasks.
type TasksResponse struct {
	Tasks []TaskInstanceDTO `json:"tasks"`
}

// AggregateItemDTO represents a task template's completion metrics in Task Type View.
type AggregateItemDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	Category   string `json:"category"`
	Frequency  string `json:"frequency"`
	Time       string `json:"time"`
	Misses     int    `json:"misses"`
	People     int    `json:"people"`
	Percentage int    `json:"percentage"`
	Completed  int    `json:"completed"`
}

// AggregateSummary holds summary strip counts for Task Type View.
type AggregateSummary struct {
	UniqueTasks int `json:"uniqueTasks"`
	RolesCount  int `json:"rolesCount"`
	PeopleCount int `json:"peopleCount"`
}

// AggregateResponse matches the JSON shape expected by fetchAggregateTasks() in React.
type AggregateResponse struct {
	Items   []AggregateItemDTO `json:"items"`
	Summary AggregateSummary   `json:"summary"`
}

// SubmitTaskRequest is the JSON payload for POST /api/tasks/{uid}/submit.
// Note: Executed under Demo Permissions mode (not real authentication).
type SubmitTaskRequest struct {
	Date    string `json:"date"`
	Today   string `json:"today"`
	Remarks string `json:"remarks"`
}

// VerifyTaskRequest is the JSON payload for POST /api/tasks/{uid}/verify.
// Note: Executed under Demo Permissions mode (not real authentication).
type VerifyTaskRequest struct {
	Date    string `json:"date"`
	Today   string `json:"today"`
	Action  string `json:"action"` // "approve" | "reject"
	Remarks string `json:"remarks"`
}

// EscalateTaskRequest is the JSON payload for POST /api/tasks/{uid}/escalate.
// Records a demo-only escalation event in the task activity history.
type EscalateTaskRequest struct {
	Date      string `json:"date"`
	Today     string `json:"today"`
	Recipient string `json:"recipient"`
	Reason    string `json:"reason"`
}

// MutationResponse is returned after submitting, approving, or rejecting a task.
type MutationResponse struct {
	Success bool   `json:"success"`
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
}

// BaseName extracts the person's name before the department separator (" — ").
func BaseName(displayLabel string) string {
	parts := strings.SplitN(displayLabel, " — ", 2)
	return parts[0]
}

// BuildFlow constructs the 4-stage workflow steps for a task instance.
// A task submission never counts as closed/complete until step 3 (Verification) is approved.
func BuildFlow(ownerDisplayLabel, verifierTitle, state string) []FlowStep {
	ownerShort := BaseName(ownerDisplayLabel)
	if verifierTitle == "" {
		verifierTitle = "Reporting Authority"
	}

	switch state {
	case "approved_early":
		return []FlowStep{
			{Status: "done", No: "1", Title: "Completed early", Person: ownerShort},
			{Status: "done", No: "2", Title: "Submitted early", Person: ownerShort},
			{Status: "done", No: "3", Title: "Verified early", Person: verifierTitle},
			{Status: "done", No: "4", Title: "Closed", Person: "System"},
		}
	case "submitted_early":
		return []FlowStep{
			{Status: "done", No: "1", Title: "Completed early", Person: ownerShort},
			{Status: "done", No: "2", Title: "Submitted early", Person: ownerShort},
			{Status: "wait", No: "3", Title: "Waiting verification", Person: verifierTitle},
			{Status: "future", No: "4", Title: "Closure pending", Person: "System"},
		}
	case "scheduled":
		return []FlowStep{
			{Status: "future", No: "1", Title: "Yet to start", Person: ownerShort},
			{Status: "future", No: "2", Title: "Submission", Person: "Scheduled"},
			{Status: "future", No: "3", Title: "Verification", Person: verifierTitle},
			{Status: "future", No: "4", Title: "Closure", Person: "Scheduled"},
		}
	case "approved":
		return []FlowStep{
			{Status: "done", No: "1", Title: "Work completed", Person: ownerShort},
			{Status: "done", No: "2", Title: "Submitted", Person: ownerShort},
			{Status: "done", No: "3", Title: "Verified", Person: verifierTitle},
			{Status: "done", No: "4", Title: "Closed", Person: "System"},
		}
	case "submitted":
		return []FlowStep{
			{Status: "done", No: "1", Title: "Work completed", Person: ownerShort},
			{Status: "done", No: "2", Title: "Submitted", Person: ownerShort},
			{Status: "wait", No: "3", Title: "Waiting verification", Person: verifierTitle},
			{Status: "", No: "4", Title: "Closure pending", Person: "System"},
		}
	default: // "overdue" or "live"
		return []FlowStep{
			{Status: "current", No: "1", Title: "Owner pending", Person: ownerShort},
			{Status: "", No: "2", Title: "Submission", Person: "Not filed"},
			{Status: "", No: "3", Title: "Verification", Person: verifierTitle},
			{Status: "", No: "4", Title: "Closure", Person: "Pending"},
		}
	}
}
