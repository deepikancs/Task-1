import React from 'react';

export default function SummaryStrip({
  mode,
  isPastDate,
  isFutureDate,
  tasks,
  aggregateSummary
}) {
  if (mode === 'type') {
    return (
      <div className="summary-strip">
        <span className="summary-pill">
          <span
            className="summary-dot"
            style={{ background: 'var(--primary)' }}
          />
          {aggregateSummary?.uniqueTasks ?? 0} Unique Tasks
        </span>

        <span className="summary-pill">
          <span
            className="summary-dot"
            style={{ background: 'var(--purple)' }}
          />
          {aggregateSummary?.rolesCount ?? 0} Roles
        </span>

        <span className="summary-pill">
          <span
            className="summary-dot"
            style={{ background: 'var(--green)' }}
          />
          {aggregateSummary?.peopleCount ?? 0} People
        </span>
      </div>
    );
  }

  if (isFutureDate) {
    const scheduled = tasks.filter((r) => r.state === 'scheduled').length;
    const submittedEarly = tasks.filter(
      (r) => r.state === 'submitted_early'
    ).length;
    const completedEarly = tasks.filter(
      (r) => r.state === 'approved_early'
    ).length;

    return (
      <div className="summary-strip">
        <span className="summary-pill">
          <span className="summary-dot" style={{ background: '#94a3b8' }} />
          {scheduled} Yet to Start
        </span>

        <span className="summary-pill">
          <span
            className="summary-dot"
            style={{ background: 'var(--purple)' }}
          />
          {submittedEarly} Completed Early · Verification Pending
        </span>

        <span className="summary-pill">
          <span
            className="summary-dot"
            style={{ background: 'var(--green)' }}
          />
          {completedEarly} Completed Early &amp; Verified
        </span>

        <span className="summary-pill">
          <span className="summary-dot" style={{ background: 'var(--red)' }} />0
          Overdue
        </span>
      </div>
    );
  }

  if (isPastDate) {
    const rolled = tasks.filter((r) => r.state !== 'approved').length;
    const completed = tasks.filter((r) => r.state === 'approved').length;

    return (
      <div className="summary-strip">
        <span className="summary-pill">
          <span className="summary-dot" style={{ background: 'var(--red)' }} />
          {rolled} Rolled to Today as Overdue
        </span>

        <span className="summary-pill">
          <span
            className="summary-dot"
            style={{ background: 'var(--green)' }}
          />
          {completed} Fully Completed
        </span>
      </div>
    );
  }

  // Today
  const overdue = tasks.filter((r) => r.state === 'overdue').length;
  const submitted = tasks.filter((r) => r.state === 'submitted').length;
  const completed = tasks.filter((r) => r.state === 'approved').length;

  return (
    <div className="summary-strip">
      <span className="summary-pill">
        <span className="summary-dot" style={{ background: 'var(--red)' }} />
        {overdue} Overdue
      </span>

      <span className="summary-pill">
        <span className="summary-dot" style={{ background: 'var(--purple)' }} />
        {submitted} Verification Pending
      </span>

      <span className="summary-pill">
        <span className="summary-dot" style={{ background: 'var(--green)' }} />
        {completed} Fully Closed
      </span>
    </div>
  );
}

