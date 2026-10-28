/* =========================================================
   DEVELOPMENT SEED DATA (STAGE 2 — DEMO ONLY)
   ---------------------------------------------------------
   IMPORTANT: This file contains development seed data derived
   from "Overall Task View.html" for local frontend testing
   before the Go/MySQL backend is connected.
   - Do not treat these rules or synthetic names as production
     business logic.
   - Preserves the prototype's current task-type mapping.
   - Shows all owners in the "All Roles" view (no .slice(0,2)).
   ========================================================= */

export const DEFAULT_DEMO_TODAY =
  import.meta.env.VITE_DEMO_TODAY || '2026-10-11';

export const peopleByRole = {
  'Principal': [
    'Dr. Priya Raman'
  ],
  'IQAC Head': [
    'Dr. Senthil Kumar'
  ],
  'Head of Department': [
    'Dr. Meena S — CSE',
    'Dr. Revathi P — ECE',
    'Dr. Anand K — Mechanical',
    'Dr. Deepika R — IT',
    'Dr. Vignesh M — EEE'
  ],
  'Lab Incharge': [
    'Mr. Sathish K — Advanced Manufacturing',
    'Ms. Harini V — AI Lab',
    'Mr. Naveen P — Electronics Lab'
  ],
  'Transport Incharge': [
    'Mr. Suresh Kumar'
  ],
  'Maintenance Head': [
    'Mr. Ramesh B'
  ],
  'Admin Head': [
    'Ms. Anitha R',
    'Mr. Vishal K'
  ],
  'Finance Head': [
    'Mr. Kannan R'
  ],
  'Placement Head': [
    'Ms. Kavitha S',
    'Mr. Arun P'
  ],
  'R&D Head': [
    'Dr. Mani V',
    'Dr. Nirmala S'
  ],
  'Industry Relations Head': [
    'Ms. Deepa K'
  ],
  'Alumni Head': [
    'Dr. Lavanya P'
  ]
};

export const verifierByRole = {
  'Principal': 'Chairman / Management',
  'IQAC Head': 'Principal',
  'Head of Department': 'Principal / Dean Academics',
  'Lab Incharge': 'Head of Department',
  'Transport Incharge': 'Admin Head',
  'Maintenance Head': 'Admin Head',
  'Admin Head': 'Principal',
  'Finance Head': 'Management',
  'Placement Head': 'Principal / Dean Placements',
  'R&D Head': 'Principal / Dean R&D',
  'Industry Relations Head': 'Principal',
  'Alumni Head': 'Principal'
};

export const taskTypes = [
  'Daily Tasks',
  'Weekly Tasks',
  'Monthly Tasks',
  'Quarterly Tasks',
  'Yearly Tasks',
  'One Time Tasks',
  'Statutory Tasks',
  'Finance Tasks',
  'Academic Tasks',
  'Others'
];

