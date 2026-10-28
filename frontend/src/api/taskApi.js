/* =========================================================
   API CLIENT & DEVELOPMENT SEED STORE (STAGE 2)
   ---------------------------------------------------------
   All application components fetch data and trigger workflow
   actions exclusively through this API layer.
   - When `VITE_USE_DEV_SEED` is false (or in Stage 4), calls
     go to the Go HTTP API (`VITE_API_BASE_URL`).
   - During Stage 2, requests are served by the clearly
     identified in-memory Development Seed Store so that
     Submit, Approve, and Reject actions persist in state.
   ========================================================= */

import {
  peopleByRole,
  verifierByRole,
  taskTypes,
  taskTemplates,
  mappedType,
  baseName,
  parseISODate,
  sameDay,
  buildFlow,
  buildDefaultHistory
} from '../seed/devSeedData';

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export const IS_USING_DEV_SEED =
  (import.meta.env.VITE_USE_DEV_SEED ?? 'true') !== 'false';

// In-memory store for user mutations (Submit / Approve / Reject) in Dev Seed mode
// Keyed by `${dateISO}__${task.id}__${owner}`
const devInstanceOverrides = new Map();

function computeInitialState(task, index, selectedDateObj, todayObj) {
  const seed =
    (selectedDateObj.getDate() + index + task.id.length) % 7;

  const isFuture =
    selectedDateObj > todayObj && !sameDay(selectedDateObj, todayObj);

  if (isFuture) {
    if (seed === 0) return 'approved_early';
    if (seed === 1) return 'submitted_early';
    return 'scheduled';
  }

  if (task.pattern === 'approved') {
    return 'approved';
  }
  if (task.pattern === 'submitted') {
    return seed % 2 === 0 ? 'submitted' : 'approved';
  }
  if (task.pattern === 'overdue') {
    return seed === 0 ? 'approved' : 'overdue';
  }
  return 'submitted';
}

function getFilteredTemplates(mode, selectedFilter) {
  let tasks = [...taskTemplates];
  if (mode === 'role' && selectedFilter !== 'All') {
    tasks = tasks.filter((task) => task.role === selectedFilter);
  }
  if (mode === 'type' && selectedFilter !== 'All') {
    tasks = tasks.filter((task) => mappedType(task) === selectedFilter);
  }
  return tasks;
}

/**
 * Generates or retrieves task instances for a given filter and date.
 * Note: Per confirmed Stage 2 requirements, "All Roles" shows ALL owners
 * in each role rather than slicing to `.slice(0, 2)`.
 */
function buildDevTaskInstances({ mode, selectedFilter, selectedPerson, dateISO, todayISO }) {
  const selectedDateObj = parseISODate(dateISO);
  const todayObj = parseISODate(todayISO);
  const isPast =
    selectedDateObj < todayObj && !sameDay(selectedDateObj, todayObj);

  const templates = getFilteredTemplates(mode, selectedFilter);
  const rows = [];

  templates.forEach((task) => {
    const allRoleOwners = peopleByRole[task.role] || [];
    const owners =
      mode === 'role' &&
      selectedFilter !== 'All' &&
      selectedPerson !== 'All'
        ? [selectedPerson]
        : allRoleOwners;

    owners.forEach((owner, index) => {
      const uid = `${task.id}_${owner}`;
      const instanceKey = `${dateISO}__${uid}`;
      const verifier = verifierByRole[task.role] || 'Reporting Authority';

      const override = devInstanceOverrides.get(instanceKey);
      const state = override
        ? override.state
        : computeInitialState(task, index, selectedDateObj, todayObj);

      const misses = override?.misses ?? task.misses;

      let effective = state;
      if (isPast && state !== 'approved') {
        effective = 'rolled';
      }

      const history = override?.history
        ? override.history
        : buildDefaultHistory({ ...task, misses }, verifier);

      const submissionNotes =
        override?.submissionNotes ||
        (['submitted', 'approved', 'submitted_early', 'approved_early'].includes(state)
          ? `Demo submission summary for "${task.name}" filed by ${baseName(owner)} (Role: ${task.role}, Scheduled Date: ${dateISO}). All checklist items reported complete for cycle review.`
          : '');

      rows.push({
        ...task,
        misses,
        owner,
        verifier,
        state,
        effective,
        uid,
        dateISO,
        submissionNotes,
        flow: buildFlow(task, owner, state),
        history
      });
    });
  });

  return rows;
}

/**
 * 1. Fetch Dashboard Metadata (Roles, People, Task Types, Organisation Summary)
 */
