package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"organisation-task-control-centre/backend/internal/models"
)

// ErrTaskNotFound is returned when a requested task instance cannot be found.
var ErrTaskNotFound = errors.New("task instance not found")

var orderedTaskTypes = []string{
	"Daily Tasks",
	"Weekly Tasks",
	"Monthly Tasks",
	"Quarterly Tasks",
	"Yearly Tasks",
	"One Time Tasks",
	"Statutory Tasks",
	"Finance Tasks",
	"Academic Tasks",
	"Others",
}

// TaskRepository executes parameterized SQL queries against MySQL.
type TaskRepository struct {
	db               *sql.DB
	defaultDemoToday string
}

// NewTaskRepository creates a new TaskRepository and ensures any optional columns
// expected by the API exist on older schemas (e.g., if created from database.sql).
func NewTaskRepository(db *sql.DB, defaultDemoToday string) *TaskRepository {
	r := &TaskRepository{
		db:               db,
		defaultDemoToday: defaultDemoToday,
	}
	r.ensureCompatibleColumns(context.Background())
	return r
}

// ensureCompatibleColumns safely adds non-breaking columns if the database was
// created from the initial `database.sql` file instead of `database/001_schema.sql`.
// It never drops tables or deletes existing data.
func (r *TaskRepository) ensureCompatibleColumns(ctx context.Context) {
	_, _ = r.db.ExecContext(ctx, `
		ALTER TABLE task_templates
		ADD COLUMN IF NOT EXISTS default_pattern ENUM('submitted', 'approved', 'overdue') NOT NULL DEFAULT 'submitted'
	`)
	_, _ = r.db.ExecContext(ctx, `
		ALTER TABLE task_instances
		ADD COLUMN IF NOT EXISTS verification_remarks TEXT NULL
	`)
}

