package repository

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"organisation-task-control-centre/backend/internal/models"
)

// Store defines the data access contract implemented by both MySQL (TaskRepository)
// and the in-memory Demo Mode repository (DemoTaskRepository).
type Store interface {
	GetMetadata(ctx context.Context) (*models.MetadataResponse, error)
	ListTasks(ctx context.Context, mode, selectedFilter, selectedPerson, dateISO, todayISO string) ([]models.TaskInstanceDTO, error)
	GetTaskByUID(ctx context.Context, uid, dateISO, todayISO string) (*models.TaskInstanceDTO, error)
	GetAggregateTasks(ctx context.Context, selectedFilter string, isMonthly bool, dateISO, monthISO, todayISO string) (*models.AggregateResponse, error)
	SubmitTask(ctx context.Context, uid, dateISO, todayISO, remarks string) (string, error)
	VerifyTask(ctx context.Context, uid, dateISO, todayISO, action, remarks string) (string, error)
	EscalateTask(ctx context.Context, uid, dateISO, todayISO, recipient, reason string) error
}

type demoTemplate struct {
	ID        string
	Name      string
	Role      string
	Category  string
	Frequency string
	TaskType  string
	Time      string
	Pattern   string
	Misses    int
}

type demoOverride struct {
	State               string
	Misses              int
	SubmissionNotes     string
	VerificationRemarks string
	History             []models.HistoryItem
}

var demoOrderedRoles = []string{
	"Principal",
	"IQAC Head",
	"Head of Department",
	"Lab Incharge",
	"Transport Incharge",
	"Maintenance Head",
	"Admin Head",
	"Finance Head",
	"Placement Head",
	"R&D Head",
	"Industry Relations Head",
	"Alumni Head",
}

var demoPeopleByRole = map[string][]string{
	"Principal":               {"Dr. Priya Raman"},
	"IQAC Head":               {"Dr. Senthil Kumar"},
	"Head of Department":      {"Dr. Meena S — CSE", "Dr. Revathi P — ECE", "Dr. Anand K — Mechanical", "Dr. Deepika R — IT", "Dr. Vignesh M — EEE"},
	"Lab Incharge":            {"Mr. Sathish K — Advanced Manufacturing", "Ms. Harini V — AI Lab", "Mr. Naveen P — Electronics Lab"},
	"Transport Incharge":      {"Mr. Suresh Kumar"},
	"Maintenance Head":        {"Mr. Ramesh B"},
	"Admin Head":              {"Ms. Anitha R", "Mr. Vishal K"},
	"Finance Head":            {"Mr. Kannan R"},
	"Placement Head":          {"Ms. Kavitha S", "Mr. Arun P"},
	"R&D Head":                {"Dr. Mani V", "Dr. Nirmala S"},
	"Industry Relations Head": {"Ms. Deepa K"},
	"Alumni Head":             {"Dr. Lavanya P"},
}

var demoVerifierByRole = map[string]string{
	"Principal":               "Chairman / Management",
	"IQAC Head":               "Principal",
	"Head of Department":      "Principal / Dean Academics",
	"Lab Incharge":            "Head of Department",
	"Transport Incharge":      "Admin Head",
	"Maintenance Head":        "Admin Head",
	"Admin Head":              "Principal",
	"Finance Head":            "Management",
	"Placement Head":          "Principal / Dean Placements",
	"R&D Head":                "Principal / Dean R&D",
	"Industry Relations Head": "Principal",
	"Alumni Head":             "Principal",
}

