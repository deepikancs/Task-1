-- ============================================================================
-- ORGANISATION TASK CONTROL CENTRE
-- Migration 001: Core Relational Schema (MySQL 8.0+)
-- ============================================================================
-- Purpose:
--   1. roles              : Organisational roles and their designated verifiers
--   2. users              : Task owners assigned to roles
--   3. task_templates     : Master definitions of recurring and one-time tasks
--   4. task_instances     : Actual scheduled task occurrences per owner & date
--   5. task_activity_logs : Chronological audit history for each task instance
-- ============================================================================

CREATE DATABASE IF NOT EXISTS `Diviga`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

-- Create dedicated application user for the Go backend
-- IMPORTANT: Before running in SQL Studio, replace YOUR_CHOSEN_PASSWORD
-- with the exact same password you put in backend/.env!
CREATE USER IF NOT EXISTS 'otcc_app'@'localhost'
  IDENTIFIED BY 'YOUR_CHOSEN_PASSWORD';
ALTER USER 'otcc_app'@'localhost'
  IDENTIFIED BY 'YOUR_CHOSEN_PASSWORD';

CREATE USER IF NOT EXISTS 'otcc_app'@'127.0.0.1'
  IDENTIFIED BY 'YOUR_CHOSEN_PASSWORD';
ALTER USER 'otcc_app'@'127.0.0.1'
  IDENTIFIED BY 'YOUR_CHOSEN_PASSWORD';

GRANT ALL PRIVILEGES ON `Diviga`.* TO 'otcc_app'@'localhost';
GRANT ALL PRIVILEGES ON `Diviga`.* TO 'otcc_app'@'127.0.0.1';
FLUSH PRIVILEGES;

USE `Diviga`;

-- 1. Roles Table
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    verifier_title VARCHAR(150) NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uq_roles_name UNIQUE (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Users (Task Owners) Table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    full_name VARCHAR(150) NOT NULL,
    department VARCHAR(100) NULL,
    display_label VARCHAR(200) NOT NULL,
    role_id INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uq_users_display_label UNIQUE (display_label),
    CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    INDEX idx_users_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Task Templates Table
CREATE TABLE IF NOT EXISTS task_templates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(200) NOT NULL,
    role_id INT NOT NULL,
    category VARCHAR(100) NOT NULL,
    frequency VARCHAR(50) NOT NULL,
    task_type VARCHAR(50) NOT NULL,
    deadline_label VARCHAR(50) NOT NULL,
    default_pattern ENUM('submitted', 'approved', 'overdue') NOT NULL DEFAULT 'submitted',
    default_misses INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uq_task_templates_code UNIQUE (code),
    CONSTRAINT fk_task_templates_role FOREIGN KEY (role_id) REFERENCES roles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    INDEX idx_task_templates_role_id (role_id),
    INDEX idx_task_templates_task_type (task_type),
    INDEX idx_task_templates_frequency (frequency)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Task Instances Table (Actual scheduled occurrences stored in MySQL)
CREATE TABLE IF NOT EXISTS task_instances (
    id INT AUTO_INCREMENT PRIMARY KEY,
    template_id INT NOT NULL,
    owner_id INT NOT NULL,
    scheduled_date DATE NOT NULL,
    status ENUM(
        'scheduled',
        'live',
        'overdue',
        'submitted',
        'approved',
        'submitted_early',
        'approved_early'
    ) NOT NULL DEFAULT 'scheduled',
    past_misses INT NOT NULL DEFAULT 0,
    submitted_at DATETIME NULL,
    verified_at DATETIME NULL,
    verified_by VARCHAR(150) NULL,
    closed_at DATETIME NULL,
    submission_notes TEXT NULL,
    verification_remarks TEXT NULL,
    evidence_url VARCHAR(500) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uq_task_instance_owner_date UNIQUE (template_id, owner_id, scheduled_date),
    CONSTRAINT fk_task_instances_template FOREIGN KEY (template_id) REFERENCES task_templates(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_task_instances_owner FOREIGN KEY (owner_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    INDEX idx_task_instances_date (scheduled_date),
    INDEX idx_task_instances_status (status),
    INDEX idx_task_instances_owner_date (owner_id, scheduled_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Task Activity Logs Table (Audit & Workflow History)
CREATE TABLE IF NOT EXISTS task_activity_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_instance_id INT NOT NULL,
    event_title VARCHAR(150) NOT NULL,
    event_meta VARCHAR(255) NOT NULL,
    dot_color ENUM('green', 'purple', 'orange', 'red', 'gray') NOT NULL DEFAULT 'gray',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_activity_logs_instance FOREIGN KEY (task_instance_id) REFERENCES task_instances(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    INDEX idx_activity_logs_instance_created (task_instance_id, created_at DESC, id DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

