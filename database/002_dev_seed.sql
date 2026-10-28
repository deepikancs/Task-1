-- ============================================================================
-- ORGANISATION TASK CONTROL CENTRE
-- Migration 002: OPTIONAL DEVELOPMENT SEED DATA (DEMO ONLY)
-- ============================================================================
-- IMPORTANT NOTICE:
--   - This file contains synthetic demo data from "Overall Task View.html"
--     for local development and demonstration only.
--   - Do NOT run this file in a production environment or treat its names,
--     roles, or sample schedules as confirmed business rules.
--   - Generates actual stored rows in `task_instances` and `task_activity_logs`
--     across a 16-day window (-7 to +8 days around 2026-10-11).
-- ============================================================================

USE Diviga;

-- 1. Seed Roles
INSERT INTO roles (id, name, verifier_title, display_order) VALUES
  (1,  'Principal',               'Chairman / Management',       1),
  (2,  'IQAC Head',               'Principal',                   2),
  (3,  'Head of Department',      'Principal / Dean Academics',  3),
  (4,  'Lab Incharge',            'Head of Department',          4),
  (5,  'Transport Incharge',      'Admin Head',                  5),
  (6,  'Maintenance Head',        'Admin Head',                  6),
  (7,  'Admin Head',              'Principal',                   7),
  (8,  'Finance Head',            'Management',                  8),
  (9,  'Placement Head',          'Principal / Dean Placements', 9),
  (10, 'R&D Head',                'Principal / Dean R&D',        10),
  (11, 'Industry Relations Head', 'Principal',                   11),
  (12, 'Alumni Head',             'Principal',                   12)
ON DUPLICATE KEY UPDATE
  verifier_title = VALUES(verifier_title),
  display_order = VALUES(display_order);

-- 2. Seed Synthetic Demo Users (Task Owners)
INSERT INTO users (id, full_name, department, display_label, role_id) VALUES
  (1,  'Dr. Priya Raman',  NULL,                     'Dr. Priya Raman',                         1),
  (2,  'Dr. Senthil Kumar',NULL,                     'Dr. Senthil Kumar',                       2),
  (3,  'Dr. Meena S',      'CSE',                    'Dr. Meena S — CSE',                       3),
  (4,  'Dr. Revathi P',    'ECE',                    'Dr. Revathi P — ECE',                     3),
  (5,  'Dr. Anand K',      'Mechanical',             'Dr. Anand K — Mechanical',                3),
  (6,  'Dr. Deepika R',    'IT',                     'Dr. Deepika R — IT',                      3),
  (7,  'Dr. Vignesh M',    'EEE',                    'Dr. Vignesh M — EEE',                     3),
  (8,  'Mr. Sathish K',    'Advanced Manufacturing', 'Mr. Sathish K — Advanced Manufacturing',  4),
  (9,  'Ms. Harini V',     'AI Lab',                 'Ms. Harini V — AI Lab',                   4),
  (10, 'Mr. Naveen P',     'Electronics Lab',        'Mr. Naveen P — Electronics Lab',          4),
  (11, 'Mr. Suresh Kumar', NULL,                     'Mr. Suresh Kumar',                        5),
  (12, 'Mr. Ramesh B',     NULL,                     'Mr. Ramesh B',                            6),
  (13, 'Ms. Anitha R',     NULL,                     'Ms. Anitha R',                            7),
  (14, 'Mr. Vishal K',     NULL,                     'Mr. Vishal K',                            7),
  (15, 'Mr. Kannan R',     NULL,                     'Mr. Kannan R',                            8),
  (16, 'Ms. Kavitha S',    NULL,                     'Ms. Kavitha S',                           9),
  (17, 'Mr. Arun P',       NULL,                     'Mr. Arun P',                              9),
  (18, 'Dr. Mani V',       NULL,                     'Dr. Mani V',                              10),
  (19, 'Dr. Nirmala S',    NULL,                     'Dr. Nirmala S',                           10),
  (20, 'Ms. Deepa K',      NULL,                     'Ms. Deepa K',                             11),
  (21, 'Dr. Lavanya P',    NULL,                     'Dr. Lavanya P',                           12)
ON DUPLICATE KEY UPDATE
  full_name = VALUES(full_name),
  department = VALUES(department),
  role_id = VALUES(role_id);

