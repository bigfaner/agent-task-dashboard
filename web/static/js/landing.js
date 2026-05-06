/**
 * landing.js — Project Card Grid renderer for the Task Dashboard landing page.
 *
 * Data is injected by Go template as window.__PROJECTS_DATA__ (JSON array of project objects).
 * Uses document.createElement — no framework needed.
 *
 * Each project object shape:
 *   { id, name, featureCount, completedTasks, totalTasks, completionPct, healthStatus, lastUpdated, warnings }
 */
(function () {
  'use strict';

  // ---------------------------------------------------------------------------
  // Constants
  // ---------------------------------------------------------------------------
  var SKELETON_COUNT = 6;
  var REDUCED_MOTION = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  // ---------------------------------------------------------------------------
  // DOM references
  // ---------------------------------------------------------------------------
  var grid = document.getElementById('project-cards');
  var emptyState = document.getElementById('empty-state');
  var warningContainer = document.getElementById('warning-banners');

  // ---------------------------------------------------------------------------
  // Theme toggle
  // ---------------------------------------------------------------------------
  var themeToggle = document.getElementById('theme-toggle');
  var iconSun = document.querySelector('.icon-sun');
  var iconMoon = document.querySelector('.icon-moon');

  function initTheme() {
    var stored = localStorage.getItem('dashboard-theme');
    if (stored === 'dark') {
      applyTheme('dark');
    }
  }

  function applyTheme(theme) {
    if (theme === 'dark') {
      document.documentElement.setAttribute('data-theme', 'dark');
      if (iconSun) iconSun.style.display = 'none';
      if (iconMoon) iconMoon.style.display = 'block';
      localStorage.setItem('dashboard-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
      if (iconSun) iconSun.style.display = 'block';
      if (iconMoon) iconMoon.style.display = 'none';
      localStorage.setItem('dashboard-theme', 'light');
    }
  }

  if (themeToggle) {
    themeToggle.addEventListener('click', function () {
      var current = document.documentElement.getAttribute('data-theme');
      applyTheme(current === 'dark' ? 'light' : 'dark');
    });
  }

  initTheme();

  // ---------------------------------------------------------------------------
  // Relative time formatter
  // ---------------------------------------------------------------------------
  function relativeTime(isoString) {
    if (!isoString) return '';
    var now = Date.now();
    var then = new Date(isoString).getTime();
    if (isNaN(then)) return '';
    var diff = now - then;
    var seconds = Math.floor(diff / 1000);
    if (seconds < 0) seconds = 0;

    var minutes = Math.floor(seconds / 60);
    var hours = Math.floor(minutes / 60);
    var days = Math.floor(hours / 24);

    if (days > 30) return formatDate(isoString);
    if (days > 0) return 'Updated ' + days + (days === 1 ? ' day ago' : ' days ago');
    if (hours > 0) return 'Updated ' + hours + (hours === 1 ? ' hour ago' : ' hours ago');
    if (minutes > 0) return 'Updated ' + minutes + (minutes === 1 ? ' minute ago' : ' minutes ago');
    return 'Updated just now';
  }

  function formatDate(isoString) {
    var d = new Date(isoString);
    var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    return 'Updated ' + months[d.getMonth()] + ' ' + d.getDate() + ', ' + d.getFullYear();
  }

  // ---------------------------------------------------------------------------
  // Skeleton cards (loading state)
  // ---------------------------------------------------------------------------
  function renderSkeletons() {
    for (var i = 0; i < SKELETON_COUNT; i++) {
      var sk = document.createElement('div');
      sk.className = 'card skeleton skeleton-card';
      sk.setAttribute('aria-hidden', 'true');
      grid.appendChild(sk);
    }
  }

  function clearSkeletons() {
    var skeletons = grid.querySelectorAll('.skeleton');
    for (var i = 0; i < skeletons.length; i++) {
      if (REDUCED_MOTION) {
        skeletons[i].remove();
      } else {
        skeletons[i].classList.add('fade-out');
      }
    }
    // Remove after animation
    if (!REDUCED_MOTION) {
      setTimeout(function () {
        var remaining = grid.querySelectorAll('.skeleton');
        for (var j = 0; j < remaining.length; j++) {
          remaining[j].remove();
        }
      }, 160);
    }
  }

  // ---------------------------------------------------------------------------
  // Warning banners (error state)
  // ---------------------------------------------------------------------------
  function renderWarnings(projects) {
    var hasWarnings = false;
    for (var i = 0; i < projects.length; i++) {
      var p = projects[i];
      if (p.warnings && p.warnings.length > 0) {
        hasWarnings = true;
        for (var w = 0; w < p.warnings.length; w++) {
          warningContainer.appendChild(createWarningBanner(p.name, p.warnings[w]));
        }
      }
    }
    return hasWarnings;
  }

  function createWarningBanner(projectName, message) {
    var banner = document.createElement('div');
    banner.className = 'warning-banner';

    var icon = document.createElement('span');
    icon.className = 'warning-banner-icon';
    icon.setAttribute('aria-hidden', 'true');
    // Alert triangle SVG
    icon.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>';

    var textContainer = document.createElement('div');
    var text = document.createElement('span');
    text.className = 'warning-banner-text';
    text.textContent = 'Path not found: ' + projectName;

    var pathText = document.createElement('div');
    pathText.className = 'warning-banner-path';
    pathText.textContent = message;

    textContainer.appendChild(text);
    textContainer.appendChild(pathText);

    banner.appendChild(icon);
    banner.appendChild(textContainer);

    return banner;
  }

  // ---------------------------------------------------------------------------
  // Health status badge
  // ---------------------------------------------------------------------------
  function createHealthBadge(status) {
    var badge = document.createElement('span');
    badge.className = 'badge health-badge health-' + status;

    var dot = document.createElement('span');
    dot.className = 'health-dot';

    var srText = document.createElement('span');
    srText.className = 'sr-only';
    srText.textContent = status;

    badge.appendChild(dot);
    badge.appendChild(srText);

    return badge;
  }

  // ---------------------------------------------------------------------------
  // Progress bar
  // ---------------------------------------------------------------------------
  function createProgressBar(completionPct) {
    var bar = document.createElement('div');
    bar.className = 'progress-bar';
    bar.setAttribute('role', 'progressbar');
    bar.setAttribute('aria-valuenow', Math.round(completionPct));
    bar.setAttribute('aria-valuemin', '0');
    bar.setAttribute('aria-valuemax', '100');

    var fill = document.createElement('div');
    fill.className = 'progress-bar-fill';
    if (completionPct >= 80) {
      fill.classList.add('high-completion');
    }
    fill.style.width = completionPct + '%';

    bar.appendChild(fill);
    return bar;
  }

  // ---------------------------------------------------------------------------
  // Project card
  // ---------------------------------------------------------------------------
  function createProjectCard(project) {
    var card = document.createElement('a');
    card.href = '/projects/' + project.id;
    card.className = 'card';
    card.setAttribute('role', 'link');
    card.setAttribute('aria-label', 'View project ' + project.name);

    // Card header: name + health badge
    var header = document.createElement('div');
    header.className = 'card-header';

    var title = document.createElement('h4');
    title.className = 'card-title';
    title.textContent = project.name;

    header.appendChild(title);
    header.appendChild(createHealthBadge(project.healthStatus));

    // Stats row
    var stats = document.createElement('div');
    stats.className = 'card-stats';

    var featureStat = document.createElement('span');
    featureStat.className = 'stat-muted';
    featureStat.textContent = project.featureCount + (project.featureCount === 1 ? ' feature' : ' features');

    var divider = document.createElement('span');
    divider.className = 'stat-divider';
    divider.textContent = '|';

    var taskStat = document.createElement('span');
    taskStat.textContent = project.completedTasks + ' / ' + project.totalTasks + ' tasks';

    var pctStat = document.createElement('span');
    pctStat.className = 'stat-muted';
    pctStat.textContent = Math.round(project.completionPct) + '%';

    stats.appendChild(featureStat);
    stats.appendChild(divider);
    stats.appendChild(taskStat);
    stats.appendChild(pctStat);

    // Progress bar
    var progressBar = createProgressBar(project.completionPct);

    // Card footer
    var footer = document.createElement('div');
    footer.className = 'card-footer';
    footer.textContent = relativeTime(project.lastUpdated);

    card.appendChild(header);
    card.appendChild(stats);
    card.appendChild(progressBar);
    card.appendChild(footer);

    return card;
  }

  // ---------------------------------------------------------------------------
  // Main render
  // ---------------------------------------------------------------------------
  function render(projects) {
    // Sort alphabetically by name
    projects.sort(function (a, b) {
      var nameA = (a.name || '').toLowerCase();
      var nameB = (b.name || '').toLowerCase();
      if (nameA < nameB) return -1;
      if (nameA > nameB) return 1;
      return 0;
    });

    // Render warning banners for errored projects
    renderWarnings(projects);

    // Empty state
    if (projects.length === 0) {
      grid.style.display = 'none';
      emptyState.style.display = '';
      if (!REDUCED_MOTION) {
        emptyState.classList.add('fade-in');
      }
      return;
    }

    // Clear skeletons
    clearSkeletons();

    // Render cards
    for (var i = 0; i < projects.length; i++) {
      var cardEl = createProjectCard(projects[i]);
      if (!REDUCED_MOTION) {
        cardEl.classList.add('fade-in');
      }
      grid.appendChild(cardEl);
    }
  }

  // ---------------------------------------------------------------------------
  // Bootstrap
  // ---------------------------------------------------------------------------
  function init() {
    // Show skeletons immediately (before data renders)
    renderSkeletons();

    // Projects data is injected by the Go template
    var projects = window.__PROJECTS_DATA__ || [];
    if (typeof projects === 'string') {
      try {
        projects = JSON.parse(projects);
      } catch (e) {
        projects = [];
      }
    }

    // Use requestAnimationFrame to ensure skeletons paint before cards replace them
    // This prevents layout shift
    requestAnimationFrame(function () {
      render(projects);
    });
  }

  // Run on DOM ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