export async function fetchDashboardMetadata() {
  if (!IS_USING_DEV_SEED) {
    const res = await fetch(`${API_BASE_URL}/metadata`);
    if (!res.ok) throw new Error('Failed to load dashboard metadata');
    return res.json();
  }

  const roles = Object.keys(peopleByRole).map((role) => ({
    name: role,
    count: (peopleByRole[role] || []).length,
    people: peopleByRole[role] || [],
    verifier: verifierByRole[role]
  }));

  const types = taskTypes.map((typeName) => ({
    name: typeName,
    count: taskTemplates.filter((t) => mappedType(t) === typeName).length
  }));

  return {
    organisationSummary: {
      activeTasks: 86,
      overdue: 9,
      repeatMiss: 6
    },
    roles,
    peopleByRole,
    verifierByRole,
    taskTypes: types
  };
}

/**
 * 2. Fetch Individual Tasks for Role View
 */
export async function fetchTasks({
  mode,
  selectedFilter,
  selectedPerson,
  dateISO,
  todayISO
}) {
  if (!IS_USING_DEV_SEED) {
    const params = new URLSearchParams({
      mode,
      filter: selectedFilter,
      person: selectedPerson,
      date: dateISO,
      today: todayISO
    });
    const res = await fetch(`${API_BASE_URL}/tasks?${params.toString()}`);
    if (!res.ok) throw new Error('Failed to load tasks');
    return res.json();
  }

  const rows = buildDevTaskInstances({
    mode,
    selectedFilter,
    selectedPerson,
    dateISO,
    todayISO
  });

  return {
    tasks: rows
  };
}

/**
 * 3. Fetch Aggregate Completion Data for Task Type View
 */
export async function fetchAggregateTasks({
  selectedFilter,
  isMonthly,
  dateISO,
  monthISO,
  todayISO
}) {
  if (!IS_USING_DEV_SEED) {
    const params = new URLSearchParams({
      filter: selectedFilter,
      monthly: String(isMonthly),
      date: dateISO,
      month: monthISO,
      today: todayISO
    });
    const res = await fetch(`${API_BASE_URL}/tasks/aggregate?${params.toString()}`);
    if (!res.ok) throw new Error('Failed to load aggregate tasks');
    return res.json();
  }

  const selectedDateObj = parseISODate(dateISO);
  const todayObj = parseISODate(todayISO);
  const [monthYear, monthNum] = monthISO.split('-').map(Number);

  const isFuture =
    selectedDateObj > todayObj && !sameDay(selectedDateObj, todayObj);

  const tasks = getFilteredTemplates('type', selectedFilter);
  const uniqueRoles = [...new Set(tasks.map((t) => t.role))];
  const totalPeople = uniqueRoles.reduce(
    (sum, role) => sum + (peopleByRole[role] || []).length,
    0
  );

  const items = tasks.map((task) => {
    const peopleCount = (peopleByRole[task.role] || []).length;
    const seed =
      (isMonthly ? monthNum : selectedDateObj.getDate()) + task.id.length;

    let percentage = Math.max(25, Math.min(100, 55 + ((seed * 13) % 46)));

    if (!isMonthly && isFuture) {
      percentage = Math.max(0, Math.min(35, (seed * 7) % 36));
    }

    const completed = Math.round((peopleCount * percentage) / 100);

    return {
      ...task,
      people: peopleCount,
      percentage,
      completed
    };
  });

  return {
    items,
    summary: {
      uniqueTasks: tasks.length,
      rolesCount: uniqueRoles.length,
      peopleCount: totalPeople
    }
  };
}

/**
 * 4. Fetch Single Task Detail & Activity History
 */
export async function fetchTaskDetail({ uid, dateISO, todayISO }) {
  if (!IS_USING_DEV_SEED) {
    const params = new URLSearchParams({ date: dateISO, today: todayISO });
    const res = await fetch(
      `${API_BASE_URL}/tasks/${encodeURIComponent(uid)}?${params.toString()}`
    );
    if (!res.ok) throw new Error('Failed to load task detail');
    return res.json();
  }

  const rows = buildDevTaskInstances({
    mode: 'role',
    selectedFilter: 'All',
    selectedPerson: 'All',
    dateISO,
    todayISO
  });
  return rows.find((r) => r.uid === uid) || null;
}

/**
 * 5. Fetch Task Activity History
 */
export async function fetchTaskHistory({ uid, dateISO, todayISO }) {
  if (!IS_USING_DEV_SEED) {
    const params = new URLSearchParams({ date: dateISO, today: todayISO });
    const res = await fetch(
      `${API_BASE_URL}/tasks/${encodeURIComponent(uid)}/history?${params.toString()}`
    );
    if (!res.ok) throw new Error('Failed to load task history');
    return res.json();
  }

  const detail = await fetchTaskDetail({ uid, dateISO, todayISO });
  return {
    uid,
    dateISO,
    history: detail?.history || []
  };
}

