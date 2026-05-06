// Shared interactions for Agent Task Dashboard prototype

document.addEventListener('DOMContentLoaded', () => {
  initDarkMode();
  initFilters();
  initDetailPanel();
  initActivitySidebar();
  initFeatureRows();
  initTaskCardClicks();
});

// Dark mode toggle
function initDarkMode() {
  const btn = document.getElementById('darkModeToggle');
  if (!btn) return;

  const stored = localStorage.getItem('theme');
  if (stored === 'dark') document.documentElement.classList.add('dark');

  btn.addEventListener('click', () => {
    document.documentElement.classList.toggle('dark');
    const isDark = document.documentElement.classList.contains('dark');
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
    updateThemeIcon(btn, isDark);
  });

  updateThemeIcon(btn, document.documentElement.classList.contains('dark'));
}

function updateThemeIcon(btn, isDark) {
  btn.innerHTML = isDark
    ? '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>'
    : '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>';
}

// Filter dropdowns
function initFilters() {
  document.querySelectorAll('.filter-dropdown').forEach(dd => {
    const btn = dd.querySelector('.filter-btn');
    const popup = dd.querySelector('.filter-popup');

    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      // Close other popups
      document.querySelectorAll('.filter-popup.open').forEach(p => { if (p !== popup) p.classList.remove('open'); });
      popup.classList.toggle('open');
    });

    popup.addEventListener('change', () => {
      const checked = popup.querySelectorAll('input[type="checkbox"]:checked');
      const countEl = btn.querySelector('.filter-count');
      if (countEl) {
        countEl.textContent = checked.length;
        countEl.style.display = checked.length > 0 ? 'inline-flex' : 'none';
      }
      applyFilters();
    });
  });

  document.addEventListener('click', () => {
    document.querySelectorAll('.filter-popup.open').forEach(p => p.classList.remove('open'));
  });
}

function applyFilters() {
  const statusChecked = getCheckedValues('statusFilter');
  const priorityChecked = getCheckedValues('priorityFilter');

  document.querySelectorAll('.feature-row').forEach(row => {
    if (statusChecked.length === 0 && priorityChecked.length === 0) {
      row.style.display = '';
      return;
    }

    const tasks = row.querySelectorAll('.task-card');
    let match = false;
    tasks.forEach(card => {
      const status = card.dataset.status;
      const priority = card.dataset.priority;
      const statusOk = statusChecked.length === 0 || statusChecked.includes(status);
      const priorityOk = priorityChecked.length === 0 || priorityChecked.includes(priority);
      if (statusOk && priorityOk) match = true;
    });

    row.style.display = match ? '' : 'none';
  });
}

function getCheckedValues(name) {
  const checks = document.querySelectorAll(`input[name="${name}"]:checked`);
  return Array.from(checks).map(c => c.value);
}

// Detail panel (UF-3)
let lastClickedCard = null;

function initDetailPanel() {
  const overlay = document.getElementById('detailOverlay');
  const panel = document.getElementById('detailPanel');
  if (!overlay || !panel) return;

  // Close handlers
  document.getElementById('closeDetail')?.addEventListener('click', closePanel);
  overlay.addEventListener('click', closePanel);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && panel.classList.contains('active')) closePanel();
  });
}

function openPanel(taskId) {
  const overlay = document.getElementById('detailOverlay');
  const panel = document.getElementById('detailPanel');
  const sidebar = document.getElementById('activitySidebar');
  if (!overlay || !panel) return;

  lastClickedCard = document.querySelector(`.task-card[data-task-id="${taskId}"]`);

  // Populate mock data
  populatePanel(taskId);

  overlay.classList.add('active');
  panel.classList.add('active');
  if (sidebar) sidebar.classList.add('hidden');

  // Focus trap
  panel.focus();
}

function closePanel() {
  const overlay = document.getElementById('detailOverlay');
  const panel = document.getElementById('detailPanel');
  const sidebar = document.getElementById('activitySidebar');
  if (!overlay || !panel) return;

  overlay.classList.remove('active');
  panel.classList.remove('active');
  if (sidebar) sidebar.classList.remove('hidden');

  // Return focus
  if (lastClickedCard) lastClickedCard.focus();
}