// GetMetadata retrieves roles, users, task types, and top-level organisation counters.
func (r *TaskRepository) GetMetadata(ctx context.Context) (*models.MetadataResponse, error) {
	// 1. Load Roles
	roleRows, err := r.db.QueryContext(ctx, `
		SELECT id, name, verifier_title
		FROM roles
		ORDER BY display_order ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer roleRows.Close()

	type roleRecord struct {
		ID       int
		Name     string
		Verifier string
	}
	var roleList []roleRecord
	verifierByRole := make(map[string]string)
	peopleByRole := make(map[string][]string)

	for roleRows.Next() {
		var rec roleRecord
		if err := roleRows.Scan(&rec.ID, &rec.Name, &rec.Verifier); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roleList = append(roleList, rec)
		verifierByRole[rec.Name] = rec.Verifier
		peopleByRole[rec.Name] = []string{}
	}
	if err := roleRows.Err(); err != nil {
		return nil, err
	}

	// 2. Load Active Users per Role
	userRows, err := r.db.QueryContext(ctx, `
		SELECT r.name, u.display_label
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.is_active = TRUE
		ORDER BY u.id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer userRows.Close()

	for userRows.Next() {
		var roleName, displayLabel string
		if err := userRows.Scan(&roleName, &displayLabel); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		peopleByRole[roleName] = append(peopleByRole[roleName], displayLabel)
	}
	if err := userRows.Err(); err != nil {
		return nil, err
	}

	var rolesMeta []models.RoleMetadata
	for _, rec := range roleList {
		people := peopleByRole[rec.Name]
		rolesMeta = append(rolesMeta, models.RoleMetadata{
			Name:     rec.Name,
			Count:    len(people),
			People:   people,
			Verifier: rec.Verifier,
		})
	}

	// 3. Load Task Type Counts from task_templates
	typeRows, err := r.db.QueryContext(ctx, `
		SELECT task_type, COUNT(*) AS cnt
		FROM task_templates
		WHERE is_active = TRUE
		GROUP BY task_type
	`)
	if err != nil {
		return nil, fmt.Errorf("query task types: %w", err)
	}
	defer typeRows.Close()

	typeCounts := make(map[string]int)
	for typeRows.Next() {
		var tName string
		var cnt int
		if err := typeRows.Scan(&tName, &cnt); err != nil {
			return nil, fmt.Errorf("scan task type: %w", err)
		}
		typeCounts[tName] = cnt
	}
	if err := typeRows.Err(); err != nil {
		return nil, err
	}

	var taskTypesMeta []models.TaskTypeMetadata
	for _, tName := range orderedTaskTypes {
		taskTypesMeta = append(taskTypesMeta, models.TaskTypeMetadata{
			Name:  tName,
			Count: typeCounts[tName],
		})
	}

	// 4. Compute Organisation Summary from stored instances on defaultDemoToday
	summary := models.OrganisationSummary{
		ActiveTasks: 86,
		Overdue:     9,
		RepeatMiss:  6,
	}

	var totalCount, overdueCount, repeatMissCount sql.NullInt64
	err = r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			SUM(CASE WHEN status = 'overdue' THEN 1 ELSE 0 END),
			SUM(CASE WHEN past_misses >= 2 THEN 1 ELSE 0 END)
		FROM task_instances
		WHERE scheduled_date = ?
	`, r.defaultDemoToday).Scan(&totalCount, &overdueCount, &repeatMissCount)
	if err == nil && totalCount.Int64 > 0 {
		summary.ActiveTasks = int(totalCount.Int64)
		summary.Overdue = int(overdueCount.Int64)
		summary.RepeatMiss = int(repeatMissCount.Int64)
	}

	return &models.MetadataResponse{
		OrganisationSummary: summary,
		Roles:               rolesMeta,
		PeopleByRole:        peopleByRole,
		VerifierByRole:      verifierByRole,
		TaskTypes:           taskTypesMeta,
	}, nil
}

// EnsureDateInstancesmaterializes scheduled task instances in MySQL for a newly
// visited date if none exist yet, so every date has real stored database rows.
func (r *TaskRepository) EnsureDateInstances(ctx context.Context, dateISO, todayISO string) error {
	var existing int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM task_instances WHERE scheduled_date = ?
	`, dateISO).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	// Materialize instances for dateISO from active task_templates and users
	_, err := r.db.ExecContext(ctx, `
		INSERT IGNORE INTO task_instances (
			template_id, owner_id, scheduled_date, status, past_misses
		)
		SELECT
			tt.id,
			u.id,
			?,
			CASE
				WHEN ? > ? THEN 'scheduled'
				WHEN tt.default_pattern = 'approved' THEN 'approved'
				WHEN tt.default_pattern = 'overdue' THEN 'overdue'
				ELSE 'submitted'
			END,
			tt.default_misses
		FROM task_templates tt
		JOIN users u ON u.role_id = tt.role_id
		WHERE tt.is_active = TRUE AND u.is_active = TRUE
	`, dateISO, dateISO, todayISO)
	if err != nil {
		return fmt.Errorf("materialize instances for %s: %w", dateISO, err)
	}

	// Seed baseline activity history rows for newly materialized instances
	_, _ = r.db.ExecContext(ctx, `
		INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
		SELECT
			ti.id,
			'Previous cycle completed',
			'Submission received from task owner',
			'green',
			CONCAT(ti.scheduled_date, ' 08:00:00')
		FROM task_instances ti
		LEFT JOIN task_activity_logs tal ON tal.task_instance_id = ti.id
		WHERE ti.scheduled_date = ? AND tal.id IS NULL
	`, dateISO)

	return nil
}