var demoTemplates = []demoTemplate{
	{ID: "t1", Name: "Daily Teaching-Learning Closure", Role: "Principal", Category: "Teaching & Learning", Frequency: "Daily Tasks", TaskType: "Academic Tasks", Time: "5:30 PM", Pattern: "submitted", Misses: 1},
	{ID: "t2", Name: "Academic Deviation & Faculty Absence Review", Role: "Principal", Category: "Teaching & Learning", Frequency: "Daily Tasks", TaskType: "Academic Tasks", Time: "4:45 PM", Pattern: "approved", Misses: 0},
	{ID: "t3", Name: "Student Discipline Exception Closure", Role: "Principal", Category: "Student Discipline", Frequency: "Daily Tasks", TaskType: "Daily Tasks", Time: "5:00 PM", Pattern: "overdue", Misses: 2},
	{ID: "t4", Name: "Accreditation Evidence Update", Role: "IQAC Head", Category: "Documentation", Frequency: "Monthly Tasks", TaskType: "Monthly Tasks", Time: "Month End", Pattern: "submitted", Misses: 0},
	{ID: "t5", Name: "Internal Quality Audit Review", Role: "IQAC Head", Category: "Statutory Committees", Frequency: "Quarterly Tasks", TaskType: "Statutory Tasks", Time: "Quarter End", Pattern: "approved", Misses: 0},
	{ID: "t6", Name: "Department Teaching Progress Review", Role: "Head of Department", Category: "Teaching & Learning", Frequency: "Daily Tasks", TaskType: "Academic Tasks", Time: "5:00 PM", Pattern: "submitted", Misses: 0},
	{ID: "t7", Name: "Student Competency Gap Follow-up", Role: "Head of Department", Category: "Students Competency", Frequency: "Weekly Tasks", TaskType: "Academic Tasks", Time: "Friday", Pattern: "overdue", Misses: 2},
	{ID: "t8", Name: "Faculty Action Closure", Role: "Head of Department", Category: "Faculty Achievement", Frequency: "Daily Tasks", TaskType: "Daily Tasks", Time: "5:30 PM", Pattern: "approved", Misses: 0},
	{ID: "t9", Name: "Department Monthly Performance Review", Role: "Head of Department", Category: "Data Management", Frequency: "Monthly Tasks", TaskType: "Monthly Tasks", Time: "10th", Pattern: "submitted", Misses: 1},
	{ID: "t10", Name: "Laboratory Utilisation Closure", Role: "Lab Incharge", Category: "Infrastructure Utilisation", Frequency: "Daily Tasks", TaskType: "Daily Tasks", Time: "5:00 PM", Pattern: "approved", Misses: 0},
	{ID: "t11", Name: "Equipment Maintenance Exception Review", Role: "Lab Incharge", Category: "Infrastructure Maintenance", Frequency: "Weekly Tasks", TaskType: "Weekly Tasks", Time: "Saturday", Pattern: "submitted", Misses: 0},
	{ID: "t12", Name: "Bus Route Readiness Confirmation", Role: "Transport Incharge", Category: "Infrastructure Maintenance", Frequency: "Daily Tasks", TaskType: "Daily Tasks", Time: "6:30 AM", Pattern: "approved", Misses: 0},
	{ID: "t13", Name: "Daily Maintenance Issue Closure", Role: "Maintenance Head", Category: "Infrastructure Maintenance", Frequency: "Daily Tasks", TaskType: "Daily Tasks", Time: "5:00 PM", Pattern: "overdue", Misses: 3},
	{ID: "t14", Name: "Master Data Integrity Review", Role: "Admin Head", Category: "Data Management", Frequency: "Monthly Tasks", TaskType: "Monthly Tasks", Time: "10th", Pattern: "approved", Misses: 0},
	{ID: "t15", Name: "Inventory Exception Closure", Role: "Admin Head", Category: "Inventory", Frequency: "Monthly Tasks", TaskType: "Monthly Tasks", Time: "15th", Pattern: "submitted", Misses: 1},
	{ID: "t16", Name: "Monthly Budget Variance Closure", Role: "Finance Head", Category: "Finance", Frequency: "Monthly Tasks", TaskType: "Finance Tasks", Time: "5th", Pattern: "overdue", Misses: 3},
	{ID: "t17", Name: "Daily Payment Approval Queue", Role: "Finance Head", Category: "Finance", Frequency: "Daily Tasks", TaskType: "Finance Tasks", Time: "4:30 PM", Pattern: "submitted", Misses: 0},
	{ID: "t18", Name: "Placement Training Outcome Review", Role: "Placement Head", Category: "Placement Training", Frequency: "Weekly Tasks", TaskType: "Weekly Tasks", Time: "Friday", Pattern: "submitted", Misses: 1},
	{ID: "t19", Name: "Placement Drive Data Closure", Role: "Placement Head", Category: "Placement", Frequency: "Daily Tasks", TaskType: "Daily Tasks", Time: "5:30 PM", Pattern: "overdue", Misses: 2},
	{ID: "t20", Name: "Sponsored Research Progress Review", Role: "R&D Head", Category: "R&D", Frequency: "Monthly Tasks", TaskType: "Monthly Tasks", Time: "7th", Pattern: "submitted", Misses: 1},
	{ID: "t21", Name: "Faculty Publication Validation", Role: "R&D Head", Category: "Faculty Achievement", Frequency: "Monthly Tasks", TaskType: "Monthly Tasks", Time: "12th", Pattern: "approved", Misses: 0},
	{ID: "t22", Name: "Industry Interaction Pipeline Review", Role: "Industry Relations Head", Category: "Industry Relations", Frequency: "Weekly Tasks", TaskType: "Weekly Tasks", Time: "Thursday", Pattern: "submitted", Misses: 0},
	{ID: "t23", Name: "Alumni Mentor Engagement Review", Role: "Alumni Head", Category: "Alumni Engagement", Frequency: "Monthly Tasks", TaskType: "Monthly Tasks", Time: "14th", Pattern: "approved", Misses: 0},
}

