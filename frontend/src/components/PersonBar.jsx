import React from 'react';

export default function PersonBar({
  mode,
  selectedFilter,
  selectedPerson,
  onSelectPerson,
  peopleByRole,
  aggregateSummary
}) {
  if (mode === 'role' && selectedFilter !== 'All') {
    const people = peopleByRole[selectedFilter] || [];
    const pluralSuffix = people.length > 1 ? 's' : '';

    return (
      <div className="person-bar">
        <span className="person-label">Specific Person</span>

        <select
          className="person-select"
          aria-label="Select Specific Person"
          value={selectedPerson}
          onChange={(e) => onSelectPerson(e.target.value)}
        >
          <option value="All">
            All {selectedFilter}
            {pluralSuffix}
          </option>
          {people.map((person) => (
            <option key={person} value={person}>
              {person}
            </option>
          ))}
        </select>

        <span className="person-context">
          {people.length} person{people.length === 1 ? '' : 's'} in this role
        </span>
      </div>
    );
  }

  if (mode === 'role') {
    return (
      <div className="person-bar">
        <span className="person-label">Scope</span>
        <span className="person-context" style={{ marginLeft: 0 }}>
          <strong>All roles and all task owners</strong>
        </span>
      </div>
    );
  }

  return (
    <div className="person-bar">
      <span className="person-label">Comparison View</span>
      <span className="person-context" style={{ marginLeft: 0 }}>
        <strong>
          {selectedFilter === 'All' ? 'All Task Types' : selectedFilter}
        </strong>
        {' · '}
        {aggregateSummary?.rolesCount ?? 0} roles
        {' · '}
        {aggregateSummary?.peopleCount ?? 0} people represented
      </span>
    </div>
  );
}

