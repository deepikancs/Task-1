import React, { useState, useEffect, useCallback } from 'react';
import TopBar from './components/TopBar';
import LeftPanel from './components/LeftPanel';
import MiddleHeader from './components/MiddleHeader';
import PersonBar from './components/PersonBar';
import PeriodSelector from './components/PeriodSelector';
import SummaryStrip from './components/SummaryStrip';
import TaskSegmentList from './components/TaskSegmentList';
import AggregateList from './components/AggregateList';
import DetailPanel from './components/DetailPanel';
import TaskAnalyticsModal from './components/TaskAnalyticsModal';
import AssignTaskModal from './components/AssignTaskModal';
import {
  fetchDashboardMetadata,
  fetchTasks,
  fetchAggregateTasks,
  submitTaskInstance,
  verifyTaskInstance,
  escalateTaskInstance
} from './api/taskApi';
import {
  DEFAULT_DEMO_TODAY,
  parseISODate,
  toISODate,
  sameDay
} from './seed/devSeedData';

export default function App() {
  // Configurable Demo Today Reference Date
  const [todayISO, setTodayISO] = useState(DEFAULT_DEMO_TODAY);
  const todayDate = parseISODate(todayISO);

  // Filter & Selection State
  const [mode, setMode] = useState('role');
  const [selectedFilter, setSelectedFilter] = useState('All');
  const [selectedPerson, setSelectedPerson] = useState('All');
  const [selectedTaskId, setSelectedTaskId] = useState(null);

  // Period State
  const [selectedDate, setSelectedDate] = useState(() =>
    parseISODate(DEFAULT_DEMO_TODAY)
  );
  const [selectedMonth, setSelectedMonth] = useState(() => {
    const d = parseISODate(DEFAULT_DEMO_TODAY);
    return new Date(d.getFullYear(), d.getMonth(), 1);
  });

  // API Data State
  const [metadata, setMetadata] = useState({
    demoMode: true,
    organisationSummary: { activeTasks: 86, overdue: 9, repeatMiss: 6 },
    roles: [],
    peopleByRole: {},
    verifierByRole: {},
    taskTypes: []
  });
  const [tasks, setTasks] = useState([]);
  const [aggregateData, setAggregateData] = useState({
    items: [],
    summary: { uniqueTasks: 0, rolesCount: 0, peopleCount: 0 }
  });

  // Dialog & Error State
  const [analyticsOpen, setAnalyticsOpen] = useState(false);
  const [assignModalOpen, setAssignModalOpen] = useState(false);
  const [errorNotice, setErrorNotice] = useState(null);

  const isMonthlyMode =
    mode === 'type' &&
    ['Monthly Tasks', 'Quarterly Tasks', 'Yearly Tasks'].includes(
      selectedFilter
    );

  const isPastDate =
    selectedDate < todayDate && !sameDay(selectedDate, todayDate);
  const isFutureDate =
    selectedDate > todayDate && !sameDay(selectedDate, todayDate);

  // 1. Load Metadata on mount
  useEffect(() => {
    fetchDashboardMetadata()
      .then((data) => setMetadata(data))
      .catch((err) => {
        console.error('Metadata load error:', err);
        setErrorNotice(
          'Unable to reach the Go API on http://localhost:8080. Start the backend server or enable VITE_USE_DEV_SEED=true.'
        );
      });
  }, []);

  // 2. Load Tasks or Aggregate Data whenever filters or period change
  const loadWorkspaceData = useCallback(async () => {
    const dateISO = toISODate(selectedDate);
    const monthISO = `${selectedMonth.getFullYear()}-${String(
      selectedMonth.getMonth() + 1
    ).padStart(2, '0')}`;

    try {
      setErrorNotice(null);
      if (mode === 'role') {
        const res = await fetchTasks({
          mode,
          selectedFilter,
          selectedPerson,
          dateISO,
          todayISO
        });
        setTasks(res.tasks || []);

        setSelectedTaskId((prevId) => {
          if (prevId && res.tasks.some((t) => t.uid === prevId)) {
            return prevId;
          }
          return res.tasks.length > 0 ? res.tasks[0].uid : null;
        });
      } else {
        const res = await fetchAggregateTasks({
          selectedFilter,
          isMonthly: isMonthlyMode,
          dateISO,
          monthISO,
          todayISO
        });
        setAggregateData(res);
      }
    } catch (err) {
      console.error('Failed to load workspace data:', err);
      setErrorNotice(
        'Could not load tasks from the API. Check that the Go backend is running on port 8080.'
      );
    }
  }, [
    mode,
    selectedFilter,
    selectedPerson,
    selectedDate,
    selectedMonth,
    todayISO,
    isMonthlyMode
  ]);

  useEffect(() => {
    loadWorkspaceData();
  }, [loadWorkspaceData]);

  // Handlers
  const handleSelectMode = (newMode) => {
    setMode(newMode);
    setSelectedFilter('All');
    setSelectedPerson('All');
    setSelectedTaskId(null);
  };

  const handleSelectFilter = (filter) => {
    setSelectedFilter(filter);
    setSelectedPerson('All');
    setSelectedTaskId(null);
  };

  const handleSelectPerson = (person) => {
    setSelectedPerson(person);
    setSelectedTaskId(null);
  };

  const handleJumpToRoleFromAggregate = (roleName) => {
    setMode('role');
    setSelectedFilter(roleName);
    setSelectedPerson('All');
    setSelectedTaskId(null);
  };

  const handleChooseDate = (year, month, date) => {
    setSelectedDate(new Date(year, month, date));
    setSelectedTaskId(null);
  };

  const handleChooseMonth = (year, month) => {
    setSelectedMonth(new Date(year, month, 1));
    setSelectedTaskId(null);
  };

  const handleShiftPeriod = (direction) => {
    if (isMonthlyMode) {
      setSelectedMonth(
        (prev) => new Date(prev.getFullYear(), prev.getMonth() + direction, 1)
      );
    } else {
      setSelectedDate((prev) => {
        const next = new Date(prev);
        next.setDate(prev.getDate() + direction);
        return next;
      });
    }
    setSelectedTaskId(null);
  };

  const handleGoCurrentPeriod = () => {
    const current = parseISODate(todayISO);
    if (isMonthlyMode) {
      setSelectedMonth(new Date(current.getFullYear(), current.getMonth(), 1));
    } else {
      setSelectedDate(current);
    }
    setSelectedTaskId(null);
  };

  const handleChangeTodayISO = (newTodayISO) => {
    setTodayISO(newTodayISO);
    const parsed = parseISODate(newTodayISO);
    setSelectedDate(parsed);
    setSelectedMonth(new Date(parsed.getFullYear(), parsed.getMonth(), 1));
    setSelectedTaskId(null);
  };

  const handleSubmitTask = async (task) => {
    try {
      await submitTaskInstance({
        task,
        dateISO: toISODate(selectedDate),
        todayISO
      });
      await loadWorkspaceData();
    } catch (err) {
      setErrorNotice(err.message || 'Failed to submit task');
    }
  };

  const handleVerifyTask = async (task, action) => {
    try {
      await verifyTaskInstance({
        task,
        dateISO: toISODate(selectedDate),
        todayISO,
        action
      });
      await loadWorkspaceData();
    } catch (err) {
      setErrorNotice(err.message || 'Failed to verify task');
    }
  };

  const handleEscalateTask = async (task, recipient, reason) => {
    try {
      await escalateTaskInstance({
        task,
        dateISO: toISODate(selectedDate),
        todayISO,
        recipient,
        reason
      });
      await loadWorkspaceData();
    } catch (err) {
      setErrorNotice(err.message || 'Failed to record escalation');
    }
  };

  const selectedTask = tasks.find((t) => t.uid === selectedTaskId) || null;

  return (
    <>
      <TopBar
        todayISO={todayISO}
        onChangeTodayISO={handleChangeTodayISO}
        backendDemoMode={Boolean(metadata.demoMode)}
      />

      <main className="workspace">
        <LeftPanel
          organisationSummary={metadata.organisationSummary}
          mode={mode}
          onSelectMode={handleSelectMode}
          selectedFilter={selectedFilter}
          onSelectFilter={handleSelectFilter}
          roles={metadata.roles}
          taskTypes={metadata.taskTypes}
        />

        <section className="middle-panel">
          <MiddleHeader
            mode={mode}
            selectedFilter={selectedFilter}
            selectedPerson={selectedPerson}
            isMonthlyMode={isMonthlyMode}
            onOpenAnalytics={() => setAnalyticsOpen(true)}
            onOpenAssignTask={() => setAssignModalOpen(true)}
          />

          <PersonBar
            mode={mode}
            selectedFilter={selectedFilter}
            selectedPerson={selectedPerson}
            onSelectPerson={handleSelectPerson}
            peopleByRole={metadata.peopleByRole}
            aggregateSummary={aggregateData.summary}
          />

          <PeriodSelector
            isMonthlyMode={isMonthlyMode}
            selectedDate={selectedDate}
            selectedMonth={selectedMonth}
            todayDate={todayDate}
            onChooseDate={handleChooseDate}
            onChooseMonth={handleChooseMonth}
            onShiftPeriod={handleShiftPeriod}
            onGoCurrentPeriod={handleGoCurrentPeriod}
          />

          <SummaryStrip
            mode={mode}
            isPastDate={isPastDate}
            isFutureDate={isFutureDate}
            tasks={tasks}
            aggregateSummary={aggregateData.summary}
          />

          {mode === 'role' ? (
            <TaskSegmentList
              tasks={tasks}
              isPastDate={isPastDate}
              isFutureDate={isFutureDate}
              selectedTaskId={selectedTaskId}
              onSelectTask={setSelectedTaskId}
            />
          ) : (
            <AggregateList
              items={aggregateData.items}
              isMonthlyMode={isMonthlyMode}
              onSelectRole={handleJumpToRoleFromAggregate}
            />
          )}
        </section>

        <DetailPanel
          mode={mode}
          selectedFilter={selectedFilter}
          selectedTask={selectedTask}
          selectedDate={selectedDate}
          isPastDate={isPastDate}
          onSubmitTask={handleSubmitTask}
          onVerifyTask={handleVerifyTask}
          onEscalateTask={handleEscalateTask}
          errorNotice={errorNotice}
          onClearError={() => setErrorNotice(null)}
        />
      </main>

      <TaskAnalyticsModal
        isOpen={analyticsOpen}
        onClose={() => setAnalyticsOpen(false)}
        tasks={tasks}
        selectedFilter={selectedFilter}
        selectedPerson={selectedPerson}
        selectedDate={selectedDate}
      />

      <AssignTaskModal
        isOpen={assignModalOpen}
        onClose={() => setAssignModalOpen(false)}
        roles={metadata.roles}
      />
    </>
  );
}