// DemoTaskRepository serves the full API contract from in-memory demo data
// when DEMO_MODE=true, without connecting to MySQL. All mutations persist in
// memory for the current server run only.
type DemoTaskRepository struct {
	mu               sync.RWMutex
	defaultDemoToday string
	overrides        map[string]demoOverride // key: dateISO + "__" + uid
	nextLogID        int
}

// NewDemoTaskRepository creates an in-memory DemoTaskRepository.
func NewDemoTaskRepository(defaultDemoToday string) *DemoTaskRepository {
	return &DemoTaskRepository{
		defaultDemoToday: defaultDemoToday,
		overrides:        make(map[string]demoOverride),
		nextLogID:        1000,
	}
}

func (d *DemoTaskRepository) GetMetadata(ctx context.Context) (*models.MetadataResponse, error) {
	rolesMeta := make([]models.RoleMetadata, 0, len(demoOrderedRoles))
	for _, roleName := range demoOrderedRoles {
		people := demoPeopleByRole[roleName]
		rolesMeta = append(rolesMeta, models.RoleMetadata{
			Name:     roleName,
			Count:    len(people),
			People:   people,
			Verifier: demoVerifierByRole[roleName],
		})
	}

	typeCounts := make(map[string]int)
	for _, tmpl := range demoTemplates {
		typeCounts[tmpl.TaskType]++
	}

	taskTypesMeta := make([]models.TaskTypeMetadata, 0, len(orderedTaskTypes))
	for _, tName := range orderedTaskTypes {
		taskTypesMeta = append(taskTypesMeta, models.TaskTypeMetadata{
			Name:  tName,
			Count: typeCounts[tName],
		})
	}

	return &models.MetadataResponse{
		OrganisationSummary: models.OrganisationSummary{
			ActiveTasks: 86,
			Overdue:     9,
			RepeatMiss:  6,
		},
		Roles:          rolesMeta,
		PeopleByRole:   demoPeopleByRole,
		VerifierByRole: demoVerifierByRole,
		TaskTypes:      taskTypesMeta,
	}, nil
}

func (d *DemoTaskRepository) ListTasks(
	ctx context.Context,
	mode, selectedFilter, selectedPerson, dateISO, todayISO string,
) ([]models.TaskInstanceDTO, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.buildInstancesLocked(mode, selectedFilter, selectedPerson, dateISO, todayISO), nil
}