-- 3. Seed Task Templates (Preserving Prototype Task-Type Mapping)
INSERT INTO task_templates (
  id, code, name, role_id, category, frequency, task_type, deadline_label, default_pattern, default_misses
) VALUES
  (1,  't1',  'Daily Teaching-Learning Closure',             1,  'Teaching & Learning',        'Daily Tasks',     'Academic Tasks',  '5:30 PM',     'submitted', 1),
  (2,  't2',  'Academic Deviation & Faculty Absence Review', 1,  'Teaching & Learning',        'Daily Tasks',     'Academic Tasks',  '4:45 PM',     'approved',  0),
  (3,  't3',  'Student Discipline Exception Closure',        1,  'Student Discipline',         'Daily Tasks',     'Daily Tasks',     '5:00 PM',     'overdue',   2),
  (4,  't4',  'Accreditation Evidence Update',               2,  'Documentation',              'Monthly Tasks',   'Monthly Tasks',   'Month End',   'submitted', 0),
  (5,  't5',  'Internal Quality Audit Review',               2,  'Statutory Committees',       'Quarterly Tasks', 'Statutory Tasks', 'Quarter End', 'approved',  0),
  (6,  't6',  'Department Teaching Progress Review',         3,  'Teaching & Learning',        'Daily Tasks',     'Academic Tasks',  '5:00 PM',     'submitted', 0),
  (7,  't7',  'Student Competency Gap Follow-up',            3,  'Students Competency',        'Weekly Tasks',    'Academic Tasks',  'Friday',      'overdue',   2),
  (8,  't8',  'Faculty Action Closure',                      3,  'Faculty Achievement',        'Daily Tasks',     'Daily Tasks',     '5:30 PM',     'approved',  0),
  (9,  't9',  'Department Monthly Performance Review',       3,  'Data Management',            'Monthly Tasks',   'Monthly Tasks',   '10th',        'submitted', 1),
  (10, 't10', 'Laboratory Utilisation Closure',              4,  'Infrastructure Utilisation', 'Daily Tasks',     'Daily Tasks',     '5:00 PM',     'approved',  0),
  (11, 't11', 'Equipment Maintenance Exception Review',      4,  'Infrastructure Maintenance', 'Weekly Tasks',    'Weekly Tasks',    'Saturday',    'submitted', 0),
  (12, 't12', 'Bus Route Readiness Confirmation',            5,  'Infrastructure Maintenance', 'Daily Tasks',     'Daily Tasks',     '6:30 AM',     'approved',  0),
  (13, 't13', 'Daily Maintenance Issue Closure',             6,  'Infrastructure Maintenance', 'Daily Tasks',     'Daily Tasks',     '5:00 PM',     'overdue',   3),
  (14, 't14', 'Master Data Integrity Review',                7,  'Data Management',            'Monthly Tasks',   'Monthly Tasks',   '10th',        'approved',  0),
  (15, 't15', 'Inventory Exception Closure',                 7,  'Inventory',                  'Monthly Tasks',   'Monthly Tasks',   '15th',        'submitted', 1),
  (16, 't16', 'Monthly Budget Variance Closure',             8,  'Finance',                    'Monthly Tasks',   'Finance Tasks',   '5th',         'overdue',   3),
  (17, 't17', 'Daily Payment Approval Queue',                8,  'Finance',                    'Daily Tasks',     'Finance Tasks',   '4:30 PM',     'submitted', 0),
  (18, 't18', 'Placement Training Outcome Review',           9,  'Placement Training',         'Weekly Tasks',    'Weekly Tasks',    'Friday',      'submitted', 1),
  (19, 't19', 'Placement Drive Data Closure',                9,  'Placement',                  'Daily Tasks',     'Daily Tasks',     '5:30 PM',     'overdue',   2),
  (20, 't20', 'Sponsored Research Progress Review',          10, 'R&D',                        'Monthly Tasks',   'Monthly Tasks',   '7th',         'submitted', 1),
  (21, 't21', 'Faculty Publication Validation',              10, 'Faculty Achievement',        'Monthly Tasks',   'Monthly Tasks',   '12th',        'approved',  0),
  (22, 't22', 'Industry Interaction Pipeline Review',        11, 'Industry Relations',         'Weekly Tasks',    'Weekly Tasks',    'Thursday',    'submitted', 0),
  (23, 't23', 'Alumni Mentor Engagement Review',             12, 'Alumni Engagement',          'Monthly Tasks',   'Monthly Tasks',   '14th',        'approved',  0)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  role_id = VALUES(role_id),
  category = VALUES(category),
  frequency = VALUES(frequency),
  task_type = VALUES(task_type),
  deadline_label = VALUES(deadline_label),
  default_pattern = VALUES(default_pattern),
  default_misses = VALUES(default_misses);