export const taskTemplates = [
  {
    id: 't1',
    name: 'Daily Teaching-Learning Closure',
    role: 'Principal',
    category: 'Teaching & Learning',
    frequency: 'Daily Tasks',
    time: '5:30 PM',
    pattern: 'submitted',
    misses: 1
  },
  {
    id: 't2',
    name: 'Academic Deviation & Faculty Absence Review',
    role: 'Principal',
    category: 'Teaching & Learning',
    frequency: 'Daily Tasks',
    time: '4:45 PM',
    pattern: 'approved',
    misses: 0
  },
  {
    id: 't3',
    name: 'Student Discipline Exception Closure',
    role: 'Principal',
    category: 'Student Discipline',
    frequency: 'Daily Tasks',
    time: '5:00 PM',
    pattern: 'overdue',
    misses: 2
  },
  {
    id: 't4',
    name: 'Accreditation Evidence Update',
    role: 'IQAC Head',
    category: 'Documentation',
    frequency: 'Monthly Tasks',
    time: 'Month End',
    pattern: 'submitted',
    misses: 0
  },
  {
    id: 't5',
    name: 'Internal Quality Audit Review',
    role: 'IQAC Head',
    category: 'Statutory Committees',
    frequency: 'Quarterly Tasks',
    time: 'Quarter End',
    pattern: 'approved',
    misses: 0
  },
  {
    id: 't6',
    name: 'Department Teaching Progress Review',
    role: 'Head of Department',
    category: 'Teaching & Learning',
    frequency: 'Daily Tasks',
    time: '5:00 PM',
    pattern: 'submitted',
    misses: 0
  },
  {
    id: 't7',
    name: 'Student Competency Gap Follow-up',
    role: 'Head of Department',
    category: 'Students Competency',
    frequency: 'Weekly Tasks',
    time: 'Friday',
    pattern: 'overdue',
    misses: 2
  },
  {
    id: 't8',
    name: 'Faculty Action Closure',
    role: 'Head of Department',
    category: 'Faculty Achievement',
    frequency: 'Daily Tasks',
    time: '5:30 PM',
    pattern: 'approved',
    misses: 0
  },
  {
    id: 't9',
    name: 'Department Monthly Performance Review',
    role: 'Head of Department',
    category: 'Data Management',
    frequency: 'Monthly Tasks',
    time: '10th',
    pattern: 'submitted',
    misses: 1
  },
  {
    id: 't10',
    name: 'Laboratory Utilisation Closure',
    role: 'Lab Incharge',
    category: 'Infrastructure Utilisation',
    frequency: 'Daily Tasks',
    time: '5:00 PM',
    pattern: 'approved',
    misses: 0
  },
  {
    id: 't11',
    name: 'Equipment Maintenance Exception Review',
    role: 'Lab Incharge',
    category: 'Infrastructure Maintenance',
    frequency: 'Weekly Tasks',
    time: 'Saturday',
    pattern: 'submitted',
    misses: 0
  },
  {
    id: 't12',
    name: 'Bus Route Readiness Confirmation',
    role: 'Transport Incharge',
    category: 'Infrastructure Maintenance',
    frequency: 'Daily Tasks',
    time: '6:30 AM',
    pattern: 'approved',
    misses: 0
  },
  {
    id: 't13',
    name: 'Daily Maintenance Issue Closure',
    role: 'Maintenance Head',
    category: 'Infrastructure Maintenance',
    frequency: 'Daily Tasks',
    time: '5:00 PM',
    pattern: 'overdue',
    misses: 3
  },
  {
    id: 't14',
    name: 'Master Data Integrity Review',
    role: 'Admin Head',
    category: 'Data Management',
    frequency: 'Monthly Tasks',
    time: '10th',
    pattern: 'approved',
    misses: 0
  },
  {
    id: 't15',
    name: 'Inventory Exception Closure',
    role: 'Admin Head',
    category: 'Inventory',
    frequency: 'Monthly Tasks',
    time: '15th',
    pattern: 'submitted',
    misses: 1
  },
  {
    id: 't16',
    name: 'Monthly Budget Variance Closure',
    role: 'Finance Head',
    category: 'Finance',
    frequency: 'Monthly Tasks',
    time: '5th',
    pattern: 'overdue',
    misses: 3
  },
  {
    id: 't17',
    name: 'Daily Payment Approval Queue',
    role: 'Finance Head',
    category: 'Finance',
    frequency: 'Daily Tasks',
    time: '4:30 PM',
    pattern: 'submitted',
    misses: 0
  },
  {
    id: 't18',
    name: 'Placement Training Outcome Review',
    role: 'Placement Head',
    category: 'Placement Training',
    frequency: 'Weekly Tasks',
    time: 'Friday',
    pattern: 'submitted',
    misses: 1
  },
  {
    id: 't19',
    name: 'Placement Drive Data Closure',
    role: 'Placement Head',
    category: 'Placement',
    frequency: 'Daily Tasks',
    time: '5:30 PM',
    pattern: 'overdue',
    misses: 2
  },
  {
    id: 't20',
    name: 'Sponsored Research Progress Review',
    role: 'R&D Head',
    category: 'R&D',
    frequency: 'Monthly Tasks',
    time: '7th',
    pattern: 'submitted',
    misses: 1
  },
  {
    id: 't21',
    name: 'Faculty Publication Validation',
    role: 'R&D Head',
    category: 'Faculty Achievement',
    frequency: 'Monthly Tasks',
    time: '12th',
    pattern: 'approved',
    misses: 0
  },
  {
    id: 't22',
    name: 'Industry Interaction Pipeline Review',
    role: 'Industry Relations Head',
    category: 'Industry Relations',
    frequency: 'Weekly Tasks',
    time: 'Thursday',
    pattern: 'submitted',
    misses: 0
  },
  {
    id: 't23',
    name: 'Alumni Mentor Engagement Review',
    role: 'Alumni Head',
    category: 'Alumni Engagement',
    frequency: 'Monthly Tasks',
    time: '14th',
    pattern: 'approved',
    misses: 0
  }
];

