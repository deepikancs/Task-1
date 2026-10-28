import React from 'react';

export default function AssignTaskModal({ isOpen, onClose, roles }) {
  if (!isOpen) return null;

  return (
    <div
      className="modal-backdrop"
      role="dialog"
      aria-modal="true"
      aria-labelledby="assignTaskModalTitle"
      onClick={onClose}
    >
      <div
        className="modal-card"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="modal-head">
          <div>
            <div className="eyebrow">Task Assignment Preview</div>
            <div className="detail-title" id="assignTaskModalTitle">
              + Assign Task (Demo Preview)
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
            New task creation rules, custom frequencies, and approval chains
            have not been finalized yet. You can inspect the existing{' '}
            {roles?.length || 12} roles and 23 seeded task templates in the
            dashboard.
          </div>

          <div className="detail-section" style={{ marginBottom: 0 }}>
            <div className="section-title">Items to Confirm Before Enabling Task Creation</div>
            <div className="field-row">
              <div className="field-label">Who Can Assign</div>
              <div className="field-value">
                Confirm whether only Principal / Admin can create tasks or also
                Department Heads.
              </div>
            </div>
            <div className="field-row">
              <div className="field-label">Recurrence Rules</div>
              <div className="field-value">
                Confirm how Daily, Weekly, Monthly, and One-Time due dates roll
                over on holidays.
              </div>
            </div>
            <div className="field-row">
              <div className="field-label">Verifier Mapping</div>
              <div className="field-value">
                Confirm whether verifiers are fixed per role or chosen per task.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