-- 4. Populate Actual Stored Task Instances Across -7 to +8 Days Around 2026-10-11
INSERT INTO task_instances (
  template_id,
  owner_id,
  scheduled_date,
  status,
  past_misses,
  submitted_at,
  verified_at,
  verified_by,
  closed_at
)
SELECT
  tt.id AS template_id,
  u.id AS owner_id,
  d.sched_date AS scheduled_date,
  CASE
    WHEN d.sched_date > '2026-10-11' THEN
      CASE
        WHEN MOD(DAYOFMONTH(d.sched_date) + u.owner_idx + CHAR_LENGTH(tt.code), 7) = 0 THEN 'approved_early'
        WHEN MOD(DAYOFMONTH(d.sched_date) + u.owner_idx + CHAR_LENGTH(tt.code), 7) = 1 THEN 'submitted_early'
        ELSE 'scheduled'
      END
    ELSE
      CASE
        WHEN tt.default_pattern = 'approved' THEN 'approved'
        WHEN tt.default_pattern = 'submitted' THEN
          IF(MOD(DAYOFMONTH(d.sched_date) + u.owner_idx + CHAR_LENGTH(tt.code), 2) = 0, 'submitted', 'approved')
        WHEN tt.default_pattern = 'overdue' THEN
          IF(MOD(DAYOFMONTH(d.sched_date) + u.owner_idx + CHAR_LENGTH(tt.code), 7) = 0, 'approved', 'overdue')
        ELSE 'submitted'
      END
  END AS status,
  tt.default_misses AS past_misses,
  CASE
    WHEN d.sched_date <= '2026-10-11' AND tt.default_pattern IN ('approved', 'submitted')
      THEN CONCAT(d.sched_date, ' 16:15:00')
    ELSE NULL
  END AS submitted_at,
  CASE
    WHEN d.sched_date <= '2026-10-11' AND tt.default_pattern = 'approved'
      THEN CONCAT(d.sched_date, ' 17:10:00')
    ELSE NULL
  END AS verified_at,
  CASE
    WHEN d.sched_date <= '2026-10-11' AND tt.default_pattern = 'approved'
      THEN r.verifier_title
    ELSE NULL
  END AS verified_by,
  CASE
    WHEN d.sched_date <= '2026-10-11' AND tt.default_pattern = 'approved'
      THEN CONCAT(d.sched_date, ' 17:10:00')
    ELSE NULL
  END AS closed_at
FROM task_templates tt
JOIN roles r ON r.id = tt.role_id
JOIN (
  SELECT
    u1.id,
    u1.role_id,
    (
      SELECT COUNT(*)
      FROM users u2
      WHERE u2.role_id = u1.role_id AND u2.id < u1.id
    ) AS owner_idx
  FROM users u1
) u ON u.role_id = tt.role_id
CROSS JOIN (
  SELECT DATE_ADD('2026-10-11', INTERVAL seq.offset_day DAY) AS sched_date
  FROM (
    SELECT -7 AS offset_day UNION ALL SELECT -6 UNION ALL SELECT -5 UNION ALL SELECT -4
    UNION ALL SELECT -3 UNION ALL SELECT -2 UNION ALL SELECT -1 UNION ALL SELECT 0
    UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4
    UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8
  ) seq
) d
ON DUPLICATE KEY UPDATE
  past_misses = VALUES(past_misses);

-- 5. Seed Initial Task Activity Logs for Stored Instances (Idempotent)
DELETE FROM task_activity_logs;

-- 5a. Recurring responsibility baseline entry (oldest)
INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
SELECT
  ti.id,
  'Recurring responsibility',
  CONCAT(tt.frequency, ' task mapped to ', r.name),
  'gray',
  DATE_SUB(CONCAT(ti.scheduled_date, ' 08:00:00'), INTERVAL 3 DAY)
FROM task_instances ti
JOIN task_templates tt ON tt.id = ti.template_id
JOIN roles r ON r.id = tt.role_id;

-- 5b. Historical delay entry (for tasks with past_misses > 0)
INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
SELECT
  ti.id,
  'Historical delay',
  CONCAT(ti.past_misses, ' previous missed / delayed cycle(s)'),
  'red',
  DATE_SUB(CONCAT(ti.scheduled_date, ' 08:00:00'), INTERVAL 2 DAY)
FROM task_instances ti
WHERE ti.past_misses > 0;

-- 5c. Previous cycle verification entry
INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
SELECT
  ti.id,
  'Verification completed',
  CONCAT('Reviewed by ', r.verifier_title),
  'green',
  DATE_SUB(CONCAT(ti.scheduled_date, ' 08:00:00'), INTERVAL 1 DAY)
FROM task_instances ti
JOIN task_templates tt ON tt.id = ti.template_id
JOIN roles r ON r.id = tt.role_id;

-- 5d. Previous cycle submission entry (most recent default entry)
INSERT INTO task_activity_logs (task_instance_id, event_title, event_meta, dot_color, created_at)
SELECT
  ti.id,
  'Previous cycle completed',
  'Submission received from task owner',
  'green',
  CONCAT(ti.scheduled_date, ' 08:00:00')
FROM task_instances ti;

