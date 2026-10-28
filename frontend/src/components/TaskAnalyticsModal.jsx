import React from 'react';

export default function TaskAnalyticsModal({
  isOpen,
  onClose,
  tasks,
  selectedFilter,
  selectedPerson,
  selectedDate
}) {
  if (!isOpen) return null;

  const totalTasks = tasks.length;
  const overdueCount = tasks.filter(
    (t) => t.state === 'overdue' || t.effective === 'rolled'
  ).length;
  const awaitingVerificationCount = tasks.filter(
    (t) => t.state === 'submitted' || t.state === 'submitted_early'
  ).length;
  const completedCount = tasks.filter(
    (t) => t.state === 'approved' || t.state === 'approved_early'
  ).length;
  const scheduledCount = tasks.filter((t) => t.state === 'scheduled').length;
  const repeatMissCount = tasks.filter((t) => t.misses >= 2).length;

  const completionRate =
    totalTasks > 0 ? Math.round((completedCount / totalTasks) * 100) : 0;

  return (
    <div
      className="modal-backdrop"
      role="dialog"
      aria-modal="true"
      aria-labelledby="analyticsModalTitle"
      onClick={onClose}
    >
      <div
        className="modal-card"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="modal-head">
          <div>
            <div className="eyebrow">Demo Analytics Summary</div>
            <div className="detail-title" id="analyticsModalTitle">
              Task Analytics Overview
            </div>
            <div className="detail-subtitle">
              Scope: {selectedFilter === 'All' ? 'All Roles' : selectedFilter}
              {selectedPerson !== 'All' ? ` · ${selectedPerson}` : ''} ·{' '}
              {selectedDate.toLocaleDateString('en', {
                day: 'numeric',
                month: 'short',
                year: 'numeric'
              })}
            </div>
          </div>

          <button
            type="button"
            className="secondary-button"
            onClick={onClose}
          >
            Close
          </button>
        </div>

        <div className="modal-body">
          <div className="demo-inline-note">
            <strong>Demo only — confirm this workflow with the project owner.</strong>{' '}
            These counts are calculated from the currently loaded demo tasks.
            Final KPI formulas and reporting periods require confirmation.
          </div>

          <div className="analytics-kpi-grid">
            <div className="analytics-kpi-box">
              <strong>{totalTasks}</strong>
              <span>Total Tasks in View</span>
            </div>

            <div className="analytics-kpi-box red">
              <strong>{overdueCount}</strong>
              <span>Overdue / Rolled</span>
            </div>

            <div className="analytics-kpi-box purple">
              <strong>{awaitingVerificationCount}</strong>
              <span>Awaiting Verification</span>
            </div>

            <div className="analytics-kpi-box green">
              <strong>{completedCount}</strong>
              <span>Completed &amp; Closed</span>
            </div>

            <div className="analytics-kpi-box">
              <strong>{scheduledCount}</strong>
              <span>Scheduled (Future)</span>
            </div>

            <div className="analytics-kpi-box red">
              <strong>{repeatMissCount}</strong>
              <span>Repeat Misses (&ge;2)</span>
            </div>
          </div>

          <div className="detail-section" style={{ marginBottom: 0 }}>
            <div className="section-title">
              Demo Completion Rate ({completionRate}%)
            </div>
            <div className="progress-track" style={{ height: 10 }}>
              <div
                className={`progress-fill ${
                  completionRate < 60
                    ? 'red'
                    : completionRate < 85
                    ? 'orange'
                    : ''
                }`}
                style={{ width: `${completionRate}%` }}
              />
            </div>
            <div className="timeline-meta" style={{ marginTop: 6 }}>
              {completedCount} of {totalTasks} task instances in the current
              selection have completed both owner submission and verifier
              approval.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

