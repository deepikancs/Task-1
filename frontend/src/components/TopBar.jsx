import React from 'react';
import { IS_USING_DEV_SEED } from '../api/taskApi';

export default function TopBar({ todayISO, onChangeTodayISO, backendDemoMode }) {
  return (
    <header className="topbar">
      <div className="brand">
        <div className="brand-mark">OC</div>
        <div className="brand-title">
          Organisation <span>Control Centre</span>
        </div>
      </div>

      <div className="top-actions">
        {(IS_USING_DEV_SEED || backendDemoMode) && (
          <span
            className="demo-seed-badge"
            title="Running in Demo Mode: Submit, Approve, and Reject changes are kept in memory for the current run only and reset when the backend stops."
          >
            Demo mode — data is temporary
          </span>
        )}

        <label
          className="demo-date-control"
          title="Configurable reference Today date for testing Past, Today, and Future states"
        >
          <span>Demo Today:</span>
          <input
            type="date"
            className="demo-date-input"
            value={todayISO}
            onChange={(e) => {
              if (e.target.value) {
                onChangeTodayISO(e.target.value);
              }
            }}
          />
        </label>

        <span className="top-pill">Executive Task Monitor</span>

        <span
          className="avatar"
          title="Demo Executive Officer view (not real authentication)"
        >
          EO
        </span>
      </div>
    </header>
  );
}