/**
 * 6. Submit Task Action (Demo Workflow)
 */
export async function submitTaskInstance({ task, dateISO, todayISO, remarks }) {
  if (!IS_USING_DEV_SEED) {
    const res = await fetch(`${API_BASE_URL}/tasks/${encodeURIComponent(task.uid)}/submit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ date: dateISO, today: todayISO, remarks })
    });
    if (!res.ok) throw new Error('Failed to submit task');
    return res.json();
  }

  const selectedDateObj = parseISODate(dateISO);
  const todayObj = parseISODate(todayISO);
  const isFuture =
    selectedDateObj > todayObj && !sameDay(selectedDateObj, todayObj);

  const nextState = isFuture ? 'submitted_early' : 'submitted';
  const instanceKey = `${dateISO}__${task.uid}`;

  const newEntry = {
    id: `act_${Date.now()}`,
    title: isFuture ? 'Submitted early (Demo Action)' : 'Task submitted (Demo Action)',
    meta: remarks
      ? `${baseName(task.owner)} — "${remarks}"`
      : `Submitted by ${baseName(task.owner)} · Waiting for ${task.verifier}`,
    colour: 'purple'
  };

  devInstanceOverrides.set(instanceKey, {
    state: nextState,
    misses: task.misses,
    history: [newEntry, ...(task.history || [])]
  });

  return { success: true, state: nextState };
}

/**
 * 7. Approve / Verify or Reject Task Action (Demo Workflow)
 */
export async function verifyTaskInstance({
  task,
  dateISO,
  todayISO,
  action, // 'approve' | 'reject'
  remarks
}) {
  if (!IS_USING_DEV_SEED) {
    const res = await fetch(`${API_BASE_URL}/tasks/${encodeURIComponent(task.uid)}/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ date: dateISO, today: todayISO, action, remarks })
    });
    if (!res.ok) throw new Error('Failed to verify task');
    return res.json();
  }

  const selectedDateObj = parseISODate(dateISO);
  const todayObj = parseISODate(todayISO);
  const isFuture =
    selectedDateObj > todayObj && !sameDay(selectedDateObj, todayObj);

  const instanceKey = `${dateISO}__${task.uid}`;

  if (action === 'approve') {
    const nextState = isFuture ? 'approved_early' : 'approved';
    const newEntry = {
      id: `act_${Date.now()}`,
      title: 'Verified & Closed (Demo Action)',
      meta: remarks
        ? `Approved by ${task.verifier} — "${remarks}"`
        : `Approved by ${task.verifier} · Task closed`,
      colour: 'green'
    };

    devInstanceOverrides.set(instanceKey, {
      state: nextState,
      misses: task.misses,
      history: [newEntry, ...(task.history || [])]
    });

    return { success: true, state: nextState };
  } else {
    // Reject returns task to owner-pending state (overdue on today/past, scheduled on future)
    const nextState = isFuture ? 'scheduled' : 'overdue';
    const newEntry = {
      id: `act_${Date.now()}`,
      title: 'Verification rejected (Demo Action)',
      meta: remarks
        ? `Returned by ${task.verifier} — "${remarks}"`
        : `Returned to ${baseName(task.owner)} by ${task.verifier} for rework`,
      colour: 'red'
    };

    devInstanceOverrides.set(instanceKey, {
      state: nextState,
      misses: task.misses,
      submissionNotes: task.submissionNotes,
      history: [newEntry, ...(task.history || [])]
    });

    return { success: true, state: nextState };
  }
}

/**
 * 8. Escalate Task Action (Demo Workflow — records history entry only, sends no real notifications)
 */
export async function escalateTaskInstance({
  task,
  dateISO,
  todayISO,
  recipient,
  reason
}) {
  if (!IS_USING_DEV_SEED) {
    const res = await fetch(
      `${API_BASE_URL}/tasks/${encodeURIComponent(task.uid)}/escalate`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          date: dateISO,
          today: todayISO,
          recipient,
          reason
        })
      }
    );
    if (!res.ok) throw new Error('Failed to record escalation');
    return res.json();
  }

  const instanceKey = `${dateISO}__${task.uid}`;
  const target = recipient || task.verifier || 'Reporting Authority';
  const newEntry = {
    id: `act_${Date.now()}`,
    title: 'Escalation recorded (Temporary Demo Mode)',
    meta: reason
      ? `Escalated to ${target} — "${reason}" (demo log only)`
      : `Demo escalation logged for review by ${target} (no external email/SMS sent)`,
    colour: 'orange'
  };

  devInstanceOverrides.set(instanceKey, {
    state: task.state,
    misses: task.misses,
    submissionNotes: task.submissionNotes,
    history: [newEntry, ...(task.history || [])]
  });

  return { success: true };
}