function populatePanel(taskId) {
  // Mock data for prototype
  const mockTasks = {
    '1.1': { title: 'Set up project scaffolding', status: 'completed', priority: 'P0', scope: 'all', est: '1h', breaking: false, deps: [], file: 'tasks/1.1-scaffolding.md', record: 'records/1.1-scaffolding.md', criteria: ['Go module initialized with Gin dependency', 'Directory structure matches tech design', 'Config file loading works'], recordContent: '<p>Completed successfully. Go module initialized, Gin added, directory structure created per tech design.</p>' },
    '1.2': { title: 'Implement config file parser', status: 'completed', priority: 'P0', scope: 'backend', est: '2h', breaking: false, deps: ['1.1'], file: 'tasks/1.2-config-parser.md', record: 'records/1.2-config-parser.md', criteria: ['YAML config file parsed correctly', 'Invalid path shows warning', 'Multiple projects supported'], recordContent: '<p>Implemented YAML parser with validation. Handles missing paths gracefully.</p>' },
    '2.1': { title: 'Implement filesystem scanner', status: 'in_progress', priority: 'P0', scope: 'backend', est: '3h', breaking: false, deps: ['1.2'], file: 'tasks/2.1-scanner.md', record: null, criteria: ['Scans all feature directories', 'Parses index.json files', 'Returns structured data'], recordContent: null },
    '2.2': { title: 'Build landing page handler', status: 'pending', priority: 'P1', scope: 'frontend', est: '2h', breaking: false, deps: ['2.1'], file: 'tasks/2.2-landing.md', record: null, criteria: ['Renders project cards', 'Shows health status', 'Progress bars work'], recordContent: null },
    '3.1': { title: 'Implement swimlane renderer', status: 'blocked', priority: 'P0', scope: 'frontend', est: '4h', breaking: true, deps: ['2.1', '1.3-deps'], file: 'tasks/3.1-swimlane.md', record: null, criteria: ['Renders feature rows', 'Phase columns correct', 'Dependency arrows visible'], recordContent: null },
    'T-1': { title: 'Integration test for scanner', status: 'pending', priority: 'P1', scope: 'all', est: '1h', breaking: true, deps: ['2.1'], file: 'tasks/T-1-scanner-test.md', record: null, criteria: ['Test with 3 projects', 'Test with invalid config', 'Test with empty features'], recordContent: null }
  };

  const task = mockTasks[taskId] || mockTasks['2.1'];

  document.getElementById('detailTaskId').textContent = taskId;
  document.getElementById('detailTitle').textContent = task.title;
  document.getElementById('detailStatusDot').className = `status-dot status-dot-${task.status}`;
  document.getElementById('detailStatusText').textContent = task.status.replace('_', ' ');
  document.getElementById('detailPriority').textContent = task.priority;
  document.getElementById('detailScope').textContent = task.scope;
  document.getElementById('detailEstTime').textContent = task.est;
  document.getElementById('detailBreaking').textContent = task.breaking ? 'Yes' : 'No';
  document.getElementById('detailBreaking').style.color = task.breaking ? 'var(--destructive)' : '';
  document.getElementById('detailFilePath').textContent = task.file;
  document.getElementById('detailRecordPath').textContent = task.record || '—';

  // Dependencies
  const depsEl = document.getElementById('detailDeps');
  depsEl.innerHTML = task.deps.length === 0
    ? '<span class="text-muted text-sm">None</span>'
    : task.deps.map(d => `<span class="chip" onclick="navigateToTask('${d}')">${d}</span>`).join(' ');

  // Acceptance criteria
  const criteriaEl = document.getElementById('detailCriteria');
  criteriaEl.innerHTML = task.criteria.map(c =>
    `<li class="criteria-item"><span class="criteria-icon">&#9675;</span><span>${c}</span></li>`
  ).join('');

  // Execution record
  const recordEl = document.getElementById('detailRecord');
  if (task.recordContent) {
    recordEl.innerHTML = `<div class="record-content">${task.recordContent}</div>`;
  } else {
    recordEl.innerHTML = '<div class="record-empty"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>No execution record</div>';
  }
}

function navigateToTask(taskId) {
  closePanel();
  setTimeout(() => {
    const card = document.querySelector(`.task-card[data-task-id="${taskId}"]`);
    if (!card) return;

    // Expand parent row if collapsed
    const row = card.closest('.feature-row');
    if (row && row.classList.contains('collapsed')) {
      row.classList.remove('collapsed');
      const taskArea = row.querySelector('.feature-task-area');
      if (taskArea) taskArea.style.display = '';
    }

    card.scrollIntoView({ behavior: 'smooth', block: 'center' });
    card.classList.add('highlighted');
    setTimeout(() => {
      card.classList.remove('highlighted');
    }, 2000);
  }, 100);
}

// Activity sidebar (UF-4)
function initActivitySidebar() {
  const sidebar = document.getElementById('activitySidebar');
  if (!sidebar) return;

  const collapseBtn = sidebar.querySelector('.sidebar-collapse-btn');
  const expandBtn = sidebar.querySelector('.sidebar-expand-btn');

  collapseBtn?.addEventListener('click', () => {
    sidebar.classList.add('collapsed');
    sidebar.querySelector('.sidebar-expanded').style.display = 'none';
    sidebar.querySelector('.sidebar-collapsed-view').style.display = '';
  });

  expandBtn?.addEventListener('click', () => {
    sidebar.classList.remove('collapsed');
    sidebar.querySelector('.sidebar-expanded').style.display = '';
    sidebar.querySelector('.sidebar-collapsed-view').style.display = 'none';
  });
}

// Feature row collapse/expand
function initFeatureRows() {
  document.querySelectorAll('.feature-row-header').forEach(header => {
    header.addEventListener('click', () => {
      const row = header.closest('.feature-row');
      const taskArea = row.querySelector('.feature-task-area');
      if (!taskArea) return;

      const chevron = header.querySelector('.chevron');
      if (taskArea.style.display === 'none') {
        taskArea.style.display = '';
        if (chevron) chevron.style.transform = 'rotate(0deg)';
        row.classList.remove('collapsed');
      } else {
        taskArea.style.display = 'none';
        if (chevron) chevron.style.transform = 'rotate(-90deg)';
        row.classList.add('collapsed');
      }
    });
  });
}

// Task card click → open detail panel
function initTaskCardClicks() {
  document.querySelectorAll('.task-card').forEach(card => {
    card.addEventListener('click', () => {
      openPanel(card.dataset.taskId);
    });
    card.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        openPanel(card.dataset.taskId);
      }
    });
  });

  // Activity event clicks
  document.querySelectorAll('.event-item').forEach(item => {
    item.addEventListener('click', () => {
      const taskId = item.dataset.taskId;
      navigateToTask(taskId);
    });
  });
}