// ListTasks queries stored task instances for the selected date, role/type filter, and person.
func (r *TaskRepository) ListTasks(
	ctx context.Context,
	mode, selectedFilter, selectedPerson, dateISO, todayISO string,
) ([]models.TaskInstanceDTO, error) {
	if err := r.EnsureDateInstances(ctx, dateISO, todayISO); err != nil {
		return nil, err
	}

	query := `
		SELECT
			ti.id,
			tt.code,
			tt.name,
			r.name AS role_name,
			tt.category,
			tt.frequency,
			tt.deadline_label,
			ti.past_misses,
			u.display_label AS owner_label,
			r.verifier_title,
			ti.status,
			DATE_FORMAT(ti.scheduled_date, '%Y-%m-%d') AS sched_date,
			COALESCE(ti.submission_notes, '') AS sub_notes,
			COALESCE(ti.verification_remarks, '') AS ver_remarks
		FROM task_instances ti
		JOIN task_templates tt ON tt.id = ti.template_id
		JOIN roles r ON r.id = tt.role_id
		JOIN users u ON u.id = ti.owner_id
		WHERE ti.scheduled_date = ?
	`
	args := []interface{}{dateISO}

	if mode == "role" && selectedFilter != "" && selectedFilter != "All" {
		query += ` AND r.name = ?`
		args = append(args, selectedFilter)
		if selectedPerson != "" && selectedPerson != "All" {
			query += ` AND u.display_label = ?`
			args = append(args, selectedPerson)
		}
	} else if mode == "type" && selectedFilter != "" && selectedFilter != "All" {
		query += ` AND tt.task_type = ?`
		args = append(args, selectedFilter)
	}

	query += ` ORDER BY tt.id ASC, u.id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query task instances: %w", err)
	}
	defer rows.Close()

	isPast := dateISO < todayISO
	isFuture := dateISO > todayISO

	var tasks []models.TaskInstanceDTO
	var instanceIDs []int
	indexByInstanceID := make(map[int]int)

	for rows.Next() {
		var dto models.TaskInstanceDTO
		var rawStatus string
		if err := rows.Scan(
			&dto.InstanceID,
			&dto.ID,
			&dto.Name,
			&dto.Role,
			&dto.Category,
			&dto.Frequency,
			&dto.Time,
			&dto.Misses,
			&dto.Owner,
			&dto.Verifier,
			&rawStatus,
			&dto.DateISO,
			&dto.SubmissionNotes,
			&dto.VerificationRemarks,
		); err != nil {
			return nil, fmt.Errorf("scan task instance: %w", err)
		}

		// Normalize status if user shifted the configurable Demo Today date relative to scheduled_date
		state := normalizeStateForDemoToday(rawStatus, isFuture)
		if dto.SubmissionNotes == "" && (state == "submitted" || state == "approved" || state == "submitted_early" || state == "approved_early") {
			dto.SubmissionNotes = fmt.Sprintf(
				"Demo submission summary for %q filed by %s (Role: %s, Scheduled Date: %s). All checklist items reported complete for cycle review.",
				dto.Name,
				models.BaseName(dto.Owner),
				dto.Role,
				dto.DateISO,
			)
		}
		effective := state
		if isPast && state != "approved" {
			effective = "rolled"
		}

		dto.UID = dto.ID + "_" + dto.Owner
		dto.State = state
		dto.Effective = effective
		dto.Flow = models.BuildFlow(dto.Owner, dto.Verifier, state)
		dto.History = []models.HistoryItem{}

		indexByInstanceID[dto.InstanceID] = len(tasks)
		instanceIDs = append(instanceIDs, dto.InstanceID)
		tasks = append(tasks, dto)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(instanceIDs) == 0 {
		return []models.TaskInstanceDTO{}, nil
	}

	// Batch-load activity history logs for all returned task instances
	placeholders := make([]string, len(instanceIDs))
	histArgs := make([]interface{}, len(instanceIDs))
	for i, id := range instanceIDs {
		placeholders[i] = "?"
		histArgs[i] = id
	}

	histQuery := fmt.Sprintf(`
		SELECT id, task_instance_id, event_title, event_meta, dot_color
		FROM task_activity_logs
		WHERE task_instance_id IN (%s)
		ORDER BY created_at DESC, id DESC
	`, strings.Join(placeholders, ","))

	histRows, err := r.db.QueryContext(ctx, histQuery, histArgs...)
	if err != nil {
		return nil, fmt.Errorf("query task history: %w", err)
	}
	defer histRows.Close()

	for histRows.Next() {
		var logID, instID int
		var item models.HistoryItem
		if err := histRows.Scan(&logID, &instID, &item.Title, &item.Meta, &item.Colour); err != nil {
			return nil, fmt.Errorf("scan history item: %w", err)
		}
		item.ID = strconv.Itoa(logID)
		if idx, ok := indexByInstanceID[instID]; ok {
			tasks[idx].History = append(tasks[idx].History, item)
		}
	}

	return tasks, histRows.Err()
}

// GetTaskByUID fetches a single task instance and its activity history by UID (`t1_Dr. Priya Raman`) and date.
func (r *TaskRepository) GetTaskByUID(ctx context.Context, uid, dateISO, todayISO string) (*models.TaskInstanceDTO, error) {
	code, ownerLabel, err := splitTaskUID(uid)
	if err != nil {
		return nil, err
	}

	if err := r.EnsureDateInstances(ctx, dateISO, todayISO); err != nil {
		return nil, err
	}

	var dto models.TaskInstanceDTO
	var rawStatus string
	err = r.db.QueryRowContext(ctx, `
		SELECT
			ti.id,
			tt.code,
			tt.name,
			r.name AS role_name,
			tt.category,
			tt.frequency,
			tt.deadline_label,
			ti.past_misses,
			u.display_label AS owner_label,
			r.verifier_title,
			ti.status,
			DATE_FORMAT(ti.scheduled_date, '%Y-%m-%d') AS sched_date,
			COALESCE(ti.submission_notes, '') AS sub_notes,
			COALESCE(ti.verification_remarks, '') AS ver_remarks
		FROM task_instances ti
		JOIN task_templates tt ON tt.id = ti.template_id
		JOIN roles r ON r.id = tt.role_id
		JOIN users u ON u.id = ti.owner_id
		WHERE tt.code = ? AND u.display_label = ? AND ti.scheduled_date = ?
	`, code, ownerLabel, dateISO).Scan(
		&dto.InstanceID,
		&dto.ID,
		&dto.Name,
		&dto.Role,
		&dto.Category,
		&dto.Frequency,
		&dto.Time,
		&dto.Misses,
		&dto.Owner,
		&dto.Verifier,
		&rawStatus,
		&dto.DateISO,
		&dto.SubmissionNotes,
		&dto.VerificationRemarks,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task by uid: %w", err)
	}

	isPast := dateISO < todayISO
	isFuture := dateISO > todayISO
	state := normalizeStateForDemoToday(rawStatus, isFuture)
	effective := state
	if isPast && state != "approved" {
		effective = "rolled"
	}

	dto.UID = uid
	dto.State = state
	dto.Effective = effective
	dto.Flow = models.BuildFlow(dto.Owner, dto.Verifier, state)

	history, err := r.GetTaskHistory(ctx, dto.InstanceID)
	if err != nil {
		return nil, err
	}
	dto.History = history

	return &dto, nil
}

// GetTaskHistory returns the chronological activity log for a task instance ID.
func (r *TaskRepository) GetTaskHistory(ctx context.Context, instanceID int) ([]models.HistoryItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_title, event_meta, dot_color
		FROM task_activity_logs
		WHERE task_instance_id = ?
		ORDER BY created_at DESC, id DESC
	`, instanceID)
	if err != nil {
		return nil, fmt.Errorf("query history for instance %d: %w", instanceID, err)
	}
	defer rows.Close()

	history := []models.HistoryItem{}
	for rows.Next() {
		var id int
		var item models.HistoryItem
		if err := rows.Scan(&id, &item.Title, &item.Meta, &item.Colour); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		item.ID = strconv.Itoa(id)
		history = append(history, item)
	}
	return history, rows.Err()
}