/**
 * Preserves the exact prototype task-type mapping (`mappedType` in HTML).
 * Note for Project Owner: Categories like "Statutory Committees", "Finance",
 * and "Teaching & Learning" override the task's frequency when filtering by
 * Task Type.
 */
export function mappedType(task) {
  if (task.category === 'Statutory Committees') {
    return 'Statutory Tasks';
  }
  if (task.category === 'Finance') {
    return 'Finance Tasks';
  }
  if (
    [
      'Teaching & Learning',
      'Exam & Assessment',
      'Students Competency'
    ].includes(task.category)
  ) {
    return 'Academic Tasks';
  }
  return task.frequency;
}

export function baseName(name) {
  return name.split(' — ')[0];
}

export function parseISODate(dateStr) {
  const [y, m, d] = dateStr.split('-').map(Number);
  return new Date(y, m - 1, d);
}

export function toISODate(dateObj) {
  const y = dateObj.getFullYear();
  const m = String(dateObj.getMonth() + 1).padStart(2, '0');
  const d = String(dateObj.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

export function sameDay(a, b) {
  return a.toDateString() === b.toDateString();
}

/**
 * Builds the 4-step workflow array matching `buildFlow()` in the HTML prototype.
 */
export function buildFlow(task, owner, state) {
  const verifier = verifierByRole[task.role] || 'Reporting Authority';
  const ownerShort = baseName(owner);

  if (state === 'approved_early') {
    return [
      { status: 'done', no: '1', title: 'Completed early', person: ownerShort },
      { status: 'done', no: '2', title: 'Submitted early', person: ownerShort },
      { status: 'done', no: '3', title: 'Verified early', person: verifier },
      { status: 'done', no: '4', title: 'Closed', person: 'System' }
    ];
  }

  if (state === 'submitted_early') {
    return [
      { status: 'done', no: '1', title: 'Completed early', person: ownerShort },
      { status: 'done', no: '2', title: 'Submitted early', person: ownerShort },
      { status: 'wait', no: '3', title: 'Waiting verification', person: verifier },
      { status: 'future', no: '4', title: 'Closure pending', person: 'System' }
    ];
  }

  if (state === 'scheduled') {
    return [
      { status: 'future', no: '1', title: 'Yet to start', person: ownerShort },
      { status: 'future', no: '2', title: 'Submission', person: 'Scheduled' },
      { status: 'future', no: '3', title: 'Verification', person: verifier },
      { status: 'future', no: '4', title: 'Closure', person: 'Scheduled' }
    ];
  }

  if (state === 'approved') {
    return [
      { status: 'done', no: '1', title: 'Work completed', person: ownerShort },
      { status: 'done', no: '2', title: 'Submitted', person: ownerShort },
      { status: 'done', no: '3', title: 'Verified', person: verifier },
      { status: 'done', no: '4', title: 'Closed', person: 'System' }
    ];
  }

  if (state === 'submitted') {
    return [
      { status: 'done', no: '1', title: 'Work completed', person: ownerShort },
      { status: 'done', no: '2', title: 'Submitted', person: ownerShort },
      { status: 'wait', no: '3', title: 'Waiting verification', person: verifier },
      { status: '', no: '4', title: 'Closure pending', person: 'System' }
    ];
  }

  // Overdue or rejected
  return [
    { status: 'current', no: '1', title: 'Owner pending', person: ownerShort },
    { status: '', no: '2', title: 'Submission', person: 'Not filed' },
    { status: '', no: '3', title: 'Verification', person: verifier },
    { status: '', no: '4', title: 'Closure', person: 'Pending' }
  ];
}

/**
 * Generates initial historical activity items for a task instance.
 */
export function buildDefaultHistory(task, verifier) {
  const items = [
    {
      id: `${task.id}_h1`,
      title: 'Previous cycle completed',
      meta: 'Submission received from task owner',
      colour: 'green'
    },
    {
      id: `${task.id}_h2`,
      title: 'Verification completed',
      meta: `Reviewed by ${verifier}`,
      colour: 'green'
    }
  ];

  if (task.misses > 0) {
    items.push({
      id: `${task.id}_h3`,
      title: 'Historical delay',
      meta: `${task.misses} previous missed / delayed cycle(s)`,
      colour: 'red'
    });
  }

  items.push({
    id: `${task.id}_h4`,
    title: 'Recurring responsibility',
    meta: `${task.frequency} task mapped to ${task.role}`,
    colour: 'gray'
  });

  return items;
}

