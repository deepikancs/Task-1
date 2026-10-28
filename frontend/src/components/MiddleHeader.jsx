import React from 'react';
import { baseName } from '../seed/devSeedData';

export default function MiddleHeader({
  mode,
  selectedFilter,
  selectedPerson,
  isMonthlyMode,
  onOpenAnalytics,
  onOpenAssignTask
}) {
  let title;
  let subtitle;

  if (mode === 'role') {
    if (selectedFilter === 'All') {
      title = 'All Organisation Tasks';
      subtitle =
        'Select a role to inspect responsibility by person and operating date.';
    } else if (selectedPerson === 'All') {
      title = `${selectedFilter} · All Owners`;
      subtitle = `View all tasks under the ${selectedFilter} role for the selected date.`;
    } else {
      title = baseName(selectedPerson);
      subtitle = `${selectedFilter} · Complete task position for the selected date.`;
    }
  } else {
    title = selectedFilter === 'All' ? 'All Task Types' : selectedFilter;
    subtitle = isMonthlyMode
      ? 'Compare recurring tasks across all responsible roles for the selected month.'
      : 'Compare unique tasks across roles and their completion position.';
  }

  return (
    <div className="middle-controls">
      <div className="control-heading">
        <div>
          <div className="eyebrow">Executive task visibility</div>
          <div className="page-title">{title}</div>
          <div className="page-subtitle">{subtitle}</div>
        </div>

        <div className="heading-actions">
          <button
            type="button"
            className="secondary-button"
            onClick={onOpenAnalytics}
          >
            Task Analytics
          </button>

          <button
            type="button"
            className="primary-button"
            onClick={onOpenAssignTask}
          >
            + Assign Task
          </button>
        </div>
      </div>
    </div>
  );
}
