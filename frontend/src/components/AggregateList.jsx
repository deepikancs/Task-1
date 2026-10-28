import React from 'react';

export default function AggregateList({ items, isMonthlyMode, onSelectRole }) {
  const heading = isMonthlyMode
    ? 'Tasks for Selected Month'
    : 'Unique Task Completion by Role';

  return (
    <div className="content-scroll">
      <section className="segment">
        <div className="segment-head">
          <div className="segment-title" style={{ color: 'var(--primary)' }}>
            {heading}
          </div>
          <span className="segment-count">{items.length}</span>
        </div>

        {items.length > 0 ? (
          items.map((task) => {
            let fillClass = '';
            if (task.percentage < 60) {
              fillClass = 'red';
            } else if (task.percentage < 85) {
              fillClass = 'orange';
            }

            return (
              <div
                className="aggregate-row"
                key={task.id}
                role="button"
                tabIndex={0}
                title={`Click to inspect ${task.role} responsibilities in Role View`}
                onClick={() => onSelectRole && onSelectRole(task.role)}
                onKeyDown={(e) => {
                  if (
                    (e.key === 'Enter' || e.key === ' ') &&
                    onSelectRole
                  ) {
                    e.preventDefault();
                    onSelectRole(task.role);
                  }
                }}
              >
                <div>
                  <div className="task-title">{task.name}</div>
                  <div className="task-meta">
                    <span className="badge badge-blue">{task.role}</span>
                    <span className="badge badge-purple">{task.category}</span>
                  </div>
                </div>

                <div className="people-count">
                  {task.people} People
                  <span>{task.completed} completed</span>
                </div>

                <div>
                  <div className="percent">{task.percentage}%</div>
                  <div className="progress-track">
                    <div
                      className={`progress-fill ${fillClass}`}
                      style={{ width: `${task.percentage}%` }}
                    />
                  </div>
                </div>

                <div className="task-score">
                  <strong>{task.misses}</strong>
                  <span>Past misses</span>
                </div>
              </div>
            );
          })
        ) : (
          <div className="empty-state">No tasks in this segment.</div>
        )}
      </section>
    </div>
  );
}
