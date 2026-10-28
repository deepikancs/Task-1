import React from 'react';

export default function TaskRow({ task, section, isSelected, onSelect }) {
  let rowClass;
  let badgeElement;
  let timeClass = '';
  let subText = '';

  if (section === 'overdue') {
    rowClass = task.misses >= 2 ? 'repeat-miss' : 'overdue';
    badgeElement = <span className="badge badge-red">Overdue</span>;
    timeClass = 'red';
    subText = 'Deadline passed';
  } else if (section === 'verification') {
    rowClass = 'verification';
    badgeElement = (
      <span className="badge badge-purple">Verification Pending</span>
    );
    subText = 'Waiting with verifier';
  } else if (section === 'completed') {
    rowClass = 'completed';
    badgeElement = <span className="badge badge-green">Fully Closed</span>;
    timeClass = 'green';
    subText = 'Verified and closed';
  } else if (section === 'scheduled') {
    rowClass = 'scheduled';
    badgeElement = <span className="badge badge-gray">Yet to Start</span>;
    subText = 'Future scheduled task';
  } else if (section === 'earlyverification') {
    rowClass = 'verification';
    badgeElement = <span className="badge badge-purple">Completed Early</span>;
    subText = 'Early submission · verification pending';
  } else if (section === 'earlycompleted') {
    rowClass = 'completed';
    badgeElement = <span className="badge badge-green">Completed Early</span>;
    timeClass = 'green';
    subText = 'Completed before due date';
  } else {
    rowClass = 'live';
    badgeElement = <span className="badge badge-orange">Live</span>;
    timeClass = 'orange';
    subText = 'Yet to complete';
  }

  return (
    <div
      role="button"
      tabIndex={0}
      className={`task-row ${rowClass} ${isSelected ? 'selected' : ''}`}
      onClick={() => onSelect(task.uid)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect(task.uid);
        }
      }}
    >
      <div>
        <div className="task-title">{task.name}</div>

        <div className="task-meta">
          <span className="badge badge-blue">{task.category}</span>
          {badgeElement}
          {task.misses >= 2 && (
            <span className="badge badge-red">Repeat Miss</span>
          )}
        </div>
      </div>

      <div>
        <div className="ribbon-flow">
          {task.flow.map((step, index) => (
            <React.Fragment key={step.no}>
              <div className={`flow-step ${step.status}`}>
                <div className="flow-dot">{step.no}</div>
                <div className="flow-text">
                  <div className="flow-title">{step.title}</div>
                  <div className="flow-person">{step.person}</div>
                </div>
              </div>
              {index < task.flow.length - 1 && (
                <span className="flow-arrow">›</span>
              )}
            </React.Fragment>
          ))}
        </div>
      </div>

      <div>
        <div className={`task-time ${timeClass}`}>{task.time}</div>
        <div className="task-sub">{subText}</div>
      </div>

      <div className="task-score">
        <strong>{task.misses}</strong>
        <span>Past misses</span>
      </div>
    </div>
  );
}