// GetAggregateTasks calculates completion metrics across roles from stored task instances.
func (r *TaskRepository) GetAggregateTasks(
	ctx context.Context,
	selectedFilter string,
	isMonthly bool,
	dateISO, monthISO, todayISO string,
) (*models.AggregateResponse, error) {
	if !isMonthly {
		if err := r.EnsureDateInstances(ctx, dateISO, todayISO); err != nil {
			return nil, err
		}
	}

	query := `
		SELECT
			tt.id,
			tt.code,
			tt.name,
			r.name AS role_name,
			tt.category,
			tt.frequency,
			tt.deadline_label,
			tt.default_misses,
			(SELECT COUNT(*) FROM users u WHERE u.role_id = r.id AND u.is_active = TRUE) AS people_count
		FROM task_templates tt
		JOIN roles r ON r.id = tt.role_id
		WHERE tt.is_active = TRUE
	`
	var args []interface{}
	if selectedFilter != "" && selectedFilter != "All" {
		query += ` AND tt.task_type = ?`
		args = append(args, selectedFilter)
	}
	query += ` ORDER BY tt.id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query aggregate templates: %w", err)
	}
	defer rows.Close()

	type tmplRow struct {
		TemplateID  int
		Code        string
		Name        string
		Role        string
		Category    string
		Frequency   string
		Time        string
		Misses      int
		PeopleCount int
	}

	var templates []tmplRow
	roleSet := make(map[string]int)

	for rows.Next() {
		var t tmplRow
		if err := rows.Scan(
			&t.TemplateID,
			&t.Code,
			&t.Name,
			&t.Role,
			&t.Category,
			&t.Frequency,
			&t.Time,
			&t.Misses,
			&t.PeopleCount,
		); err != nil {
			return nil, fmt.Errorf("scan aggregate template: %w", err)
		}
		templates = append(templates, t)
		roleSet[t.Role] = t.PeopleCount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPeople := 0
	for _, cnt := range roleSet {
		totalPeople += cnt
	}

	items := make([]models.AggregateItemDTO, 0, len(templates))
	for _, t := range templates {
		var totalInstances, approvedInstances int
		if isMonthly {
			_ = r.db.QueryRowContext(ctx, `
				SELECT
					COUNT(*),
					COALESCE(SUM(CASE WHEN status IN ('approved', 'approved_early') THEN 1 ELSE 0 END), 0)
				FROM task_instances
				WHERE template_id = ? AND DATE_FORMAT(scheduled_date, '%Y-%m') = ?
			`, t.TemplateID, monthISO).Scan(&totalInstances, &approvedInstances)
		} else {
			_ = r.db.QueryRowContext(ctx, `
				SELECT
					COUNT(*),
					COALESCE(SUM(CASE WHEN status IN ('approved', 'approved_early') THEN 1 ELSE 0 END), 0)
				FROM task_instances
				WHERE template_id = ? AND scheduled_date = ?
			`, t.TemplateID, dateISO).Scan(&totalInstances, &approvedInstances)
		}

		percentage := 0
		completedPeople := 0
		if totalInstances > 0 {
			percentage = int(math.Round((float64(approvedInstances) / float64(totalInstances)) * 100.0))
			completedPeople = int(math.Round((float64(t.PeopleCount) * float64(percentage)) / 100.0))
		}

		items = append(items, models.AggregateItemDTO{
			ID:         t.Code,
			Name:       t.Name,
			Role:       t.Role,
			Category:   t.Category,
			Frequency:  t.Frequency,
			Time:       t.Time,
			Misses:     t.Misses,
			People:     t.PeopleCount,
			Percentage: percentage,
			Completed:  completedPeople,
		})
	}

	return &models.AggregateResponse{
		Items: items,
		Summary: models.AggregateSummary{
			UniqueTasks: len(templates),
			RolesCount:  len(roleSet),
			PeopleCount: totalPeople,
		},
	}, nil
}

// SubmitTask transitions a task instance to `submitted` (or `submitted_early` if scheduled_date > todayISO)
// and appends an audit entry inside a single database transaction.
// Note: A submitted task is NOT marked closed/completed until approved by its verifier.
func (r *TaskRepository) SubmitTask(
	ctx context.Context,
	uid, dateISO, todayISO, remarks string,
) (string, error) {
	code, ownerLabel, err := splitTaskUID(uid)
	if err != nil {
		return "", err
	}

	if err := r.EnsureDateInstances(ctx, dateISO, todayISO); err != nil {
		return "", err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var instanceID int
	var verifierTitle string
	err = tx.QueryRowContext(ctx, `
		SELECT ti.id, r.verifier_title
		FROM task_instances ti
		JOIN task_templates tt ON tt.id = ti.template_id
		JOIN roles r ON r.id = tt.role_id
		JOIN users u ON u.id = ti.owner_id
		WHERE tt.code = ? AND u.display_label = ? AND ti.scheduled_date = ?
		FOR UPDATE
	`, code, ownerLabel, dateISO).Scan(&instanceID, &verifierTitle)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTaskNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lock task instance for submit: %w", err)
	}

	isFuture := dateISO > todayISO
	nextState := "submitted"
	eventTitle := "Task submitted (Demo Action)"
	if isFuture {
		nextState = "submitted_early"
		eventTitle = "Submitted early (Demo Action)"
	}

	ownerShort := models.BaseName(ownerLabel)
	eventMeta := fmt.Sprintf("Submitted by %s · Waiting for %s", ownerShort, verifierTitle)
	if strings.TrimSpace(remarks) != "" {
		eventMeta = fmt.Sprintf("%s — %q", ownerShort, strings.TrimSpace(remarks))
	}

	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		UPDATE task_instances
		SET status = ?,
		    submitted_at = ?,
		    submission_notes = ?,
		    closed_at = NULL
		WHERE id = ?
	`, nextState, now, strings.TrimSpace(remarks), instanceID)
	if err != nil {
		return "", fmt.Errorf("update task status on submit: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
		VALUES (?, ?, ?, 'purple', ?)
	`, instanceID, eventTitle, eventMeta, now)
	if err != nil {
		return "", fmt.Errorf("insert submit activity log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit submit transaction: %w", err)
	}

	return nextState, nil
}

// VerifyTask either approves (closing the task) or rejects (returning the task to the owner)
// a submitted task inside a single database transaction.
func (r *TaskRepository) VerifyTask(
	ctx context.Context,
	uid, dateISO, todayISO, action, remarks string,
) (string, error) {
	code, ownerLabel, err := splitTaskUID(uid)
	if err != nil {
		return "", err
	}

	if err := r.EnsureDateInstances(ctx, dateISO, todayISO); err != nil {
		return "", err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var instanceID int
	var verifierTitle string
	err = tx.QueryRowContext(ctx, `
		SELECT ti.id, r.verifier_title
		FROM task_instances ti
		JOIN task_templates tt ON tt.id = ti.template_id
		JOIN roles r ON r.id = tt.role_id
		JOIN users u ON u.id = ti.owner_id
		WHERE tt.code = ? AND u.display_label = ? AND ti.scheduled_date = ?
		FOR UPDATE
	`, code, ownerLabel, dateISO).Scan(&instanceID, &verifierTitle)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTaskNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lock task instance for verify: %w", err)
	}

	isFuture := dateISO > todayISO
	ownerShort := models.BaseName(ownerLabel)
	now := time.Now()
	cleanRemarks := strings.TrimSpace(remarks)

	var nextState, eventTitle, eventMeta, dotColor string

	if action == "approve" {
		if isFuture {
			nextState = "approved_early"
		} else {
			nextState = "approved"
		}
		eventTitle = "Verified & Closed (Demo Action)"
		dotColor = "green"
		if cleanRemarks != "" {
			eventMeta = fmt.Sprintf("Approved by %s — %q", verifierTitle, cleanRemarks)
		} else {
			eventMeta = fmt.Sprintf("Approved by %s · Task closed", verifierTitle)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE task_instances
			SET status = ?,
			    verified_at = ?,
			    verified_by = ?,
			    closed_at = ?,
			    verification_remarks = ?
			WHERE id = ?
		`, nextState, now, verifierTitle, now, cleanRemarks, instanceID)
	} else {
		// Reject returns the task to owner-pending state (overdue on today/past, scheduled on future)
		if isFuture {
			nextState = "scheduled"
		} else {
			nextState = "overdue"
		}
		eventTitle = "Verification rejected (Demo Action)"
		dotColor = "red"
		if cleanRemarks != "" {
			eventMeta = fmt.Sprintf("Returned by %s — %q", verifierTitle, cleanRemarks)
		} else {
			eventMeta = fmt.Sprintf("Returned to %s by %s for rework", ownerShort, verifierTitle)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE task_instances
			SET status = ?,
			    verified_at = NULL,
			    verified_by = NULL,
			    closed_at = NULL,
			    verification_remarks = ?
			WHERE id = ?
		`, nextState, cleanRemarks, instanceID)
	}
	if err != nil {
		return "", fmt.Errorf("update task status on verify: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, instanceID, eventTitle, eventMeta, dotColor, now)
	if err != nil {
		return "", fmt.Errorf("insert verify activity log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit verify transaction: %w", err)
	}

	return nextState, nil
}

// EscalateTask records a demo-only escalation entry in task_activity_logs without sending external emails/SMS.
func (r *TaskRepository) EscalateTask(
	ctx context.Context,
	uid, dateISO, todayISO, recipient, reason string,
) error {
	code, ownerLabel, err := splitTaskUID(uid)
	if err != nil {
		return err
	}

	if err := r.EnsureDateInstances(ctx, dateISO, todayISO); err != nil {
		return err
	}

	var instanceID int
	var verifierTitle string
	err = r.db.QueryRowContext(ctx, `
		SELECT ti.id, r.verifier_title
		FROM task_instances ti
		JOIN task_templates tt ON tt.id = ti.template_id
		JOIN roles r ON r.id = tt.role_id
		JOIN users u ON u.id = ti.owner_id
		WHERE tt.code = ? AND u.display_label = ? AND ti.scheduled_date = ?
	`, code, ownerLabel, dateISO).Scan(&instanceID, &verifierTitle)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("find task instance for escalate: %w", err)
	}

	target := strings.TrimSpace(recipient)
	if target == "" {
		target = verifierTitle
	}
	cleanReason := strings.TrimSpace(reason)

	eventMeta := fmt.Sprintf("Demo escalation logged for review by %s (no external email/SMS sent)", target)
	if cleanReason != "" {
		eventMeta = fmt.Sprintf("Escalated to %s — %q (demo log only)", target, cleanReason)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
		VALUES (?, 'Escalation recorded (Demo Action)', ?, 'orange', ?)
	`, instanceID, eventMeta, time.Now())
	return err
}

func splitTaskUID(uid string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(uid), "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid task uid format")
	}
	return parts[0], parts[1], nil
}

// normalizeStateForDemoToday ensures that if the user changes the configurable
// "Demo Today" picker in the UI, stored states remain valid for that relative period
// (e.g., future dates cannot have overdue states).
func normalizeStateForDemoToday(rawStatus string, isFuture bool) string {
	if isFuture {
		switch rawStatus {
		case "approved", "approved_early":
			return "approved_early"
		case "submitted", "submitted_early":
			return "submitted_early"
		default:
			return "scheduled"
		}
	}

	switch rawStatus {
	case "approved_early":
		return "approved"
	case "submitted_early":
		return "submitted"
	case "scheduled":
		return "overdue"
	default:
		return rawStatus
	}
}
