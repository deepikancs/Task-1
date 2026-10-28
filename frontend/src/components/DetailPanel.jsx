import React, { useState, useEffect } from 'react';
import { baseName } from '../seed/devSeedData';

function TimelineItem({ title, meta, colour }) {
  return (
    <div className="timeline-item">
      <span className={`timeline-dot ${colour || ''}`} />
      <div className="timeline-title">{title}</div>
      <div className="timeline-meta">{meta}</div>
    </div>
  );
}

export default function DetailPanel({
  mode,
  selectedFilter,
  selectedTask,
  selectedDate,
  isPastDate,
  onSubmitTask,
  onVerifyTask,
  onEscalateTask,
  errorNotice,
  onClearError
}) {
  const [actionBusy, setActionBusy] = useState(false);
  // Active demo tool panel: null | 'message' | 'submission' | 'evidence' | 'escalate'
  const [activeTool, setActiveTool] = useState(null);
  const [copiedFeedback, setCopiedFeedback] = useState(false);
  const [escalationReason, setEscalationReason] = useState('');
  const [statusFeedback, setStatusFeedback] = useState('');

  // Reset active sub-panel when switching tasks
  useEffect(() => {
    setActiveTool(null);
    setCopiedFeedback(false);
    setEscalationReason('');
    setStatusFeedback('');
  }, [selectedTask?.uid, selectedTask?.dateISO]);

  if (mode === 'type') {
    return (
      <aside className="right-panel">
        <div className="detail-head">
          <div className="detail-type">Task Type View</div>
          <div className="detail-title">
            {selectedFilter === 'All'
              ? 'Organisation Task Comparison'
              : selectedFilter}
          </div>
          <div className="detail-subtitle">
            Select an individual responsibility in Role View to inspect its
            complete process history. Click any row in Task Type View to jump
            to its Role View.
          </div>
        </div>
      </aside>
    );
  }

  if (!selectedTask) {
    return (
      <aside className="right-panel">
        <div className="detail-head">
          <div className="detail-type">Task Inspector</div>
          <div className="detail-title">No Task Selected</div>
          <div className="detail-subtitle">
            Select a task row to inspect its workflow and history.
          </div>
        </div>
      </aside>
    );
  }

  const verifier = selectedTask.verifier || 'Reporting Authority';
  const ownerShort = baseName(selectedTask.owner);

  const isWaitingVerifier =
    selectedTask.state === 'submitted' ||
    selectedTask.state === 'submitted_early';
  const isClosed =
    selectedTask.state === 'approved' ||
    selectedTask.state === 'approved_early';

  const currentHolder = isWaitingVerifier
    ? `${verifier} (Verifier)`
    : isClosed
    ? `Closed (${ownerShort})`
    : `${selectedTask.owner} (Owner)`;

  const sampleMessage = isWaitingVerifier
    ? `Hello ${verifier}, the task "${selectedTask.name}" submitted by ${ownerShort} (due ${selectedTask.time}) is awaiting your verification in the Organisation Task Control Centre.`
    : `Hello ${ownerShort}, please review the task "${selectedTask.name}" (${selectedTask.frequency}, deadline: ${selectedTask.time}) and submit your update in the Organisation Task Control Centre.`;

  const hasSubmission =
    selectedTask.state === 'submitted' ||
    selectedTask.state === 'approved' ||
    selectedTask.state === 'submitted_early' ||
    selectedTask.state === 'approved_early' ||
    Boolean(selectedTask.submissionNotes);

  let bannerElement = null;

  if (isPastDate && selectedTask.state !== 'approved') {
    bannerElement = (
      <div className="danger-banner">
        <strong>
          This task was not fully closed on the selected past date.
        </strong>{' '}
        It has automatically rolled forward to today&apos;s Overdue queue.
      </div>
    );
  } else if (selectedTask.state === 'scheduled') {
    bannerElement = (
      <div className="future-banner">
        <strong>Future scheduled task.</strong> The due period has not started.
        This task cannot be considered overdue.
      </div>
    );
  } else if (selectedTask.state === 'submitted_early') {
    bannerElement = (
      <div
        className="warning-banner"
        style={{
          background: 'var(--purple-bg)',
          borderColor: 'var(--purple-border)',
          color: 'var(--purple)'
        }}
      >
        <strong>
          This future task was completed and submitted in advance.
        </strong>{' '}
        It is currently waiting for verification by {verifier}.
      </div>
    );
  } else if (selectedTask.state === 'approved_early') {
    bannerElement = (
      <div
        className="future-banner"
        style={{
          background: 'var(--green-bg)',
          borderColor: 'var(--green-border)',
          color: 'var(--green)'
        }}
      >
        <strong>This future task has already been fully completed.</strong>{' '}
        Work, submission, verification and closure were completed before the
        scheduled due date.
      </div>
    );
  } else if (selectedTask.state === 'overdue') {
    bannerElement = (
      <div className="danger-banner">
        <strong>This task is overdue.</strong> The responsibility is currently
        pending with {ownerShort}.
      </div>
    );
  } else if (selectedTask.state === 'submitted') {
    bannerElement = (
      <div
        className="warning-banner"
        style={{
          background: 'var(--purple-bg)',
          borderColor: 'var(--purple-border)',
          color: 'var(--purple)'
        }}
      >
        <strong>Work has been completed and submitted.</strong> The task remains
        open because verification is pending with {verifier}.
      </div>
    );
  }

  const canSubmit =
    selectedTask.state === 'overdue' || selectedTask.state === 'scheduled';
  const canVerifyOrReject =
    selectedTask.state === 'submitted' ||
    selectedTask.state === 'submitted_early';

  const handleDemoSubmit = async () => {
    setActionBusy(true);
    setStatusFeedback('');
    try {
      await onSubmitTask(selectedTask);
      setStatusFeedback('Demo submission recorded for this run.');
    } finally {
      setActionBusy(false);
    }
  };

  const handleDemoVerify = async (action) => {
    setActionBusy(true);
    setStatusFeedback('');
    try {
      await onVerifyTask(selectedTask, action);
      setStatusFeedback(
        action === 'approve'
          ? 'Task approved and marked closed for this run.'
          : 'Task rejected and returned to owner for this run.'
      );
    } finally {
      setActionBusy(false);
    }
  };

  const handleCopyMessage = async () => {
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(sampleMessage);
      }
      setCopiedFeedback(true);
      setTimeout(() => setCopiedFeedback(false), 2500);
    } catch {
      setCopiedFeedback(true);
    }
  };

  const handleConfirmEscalate = async () => {
    setActionBusy(true);
    try {
      await onEscalateTask(selectedTask, verifier, escalationReason);
      setActiveTool(null);
      setEscalationReason('');
      setStatusFeedback(
        `Demo escalation logged for ${verifier} in Historical Activity.`
      );
    } finally {
      setActionBusy(false);
    }
  };

  const toggleTool = (toolName) => {
    setActiveTool((prev) => (prev === toolName ? null : toolName));
    setCopiedFeedback(false);
  };

  return (
    <aside className="right-panel">
      <div className="detail-head">
        <div className="detail-type">{selectedTask.category}</div>
        <div className="detail-title">{selectedTask.name}</div>
        <div className="detail-subtitle">
          {ownerShort} · {selectedTask.role} · {selectedTask.frequency}
        </div>
      </div>

      <div className="detail-body">
        {errorNotice && (
          <div className="danger-banner">
            <strong>Action Error:</strong> {errorNotice}{' '}
            <button
              type="button"
              className="action-button secondary"
              style={{ minHeight: '22px', padding: '2px 6px', marginLeft: 4 }}
              onClick={onClearError}
            >
              Dismiss
            </button>
          </div>
        )}

        {statusFeedback && (
          <div
            className="future-banner"
            style={{
              background: 'var(--green-bg)',
              borderColor: 'var(--green-border)',
              color: 'var(--green)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              gap: 6
            }}
          >
            <span>{statusFeedback}</span>
            <button
              type="button"
              className="action-button secondary"
              style={{ minHeight: 20, padding: '1px 6px' }}
              onClick={() => setStatusFeedback('')}
            >
              ×
            </button>
          </div>
        )}

        {bannerElement}

        {/* Demo Workflow Actions */}
        <div className="demo-workflow-box">
          <div className="demo-workflow-header">
            <span className="demo-workflow-title">Demo Workflow Actions</span>
            <span
              className="demo-workflow-tag"
              title="Simulated permissions for demonstration; not real authentication"
            >
              Demo Only
            </span>
          </div>

          <div className="demo-workflow-grid">
            <button
              type="button"
              className="action-button"
              disabled={!canSubmit || actionBusy}
              title="Simulate Owner submitting this task"
              onClick={handleDemoSubmit}
            >
              Submit Task
            </button>

            <button
              type="button"
              className="action-button success"
              disabled={!canVerifyOrReject || actionBusy}
              title="Simulate Verifier approving and closing this task"
              onClick={() => handleDemoVerify('approve')}
            >
              Approve
            </button>

            <button
              type="button"
              className="action-button danger"
              disabled={!canVerifyOrReject || actionBusy}
              title="Simulate Verifier rejecting and returning task to Owner"
              onClick={() => handleDemoVerify('reject')}
            >
              Reject
            </button>
          </div>
        </div>

        {/* Preserved Prototype Action Grid — all 4 buttons open interactive demo panels */}
        <div className="action-grid">
          <button
            type="button"
            className={`action-button ${
              activeTool === 'message' ? 'active-tool' : ''
            }`}
            onClick={() => toggleTool('message')}
          >
            Message Current Holder
          </button>

          <button
            type="button"
            className={`action-button secondary ${
              activeTool === 'submission' ? 'active-tool' : ''
            }`}
            onClick={() => toggleTool('submission')}
          >
            View Submission
          </button>

          <button
            type="button"
            className={`action-button secondary ${
              activeTool === 'evidence' ? 'active-tool' : ''
            }`}
            onClick={() => toggleTool('evidence')}
          >
            View Evidence
          </button>

          <button
            type="button"
            className={`action-button danger ${
              activeTool === 'escalate' ? 'active-tool' : ''
            }`}
            onClick={() => toggleTool('escalate')}
          >
            Escalate
          </button>
        </div>

        {/* Interactive Demo Sub-Panel for the 4 Action Buttons */}
        {activeTool === 'message' && (
          <div className="demo-tool-panel">
            <div className="demo-tool-head">
              <strong>Message Current Holder (Demo)</strong>
              <button
                type="button"
                className="demo-tool-close"
                onClick={() => setActiveTool(null)}
              >
                Close
              </button>
            </div>
            <div className="demo-tool-meta">
              Current Holder: <strong>{currentHolder}</strong>
            </div>
            <textarea
              className="demo-tool-textarea"
              readOnly
              rows={3}
              value={sampleMessage}
            />
            <div className="demo-tool-actions">
              <button
                type="button"
                className="action-button"
                onClick={handleCopyMessage}
              >
                {copiedFeedback ? '✓ Copied to Clipboard' : 'Copy Message'}
              </button>
              <span className="demo-tool-note">
                Demo only — no real message is sent. Confirm messaging channel
                with the project owner.
              </span>
            </div>
          </div>
        )}

        {activeTool === 'submission' && (
          <div className="demo-tool-panel">
            <div className="demo-tool-head">
              <strong>Submission Preview (Demo)</strong>
              <button
                type="button"
                className="demo-tool-close"
                onClick={() => setActiveTool(null)}
              >
                Close
              </button>
            </div>
            {hasSubmission ? (
              <>
                <div className="demo-tool-box">
                  {selectedTask.submissionNotes ||
                    `Demo submission filed by ${ownerShort} for "${selectedTask.name}". Waiting for ${verifier} review.`}
                </div>
                <div className="demo-tool-note">
                  Demo only — confirm submission fields with the project owner.
                </div>
              </>
            ) : (
              <div className="empty-state" style={{ marginTop: 4 }}>
                No demo submission available yet for this task instance. Click{' '}
                <strong>Submit Task</strong> above to simulate filing a
                submission.
              </div>
            )}
          </div>
        )}

        {activeTool === 'evidence' && (
          <div className="demo-tool-panel">
            <div className="demo-tool-head">
              <strong>Evidence Preview (Demo)</strong>
              <button
                type="button"
                className="demo-tool-close"
                onClick={() => setActiveTool(null)}
              >
                Close
              </button>
            </div>
            {hasSubmission ? (
              <div className="demo-tool-box">
                <div>
                  <strong>Sample Reference:</strong>{' '}
                  {selectedTask.id.toUpperCase()}-
                  {selectedTask.dateISO || 'DEMO'}-checklist.pdf (Simulated
                  preview)
                </div>
                <div style={{ marginTop: 4 }}>
                  Logged by: {ownerShort} · Category: {selectedTask.category}
                </div>
              </div>
            ) : (
              <div className="empty-state" style={{ marginTop: 4 }}>
                No demo evidence attached yet because this task has not been
                submitted.
              </div>
            )}
            <div className="demo-tool-note" style={{ marginTop: 6 }}>
              Demo only — evidence attachment and storage rules are awaiting
              confirmation with the project owner.
            </div>
          </div>
        )}

        {activeTool === 'escalate' && (
          <div className="demo-tool-panel danger">
            <div className="demo-tool-head">
              <strong>Confirm Demo Escalation</strong>
              <button
                type="button"
                className="demo-tool-close"
                onClick={() => setActiveTool(null)}
              >
                Cancel
              </button>
            </div>
            <div className="demo-tool-meta">
              Task: <strong>{selectedTask.name}</strong>
              <br />
              Proposed Recipient: <strong>{verifier}</strong>
            </div>
            <input
              type="text"
              className="demo-tool-input"
              placeholder="Optional escalation note (demo)..."
              value={escalationReason}
              onChange={(e) => setEscalationReason(e.target.value)}
            />
            <div className="demo-tool-actions">
              <button
                type="button"
                className="action-button danger"
                disabled={actionBusy}
                onClick={handleConfirmEscalate}
              >
                Record Demo Escalation
              </button>
              <span className="demo-tool-note">
                Demo only — records an entry in Historical Activity below; no
                emails or notifications are sent.
              </span>
            </div>
          </div>
        )}

        <div className="detail-section">
          <div className="section-title">Current Workflow</div>

          {selectedTask.flow.map((step, index) => {
            let colour = 'gray';
            if (step.status === 'done') {
              colour = 'green';
            } else if (step.status === 'wait') {
              colour = 'purple';
            } else if (step.status === 'current') {
              colour = 'orange';
            }

            return (
              <TimelineItem
                key={step.no}
                title={`${index + 1}. ${step.title}`}
                meta={step.person}
                colour={colour}
              />
            );
          })}
        </div>

        <div className="detail-section">
          <div className="section-title">Task Information</div>

          <div className="field-row">
            <div className="field-label">Owner</div>
            <div className="field-value">{selectedTask.owner}</div>
          </div>

          <div className="field-row">
            <div className="field-label">Verifier</div>
            <div className="field-value">{verifier}</div>
          </div>

          <div className="field-row">
            <div className="field-label">Frequency</div>
            <div className="field-value">{selectedTask.frequency}</div>
          </div>

          <div className="field-row">
            <div className="field-label">Deadline</div>
            <div className="field-value">{selectedTask.time}</div>
          </div>

          <div className="field-row">
            <div className="field-label">Past Misses</div>
            <div className="field-value">{selectedTask.misses}</div>
          </div>

          <div className="field-row">
            <div className="field-label">Selected Date</div>
            <div className="field-value">
              {selectedDate.toLocaleDateString('en', {
                weekday: 'long',
                day: 'numeric',
                month: 'long',
                year: 'numeric'
              })}
            </div>
          </div>
        </div>

        <div className="detail-section">
          <div className="section-title">Historical Activity</div>

          {(selectedTask.history || []).map((item) => (
            <TimelineItem
              key={item.id}
              title={item.title}
              meta={item.meta}
              colour={item.colour}
            />
          ))}
        </div>
      </div>
    </aside>
  );
}
