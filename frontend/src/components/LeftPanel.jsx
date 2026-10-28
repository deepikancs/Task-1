import React, { useState } from 'react';

export default function LeftPanel({
  organisationSummary,
  mode,
  onSelectMode,
  selectedFilter,
  onSelectFilter,
  roles,
  taskTypes
}) {
  const [searchQuery, setSearchQuery] = useState('');

  const rawOptions = mode === 'role' ? roles : taskTypes;
  const filteredOptions = rawOptions.filter((item) =>
    item.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totalAllCount = rawOptions.length;

  return (
    <aside className="left-panel">
      <div className="left-head">
        <div className="eyebrow">Organisation operations</div>

        <div className="panel-title">Task Monitor</div>

        <div className="panel-subtitle">
          Monitor recurring and one-time responsibilities by role, owner, task
          type and operating period.
        </div>

        <div className="control-summary">
          <div className="summary-name">Current Organisation Position</div>

          <div className="summary-grid">
            <div className="summary-box">
              <strong>{organisationSummary?.activeTasks ?? 86}</strong>
              <span>Active Tasks</span>
            </div>

            <div className="summary-box">
              <strong>{organisationSummary?.overdue ?? 9}</strong>
              <span>Overdue</span>
            </div>

            <div className="summary-box">
              <strong>{organisationSummary?.repeatMiss ?? 6}</strong>
              <span>Repeat Miss</span>
            </div>
          </div>
        </div>
      </div>

      <div className="filter-mode">
        <div className="toggle" role="tablist" aria-label="Filter Mode">
          <button
            type="button"
            className={`toggle-button ${mode === 'role' ? 'active' : ''}`}
            onClick={() => onSelectMode('role')}
          >
            By Role
          </button>

          <button
            type="button"
            className={`toggle-button ${mode === 'type' ? 'active' : ''}`}
            onClick={() => onSelectMode('type')}
          >
            By Task Type
          </button>
        </div>
      </div>

      <div className="search-section">
        <div className="search-wrap">
          <span className="search-symbol" aria-hidden="true">
            ⌕
          </span>

          <input
            type="search"
            className="search-input"
            placeholder="Search..."
            aria-label="Search roles or task types"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      <div className="left-scroll">
        <div className="filter-title">
          {mode === 'role' ? 'Responsibility View' : 'Task Type View'}
        </div>

        <div className="filter-list">
          <button
            type="button"
            className={`filter-button ${
              selectedFilter === 'All' ? 'active' : ''
            }`}
            onClick={() => onSelectFilter('All')}
          >
            <span>All</span>
            <span className="filter-count">{totalAllCount}</span>
          </button>

          {filteredOptions.map((option) => (
            <button
              key={option.name}
              type="button"
              className={`filter-button ${
                selectedFilter === option.name ? 'active' : ''
              }`}
              onClick={() => onSelectFilter(option.name)}
            >
              <span>{option.name}</span>
              <span className="filter-count">{option.count}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="left-foot">
        Past incomplete tasks roll into today&apos;s Overdue queue. A task is
        considered fully completed only after all required verification or
        approval stages are closed.
      </div>
    </aside>
  );
}