func (d *DemoTaskRepository) GetTaskByUID(
	ctx context.Context,
	uid, dateISO, todayISO string,
) (*models.TaskInstanceDTO, error) {
	if _, _, err := splitTaskUID(uid); err != nil {
		return nil, err
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	all := d.buildInstancesLocked("role", "All", "All", dateISO, todayISO)
	for i := range all {
		if all[i].UID == uid {
			return &all[i], nil
		}
	}
	return nil, ErrTaskNotFound
}

func (d *DemoTaskRepository) GetAggregateTasks(
	ctx context.Context,
	selectedFilter string,
	isMonthly bool,
	dateISO, monthISO, todayISO string,
) (*models.AggregateResponse, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	dayNum := extractDayNumber(dateISO)
	monthNum := extractMonthNumber(monthISO)
	isFuture := dateISO > todayISO

	var filtered []demoTemplate
	roleSet := make(map[string]int)

	for _, tmpl := range demoTemplates {
		if selectedFilter != "" && selectedFilter != "All" && tmpl.TaskType != selectedFilter {
			continue
		}
		filtered = append(filtered, tmpl)
		roleSet[tmpl.Role] = len(demoPeopleByRole[tmpl.Role])
	}

	totalPeople := 0
	for _, c := range roleSet {
		totalPeople += c
	}

	items := make([]models.AggregateItemDTO, 0, len(filtered))
	for _, tmpl := range filtered {
		peopleCount := len(demoPeopleByRole[tmpl.Role])
		seedBase := dayNum
		if isMonthly {
			seedBase = monthNum
		}
		seed := seedBase + len(tmpl.ID)

		percentage := 55 + ((seed * 13) % 46)
		if percentage < 25 {
			percentage = 25
		}
		if percentage > 100 {
			percentage = 100
		}

		if !isMonthly && isFuture {
			percentage = (seed * 7) % 36
			if percentage < 0 {
				percentage = 0
			}
			if percentage > 35 {
				percentage = 35
			}
		}

		completed := int(math.Round((float64(peopleCount) * float64(percentage)) / 100.0))

		items = append(items, models.AggregateItemDTO{
			ID:         tmpl.ID,
			Name:       tmpl.Name,
			Role:       tmpl.Role,
			Category:   tmpl.Category,
			Frequency:  tmpl.Frequency,
			Time:       tmpl.Time,
			Misses:     tmpl.Misses,
			People:     peopleCount,
			Percentage: percentage,
			Completed:  completed,
		})
	}

	return &models.AggregateResponse{
		Items: items,
		Summary: models.AggregateSummary{
			UniqueTasks: len(filtered),
			RolesCount:  len(roleSet),
			PeopleCount: totalPeople,
		},
	}, nil
}

func (d *DemoTaskRepository) SubmitTask(
	ctx context.Context,
	uid, dateISO, todayISO, remarks string,
) (string, error) {
	if _, _, err := splitTaskUID(uid); err != nil {
		return "", err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	current := d.findInstanceLocked(uid, dateISO, todayISO)
	if current == nil {
		return "", ErrTaskNotFound
	}

	isFuture := dateISO > todayISO
	nextState := "submitted"
	eventTitle := "Task submitted (Temporary Demo Mode)"
	if isFuture {
		nextState = "submitted_early"
		eventTitle = "Submitted early (Temporary Demo Mode)"
	}

	ownerShort := models.BaseName(current.Owner)
	eventMeta := fmt.Sprintf("Submitted by %s · Waiting for %s", ownerShort, current.Verifier)
	if strings.TrimSpace(remarks) != "" {
		eventMeta = fmt.Sprintf("%s — %q", ownerShort, strings.TrimSpace(remarks))
	}

	d.nextLogID++
	newEntry := models.HistoryItem{
		ID:     strconv.Itoa(d.nextLogID),
		Title:  eventTitle,
		Meta:   eventMeta,
		Colour: "purple",
	}

	key := dateISO + "__" + uid
	subNotes := strings.TrimSpace(remarks)
	if subNotes == "" {
		subNotes = fmt.Sprintf("Demo submission filed by %s for %s (%s).", ownerShort, current.Name, dateISO)
	}
	d.overrides[key] = demoOverride{
		State:               nextState,
		Misses:              current.Misses,
		SubmissionNotes:     subNotes,
		VerificationRemarks: current.VerificationRemarks,
		History:             append([]models.HistoryItem{newEntry}, current.History...),
	}

	return nextState, nil
}

func (d *DemoTaskRepository) VerifyTask(
	ctx context.Context,
	uid, dateISO, todayISO, action, remarks string,
) (string, error) {
	if _, _, err := splitTaskUID(uid); err != nil {
		return "", err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	current := d.findInstanceLocked(uid, dateISO, todayISO)
	if current == nil {
		return "", ErrTaskNotFound
	}

	isFuture := dateISO > todayISO
	ownerShort := models.BaseName(current.Owner)
	cleanRemarks := strings.TrimSpace(remarks)

	var nextState, eventTitle, eventMeta, dotColor string

	if action == "approve" {
		if isFuture {
			nextState = "approved_early"
		} else {
			nextState = "approved"
		}
		eventTitle = "Verified & Closed (Temporary Demo Mode)"
		dotColor = "green"
		if cleanRemarks != "" {
			eventMeta = fmt.Sprintf("Approved by %s — %q", current.Verifier, cleanRemarks)
		} else {
			eventMeta = fmt.Sprintf("Approved by %s · Task closed", current.Verifier)
		}
	} else {
		if isFuture {
			nextState = "scheduled"
		} else {
			nextState = "overdue"
		}
		eventTitle = "Verification rejected (Temporary Demo Mode)"
		dotColor = "red"
		if cleanRemarks != "" {
			eventMeta = fmt.Sprintf("Returned by %s — %q", current.Verifier, cleanRemarks)
		} else {
			eventMeta = fmt.Sprintf("Returned to %s by %s for rework", ownerShort, current.Verifier)
		}
	}

	d.nextLogID++
	newEntry := models.HistoryItem{
		ID:     strconv.Itoa(d.nextLogID),
		Title:  eventTitle,
		Meta:   eventMeta,
		Colour: dotColor,
	}

	key := dateISO + "__" + uid
	d.overrides[key] = demoOverride{
		State:               nextState,
		Misses:              current.Misses,
		SubmissionNotes:     current.SubmissionNotes,
		VerificationRemarks: cleanRemarks,
		History:             append([]models.HistoryItem{newEntry}, current.History...),
	}

	return nextState, nil
}

func (d *DemoTaskRepository) EscalateTask(
	ctx context.Context,
	uid, dateISO, todayISO, recipient, reason string,
) error {
	if _, _, err := splitTaskUID(uid); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	current := d.findInstanceLocked(uid, dateISO, todayISO)
	if current == nil {
		return ErrTaskNotFound
	}

	target := strings.TrimSpace(recipient)
	if target == "" {
		target = current.Verifier
	}
	cleanReason := strings.TrimSpace(reason)

	eventMeta := fmt.Sprintf("Demo escalation logged for review by %s (no external email/SMS sent)", target)
	if cleanReason != "" {
		eventMeta = fmt.Sprintf("Escalated to %s — %q (demo log only)", target, cleanReason)
	}

	d.nextLogID++
	newEntry := models.HistoryItem{
		ID:     strconv.Itoa(d.nextLogID),
		Title:  "Escalation recorded (Temporary Demo Mode)",
		Meta:   eventMeta,
		Colour: "orange",
	}

	key := dateISO + "__" + uid
	d.overrides[key] = demoOverride{
		State:               current.State,
		Misses:              current.Misses,
		SubmissionNotes:     current.SubmissionNotes,
		VerificationRemarks: current.VerificationRemarks,
		History:             append([]models.HistoryItem{newEntry}, current.History...),
	}

	return nil
}

func (d *DemoTaskRepository) findInstanceLocked(uid, dateISO, todayISO string) *models.TaskInstanceDTO {
	all := d.buildInstancesLocked("role", "All", "All", dateISO, todayISO)
	for i := range all {
		if all[i].UID == uid {
			return &all[i]
		}
	}
	return nil
}

func (d *DemoTaskRepository) buildInstancesLocked(
	mode, selectedFilter, selectedPerson, dateISO, todayISO string,
) []models.TaskInstanceDTO {
	dayNum := extractDayNumber(dateISO)
	isPast := dateISO < todayISO
	isFuture := dateISO > todayISO

	var rows []models.TaskInstanceDTO
	instanceCounter := 0

	for _, tmpl := range demoTemplates {
		if mode == "role" && selectedFilter != "" && selectedFilter != "All" && tmpl.Role != selectedFilter {
			continue
		}
		if mode == "type" && selectedFilter != "" && selectedFilter != "All" && tmpl.TaskType != selectedFilter {
			continue
		}

		allOwners := demoPeopleByRole[tmpl.Role]
		owners := allOwners
		if mode == "role" && selectedFilter != "" && selectedFilter != "All" && selectedPerson != "" && selectedPerson != "All" {
			owners = []string{selectedPerson}
		}

		for idx, owner := range owners {
			instanceCounter++
			uid := tmpl.ID + "_" + owner
			key := dateISO + "__" + uid
			verifier := demoVerifierByRole[tmpl.Role]

			var state string
			misses := tmpl.Misses
			var subNotes, verRemarks string
			var history []models.HistoryItem

			if ov, ok := d.overrides[key]; ok {
				state = ov.State
				misses = ov.Misses
				subNotes = ov.SubmissionNotes
				verRemarks = ov.VerificationRemarks
				history = ov.History
			} else {
				state = computeDemoState(tmpl, idx, dayNum, isFuture)
				history = buildDemoDefaultHistory(tmpl, misses, verifier)
				if state == "submitted" || state == "approved" || state == "submitted_early" || state == "approved_early" {
					subNotes = fmt.Sprintf(
						"Demo submission summary for %q filed by %s (Role: %s, Scheduled Date: %s). All checklist items reported complete for cycle review.",
						tmpl.Name,
						models.BaseName(owner),
						tmpl.Role,
						dateISO,
					)
				}
				if state == "approved" || state == "approved_early" {
					verRemarks = fmt.Sprintf("Verified and closed by %s (Demo sample record).", verifier)
				}
			}

			effective := state
			if isPast && state != "approved" {
				effective = "rolled"
			}

			rows = append(rows, models.TaskInstanceDTO{
				InstanceID:          instanceCounter,
				ID:                  tmpl.ID,
				UID:                 uid,
				Name:                tmpl.Name,
				Role:                tmpl.Role,
				Category:            tmpl.Category,
				Frequency:           tmpl.Frequency,
				Time:                tmpl.Time,
				Misses:              misses,
				Owner:               owner,
				Verifier:            verifier,
				State:               state,
				Effective:           effective,
				DateISO:             dateISO,
				SubmissionNotes:     subNotes,
				VerificationRemarks: verRemarks,
				Flow:                models.BuildFlow(owner, verifier, state),
				History:             history,
			})
		}
	}

	if rows == nil {
		return []models.TaskInstanceDTO{}
	}
	return rows
}

func computeDemoState(tmpl demoTemplate, index, dayNum int, isFuture bool) string {
	seed := (dayNum + index + len(tmpl.ID)) % 7
	if isFuture {
		if seed == 0 {
			return "approved_early"
		}
		if seed == 1 {
			return "submitted_early"
		}
		return "scheduled"
	}

	switch tmpl.Pattern {
	case "approved":
		return "approved"
	case "submitted":
		if seed%2 == 0 {
			return "submitted"
		}
		return "approved"
	case "overdue":
		if seed == 0 {
			return "approved"
		}
		return "overdue"
	default:
		return "submitted"
	}
}

func buildDemoDefaultHistory(tmpl demoTemplate, misses int, verifier string) []models.HistoryItem {
	items := []models.HistoryItem{
		{
			ID:     tmpl.ID + "_h1",
			Title:  "Previous cycle completed",
			Meta:   "Submission received from task owner",
			Colour: "green",
		},
		{
			ID:     tmpl.ID + "_h2",
			Title:  "Verification completed",
			Meta:   "Reviewed by " + verifier,
			Colour: "green",
		},
	}

	if misses > 0 {
		items = append(items, models.HistoryItem{
			ID:     tmpl.ID + "_h3",
			Title:  "Historical delay",
			Meta:   fmt.Sprintf("%d previous missed / delayed cycle(s)", misses),
			Colour: "red",
		})
	}

	items = append(items, models.HistoryItem{
		ID:     tmpl.ID + "_h4",
		Title:  "Recurring responsibility",
		Meta:   fmt.Sprintf("%s task mapped to %s", tmpl.Frequency, tmpl.Role),
		Colour: "gray",
	})

	return items
}

func extractDayNumber(dateISO string) int {
	t, err := time.Parse("2006-01-02", dateISO)
	if err != nil {
		return 11
	}
	return t.Day()
}

func extractMonthNumber(monthISO string) int {
	t, err := time.Parse("2006-01", monthISO)
	if err != nil {
		return 10
	}
	return int(t.Month())
}

