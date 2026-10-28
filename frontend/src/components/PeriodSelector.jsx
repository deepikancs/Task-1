import React from 'react';
import { sameDay } from '../seed/devSeedData';

function makeDays(selectedDate) {
  const output = [];
  for (let i = -7; i <= 8; i++) {
    const date = new Date(selectedDate);
    date.setDate(selectedDate.getDate() + i);
    output.push(date);
  }
  return output;
}

function makeMonths(selectedMonth) {
  const output = [];
  for (let i = -6; i <= 7; i++) {
    const date = new Date(selectedMonth);
    date.setMonth(selectedMonth.getMonth() + i);
    output.push(date);
  }
  return output;
}

export default function PeriodSelector({
  isMonthlyMode,
  selectedDate,
  selectedMonth,
  todayDate,
  onChooseDate,
  onChooseMonth,
  onShiftPeriod,
  onGoCurrentPeriod
}) {
  const periods = isMonthlyMode
    ? makeMonths(selectedMonth)
    : makeDays(selectedDate);

  return (
    <div className="period-shell">
      <div className="period-head">
        <div className="period-title">
          {isMonthlyMode ? 'Select Month' : 'Select Date'}
        </div>

        <div className="period-actions">
          <button
            type="button"
            className="period-nav"
            aria-label="Previous period"
            onClick={() => onShiftPeriod(-1)}
          >
            ‹
          </button>

          <button
            type="button"
            className="period-nav"
            aria-label="Current period"
            onClick={onGoCurrentPeriod}
          >
            ●
          </button>

          <button
            type="button"
            className="period-nav"
            aria-label="Next period"
            onClick={() => onShiftPeriod(1)}
          >
            ›
          </button>
        </div>
      </div>

      <div className="period-scroll">
        {periods.map((date) => {
          if (isMonthlyMode) {
            const active =
              date.getFullYear() === selectedMonth.getFullYear() &&
              date.getMonth() === selectedMonth.getMonth();

            return (
              <button
                key={`${date.getFullYear()}-${date.getMonth()}`}
                type="button"
                className={`period-item ${active ? 'active' : ''}`}
                onClick={() =>
                  onChooseMonth(date.getFullYear(), date.getMonth())
                }
              >
                <div className="period-day">
                  {date.toLocaleString('en', { month: 'short' })}
                </div>
                <div className="period-number">{date.getFullYear()}</div>
                <div className="period-small">View tasks</div>
              </button>
            );
          }

          const periodText =
            date < todayDate && !sameDay(date, todayDate)
              ? 'Past'
              : sameDay(date, todayDate)
              ? 'Today'
              : 'Future';

          const active = sameDay(date, selectedDate);

          return (
            <button
              key={date.toDateString()}
              type="button"
              className={`period-item ${active ? 'active' : ''}`}
              onClick={() =>
                onChooseDate(
                  date.getFullYear(),
                  date.getMonth(),
                  date.getDate()
                )
              }
            >
              <div className="period-day">
                {date.toLocaleString('en', { weekday: 'short' })}
              </div>
              <div className="period-number">{date.getDate()}</div>
              <div className="period-small">{periodText}</div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

