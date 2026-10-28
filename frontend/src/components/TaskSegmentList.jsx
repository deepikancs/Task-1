import React from 'react';
import TaskRow from './TaskRow';

export default function TaskSegmentList({
  tasks,
  isPastDate,
  isFutureDate,
  selectedTaskId,
  onSelectTask
}) {
  let groups = [];

  if (isPastDate) {
    groups = [
      {
        key: 'rolled',
        title: 'Incomplete on this Past Date · Rolled to Today as Overdue',
        cls: 'red',
        filter: (row) => row.state !== 'approved',
        render: 'overdue'
      },
      {
        key: 'completed',
        title: 'Fully Completed & Verified on this Date',
        cls: 'green',
        filter: (row) => row.state === 'approved',
        render: 'completed'
      }
    ];
  } else if (isFutureDate) {
    groups = [
      {
        key: 'scheduled',
        title: 'Yet to Start / Scheduled',
        cls: 'gray',
        filter: (row) => row.state === 'scheduled',
        render: 'scheduled'
      },
      {
        key: 'earlyverification',
        title: 'Completed Early · Waiting for Verification / Approval',
        cls: 'purple',
        filter: (row) => row.state === 'submitted_early',
        render: 'earlyverification'
      },
      {
        key: 'earlycompleted',
        title: 'Completed Early & Fully Verified',
        cls: 'green',
        filter: (row) => row.state === 'approved_early',
        render: 'earlycompleted'
      }
    ];
  } else {
    groups = [
      {
        key: 'overdue',
        title: 'Overdue · Still Not Completed',
        cls: 'red',
        filter: (row) => row.state === 'overdue',
        render: 'overdue'
      },
      {
        key: 'verification',
        title: 'Submitted · Waiting for Verification / Approval',
        cls: 'purple',
        filter: (row) => row.state === 'submitted',
        render: 'verification'
      },
      {
        key: 'completed',
        title: 'Completed & Fully Verified',
        cls: 'green',
        filter: (row) => row.state === 'approved',
        render: 'completed'
      }
    ];
  }

  return (
    <div className="content-scroll">
      {groups.map((group) => {
        const groupRows = tasks.filter(group.filter);
        return (
          <section className="segment" key={group.key}>
            <div className="segment-head">
              <div className={`segment-title ${group.cls}`}>{group.title}</div>
              <span className="segment-count">{groupRows.length}</span>
            </div>

            {groupRows.length > 0 ? (
              groupRows.map((row) => (
                <TaskRow
                  key={row.uid}
                  task={row}
                  section={group.render}
                  isSelected={selectedTaskId === row.uid}
                  onSelect={onSelectTask}
                />
              ))
            ) : (
              <div className="empty-state">No tasks in this segment.</div>
            )}
          </section>
        );
      })}
    </div>
  );
}

