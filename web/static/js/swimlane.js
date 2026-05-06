/**
 * Swimlane DAG Renderer — Agent Task Dashboard
 *
 * Renders the feature-per-row swimlane board with phase columns,
 * task cards with status colors, and dependency arrows via SVG.
 *
 * Data is injected by Go template as window.__PROJECT_DATA__ JSON.
 * Uses dagre for graph layout of dependency arrows.
 */
(function () {
  'use strict';

  // ---------------------------------------------------------------------------
  // Constants
  // ---------------------------------------------------------------------------

  var PHASE_COLUMNS = ['Phase 1', 'Phase 2', 'Phase 3+', 'Testing', 'Other'];
  var CARD_MIN_WIDTH = 140;
  var CARD_PADDING_X = 12;
  var CARD_PADDING_Y = 12;
  var ARROW_DENSE_THRESHOLD = 50;
  var CLICK_DEBOUNCE_MS = 300;

  // ---------------------------------------------------------------------------
  // State
  // ---------------------------------------------------------------------------

  var projectData = null;     // { projectId, projectName, features[] }
  var featureCache = {};      // slug -> { tasks, phases, dependencies }
  var expandedRows = {};      // slug -> boolean
  var statusFilters = [];     // active status filter values
  var priorityFilters = [];   // active priority filter values
  var lastClickTime = 0;

  // ---------------------------------------------------------------------------
  // Initialization
  // ---------------------------------------------------------------------------

  function init() {
    var dataEl = document.getElementById('__INITIAL_DATA__');
    if (!dataEl) {
      console.error('Swimlane: __INITIAL_DATA__ script element not found');
      return;
    }

    try {
      projectData = JSON.parse(dataEl.textContent);
    } catch (e) {
      console.error('Swimlane: failed to parse initial data', e);
      return;
    }

    if (!projectData || !projectData.projectId) {
      console.error('Swimlane: invalid project data');
      return;
    }

    renderPage();
    bindGlobalEvents();
  }

  // ---------------------------------------------------------------------------
  // Data fetching
  // ---------------------------------------------------------------------------

  function fetchFeatureDetail(slug) {
    if (featureCache[slug]) {
      return Promise.resolve(featureCache[slug]);
    }

    var url = '/api/projects/' + encodeURIComponent(projectData.projectId) +
              '/features/' + encodeURIComponent(slug);

    return fetch(url)
      .then(function (res) {
        if (!res.ok) throw new Error('Failed to load feature: ' + slug);
        return res.json();
      })
      .then(function (data) {
        featureCache[slug] = data;
        return data;
      });
  }

  function fetchAllFeatures() {
    if (!projectData.features || projectData.features.length === 0) {
      return Promise.resolve([]);
    }

    return Promise.all(
      projectData.features.map(function (f) {
        return fetchFeatureDetail(f.slug).catch(function () {
          return null;
        });
      })
    ).then(function (results) {
      return results.filter(function (r) { return r !== null; });
    });
  }

  // ---------------------------------------------------------------------------
  // Page rendering
  // ---------------------------------------------------------------------------

  function renderPage() {
    var container = document.getElementById('swimlane-container');
    if (!container) return;

    // Show skeleton loading
    container.innerHTML = renderSkeletons();

    fetchAllFeatures().then(function (features) {
      if (features.length === 0) {
        container.innerHTML = renderEmptyState();
        return;
      }

      // Determine phase columns to show
      var phases = derivePhaseColumns(features);

      // Build the swimlane HTML
      var html = '';

      // Phase header row
      html += renderPhaseHeaders(phases);

      // Feature rows (sorted)
      var sortedFeatures = sortFeatures(features);
      for (var i = 0; i < sortedFeatures.length; i++) {
        html += renderFeatureRow(sortedFeatures[i], phases);
      }

      // SVG overlay for dependency arrows
      html += '<svg class="svg-overlay" id="svg-overlay"><defs>' +
              '<marker id="arrowhead" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">' +
              '<polygon points="0 0, 10 3.5, 0 7" fill="var(--muted-foreground)" />' +
              '</marker>' +
              '<marker id="arrowhead-dashed" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">' +
              '<polygon points="0 0, 10 3.5, 0 7" fill="var(--muted-foreground)" />' +
              '</marker>' +
              '</defs></svg>';

      container.innerHTML = html;

      // Initialize row states
      initRowStates(sortedFeatures);

      // Render dependency arrows after DOM is ready
      requestAnimationFrame(function () {
        renderDependencyArrows(features);
      });
    });
  }

  function renderSkeletons() {
    var html = '';
    for (var i = 0; i < 8; i++) {
      html += '<div class="skeleton skeleton-row"></div>';
    }
    return html;
  }

  function renderEmptyState() {
    return '<div class="empty-state">' +
           '<div class="empty-state-icon">' +
           '<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">' +
           '<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>' +
           '</svg></div>' +
           '<div class="empty-state-title">No features found in this project</div>' +
           '<div class="empty-state-description">Check that docs/features/ contains valid task data</div>' +
           '</div>';
  }

  // ---------------------------------------------------------------------------
  // Phase columns
  // ---------------------------------------------------------------------------

  function derivePhaseColumns(features) {
    var hasOther = false;
    for (var i = 0; i < features.length; i++) {
      var tasks = features[i].tasks;
      if (!tasks) continue;
      var keys = Object.keys(tasks);
      for (var j = 0; j < keys.length; j++) {
        var phase = tasks[keys[j]].phase;
        if (phase === 0 || phase > 3) {
          // Phase 0 is "Testing" or unrecognized
          var id = tasks[keys[j]].id;
          if (id && id.charAt(0) !== 'T') {
            hasOther = true;
          }
        }
      }
    }

    var columns = PHASE_COLUMNS.slice();
    if (!hasOther) {
      columns = columns.slice(0, 4); // Remove "Other"
    }
    return columns;
  }

  function getPhaseColumnIndex(task, columns) {
    var phase = task.phase;
    var id = task.id || '';

    // Testing tasks: T-* prefix
    if (id.charAt(0) === 'T' && id.charAt(1) === '-') {
      var idx = columns.indexOf('Testing');
      return idx >= 0 ? idx : columns.length - 1;
    }

    // Phase 1
    if (phase === 1) {
      var idx = columns.indexOf('Phase 1');
      return idx >= 0 ? idx : 0;
    }

    // Phase 2
    if (phase === 2) {
      var idx = columns.indexOf('Phase 2');
      return idx >= 0 ? idx : 1;
    }

    // Phase 3+
    if (phase >= 3) {
      var idx = columns.indexOf('Phase 3+');
      return idx >= 0 ? idx : 2;
    }

    // Unrecognized phase -> Other
    var idx = columns.indexOf('Other');
    return idx >= 0 ? idx : columns.length - 1;
  }

  // ---------------------------------------------------------------------------
  // Phase headers
  // ---------------------------------------------------------------------------

  function renderPhaseHeaders(phases) {
    var html = '<div class="phase-headers" style="grid-template-columns: repeat(' + phases.length + ', 1fr);">';
    for (var i = 0; i < phases.length; i++) {
      html += '<div class="phase-header">' + escapeHtml(phases[i]) + '</div>';
    }
    html += '</div>';
    return html;
  }

  // ---------------------------------------------------------------------------
  // Feature sorting
  // ---------------------------------------------------------------------------

  function sortFeatures(features) {
    return features.slice().sort(function (a, b) {
      // 1. Features with blocked tasks come first
      var aBlocked = hasBlockedTasks(a) ? 0 : 1;
      var bBlocked = hasBlockedTasks(b) ? 0 : 1;
      if (aBlocked !== bBlocked) return aBlocked - bBlocked;

      // 2. Completion % ascending (most incomplete first)
      var aPct = a.completionPct || 0;
      var bPct = b.completionPct || 0;
      if (aPct !== bPct) return aPct - bPct;

      // 3. Alphabetical by slug
      var aSlug = a.slug || '';
      var bSlug = b.slug || '';
      return aSlug.localeCompare(bSlug);
    });
  }

  function hasBlockedTasks(feature) {
    var tasks = feature.tasks;
    if (!tasks) return false;
    var keys = Object.keys(tasks);
    for (var i = 0; i < keys.length; i++) {
      if (tasks[keys[i]].status === 'blocked') return true;
    }
    return false;
  }

  // ---------------------------------------------------------------------------
  // Feature row rendering
  // ---------------------------------------------------------------------------

  function renderFeatureRow(feature, phases) {
    var slug = feature.slug || '';
    var completionPct = Math.round(feature.completionPct || 0);
    var completed = feature.completedTasks || 0;
    var total = feature.totalTasks || 0;
    var isBlocked = hasBlockedTasks(feature);

    var html = '<div class="feature-row expanded" data-feature="' + escapeAttr(slug) +
               '" data-has-blocked="' + (isBlocked ? 'true' : 'false') + '">';

    // Row header
    html += '<div class="feature-row-header" data-toggle="' + escapeAttr(slug) + '">';
    html += '<svg class="chevron" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>';
    html += '<span class="feature-slug">' + escapeHtml(slug) + '</span>';
    html += '<span class="completion-badge">' + completionPct + '% (' + completed + '/' + total + ')</span>';
    html += '</div>';

    // Collapsed inline progress
    html += '<div class="collapsed-progress" style="display:none;">';
    html += '<div class="collapsed-progress-bar"><div class="collapsed-progress-fill" style="width:' + completionPct + '%;"></div></div>';
    html += '<span class="collapsed-progress-text">' + completionPct + '%</span>';
    html += '</div>';

    // Task area with phase columns
    html += '<div class="feature-row-content" style="grid-template-columns: repeat(' + phases.length + ', 1fr);">';

    // Render task cards into the correct phase column
    var columns = buildColumnContents(feature, phases);
    for (var c = 0; c < phases.length; c++) {
      html += '<div class="phase-cell">';
      if (columns[c]) {
        for (var t = 0; t < columns[c].length; t++) {
          html += renderTaskCard(columns[c][t]);
        }
      }
      html += '</div>';
    }

    html += '</div>'; // feature-row-content
    html += '</div>'; // feature-row

    return html;
  }

  function buildColumnContents(feature, phases) {
    var columns = [];
    for (var i = 0; i < phases.length; i++) {
      columns.push([]);
    }

    var tasks = feature.tasks;
    if (!tasks) return columns;

    var keys = Object.keys(tasks);
    for (var i = 0; i < keys.length; i++) {
      var task = tasks[keys[i]];
      var colIdx = getPhaseColumnIndex(task, phases);
      columns[colIdx].push(task);
    }

    // Sort tasks within each column by ID
    for (var i = 0; i < columns.length; i++) {
      columns[i].sort(function (a, b) {
        return (a.id || '').localeCompare(b.id || '', undefined, { numeric: true });
      });
    }

    return columns;
  }

  // ---------------------------------------------------------------------------
  // Task card rendering
  // ---------------------------------------------------------------------------

  function renderTaskCard(task) {
    var id = task.id || '';
    var title = truncateTitle(task.title || '', 30);
    var status = task.status || 'pending';
    var priority = task.priority || '';
    var statusClass = 'status-' + status.replace('_', '-');

    var html = '<div class="task-card" data-task-id="' + escapeAttr(id) +
               '" data-status="' + escapeAttr(status) +
               '" data-priority="' + escapeAttr(priority) +
               '" data-feature="' + escapeAttr(task._feature || '') +
               '" tabindex="0">';

    // Header: ID + priority badge
    html += '<div class="task-card-header">';
    html += '<span class="task-card-id">' + escapeHtml(id) + '</span>';
    html += '<span class="priority-badge priority-' + priority.toLowerCase() + '">' + escapeHtml(priority) + '</span>';
    html += '</div>';

    // Title
    html += '<div class="task-card-title">' + escapeHtml(title) + '</div>';

    // Status dot + text
    html += '<div class="task-card-status ' + statusClass + '">';
    html += '<span class="status-dot"></span>';
    html += '<span>' + escapeHtml(status.replace('_', ' ')) + '</span>';
    html += '</div>';

    html += '</div>';
    return html;
  }

  // ---------------------------------------------------------------------------
  // Row state management
  // ---------------------------------------------------------------------------

  function initRowStates(features) {
    for (var i = 0; i < features.length; i++) {
      expandedRows[features[i].slug] = true;
    }
  }

  function toggleRow(slug) {
    var row = document.querySelector('.feature-row[data-feature="' + CSS.escape(slug) + '"]');
    if (!row) return;

    var isExpanded = row.classList.contains('expanded');
    if (isExpanded) {
      row.classList.remove('expanded');
      row.classList.add('collapsed');
      expandedRows[slug] = false;

      // Show collapsed progress, hide content
      var collapsedProgress = row.querySelector('.collapsed-progress');
      if (collapsedProgress) collapsedProgress.style.display = '';
    } else {
      row.classList.remove('collapsed');
      row.classList.add('expanded');
      expandedRows[slug] = true;

      // Hide collapsed progress, show content
      var collapsedProgress = row.querySelector('.collapsed-progress');
      if (collapsedProgress) collapsedProgress.style.display = 'none';
    }

    // Re-render arrows (collapsed rows should hide their arrows)
    requestAnimationFrame(function () {
      renderDependencyArrows(getCurrentFeatures());
    });
  }

  function getCurrentFeatures() {
    return Object.keys(featureCache).map(function (slug) {
      return featureCache[slug];
    }).filter(function (f) { return f !== null; });
  }

  // ---------------------------------------------------------------------------
  // Dependency arrows
  // ---------------------------------------------------------------------------

  function renderDependencyArrows(features) {
    var svg = document.getElementById('svg-overlay');
    if (!svg) return;

    // Clear existing paths (keep <defs>)
    var defs = svg.querySelector('defs');
    svg.innerHTML = '';
    if (defs) svg.appendChild(defs);

    if (!features || features.length === 0) return;

    // Build task position map
    var taskPositions = {};
    var taskCards = document.querySelectorAll('.task-card');
    taskCards.forEach(function (card) {
      var taskId = card.dataset.taskId;
      var feature = card.dataset.feature;
      // Check if the card is in a collapsed row
      var row = card.closest('.feature-row');
      if (row && row.classList.contains('collapsed')) return;

      var rect = card.getBoundingClientRect();
      var containerRect = svg.parentElement.getBoundingClientRect();

      taskPositions[taskId] = {
        x: rect.left + rect.width / 2 - containerRect.left,
        y: rect.top + rect.height / 2 - containerRect.top,
        right: rect.right - containerRect.left,
        left: rect.left - containerRect.left,
        top: rect.top - containerRect.top,
        bottom: rect.bottom - containerRect.top,
        feature: feature
      };
    });

    // Count cross-feature arrows for dense graph simplification
    var crossFeatureEdges = [];
    var withinEdges = [];

    for (var i = 0; i < features.length; i++) {
      var feature = features[i];
      if (!feature.tasks) continue;
      var tasks = feature.tasks;
      var keys = Object.keys(tasks);

      for (var j = 0; j < keys.length; j++) {
        var task = tasks[keys[j]];
        var deps = task.dependencies || [];

        for (var k = 0; k < deps.length; k++) {
          var depId = deps[k];
          var sourcePos = taskPositions[task.id];
          var targetPos = taskPositions[depId];

          if (!sourcePos || !targetPos) continue;

          // Determine if cross-feature
          var isCross = sourcePos.feature !== targetPos.feature;

          if (isCross) {
            crossFeatureEdges.push({ source: sourcePos, target: targetPos, sourceId: task.id, targetId: depId });
          } else {
            withinEdges.push({ source: sourcePos, target: targetPos, sourceId: task.id, targetId: depId });
          }
        }
      }
    }

    // Dense graph simplification: collapse cross-feature arrows if too many
    var renderCrossFeature = crossFeatureEdges.length <= ARROW_DENSE_THRESHOLD;

    if (!renderCrossFeature) {
      // Show summary badges on feature row headers
      var crossByFeature = {};
      for (var i = 0; i < crossFeatureEdges.length; i++) {
        var edge = crossFeatureEdges[i];
        var featSlug = edge.source.feature;
        if (!crossByFeature[featSlug]) crossByFeature[featSlug] = 0;
        crossByFeature[featSlug]++;
      }
      // Add dep count badges to row headers
      var slugs = Object.keys(crossByFeature);
      for (var i = 0; i < slugs.length; i++) {
        var row = document.querySelector('.feature-row[data-feature="' + CSS.escape(slugs[i]) + '"] .feature-row-header');
        if (row) {
          var existing = row.querySelector('.dep-count-badge');
          if (!existing) {
            var badge = document.createElement('span');
            badge.className = 'dep-count-badge';
            badge.textContent = crossByFeature[slugs[i]] + ' deps';
            row.appendChild(badge);
          }
        }
      }
    }

    // Render within-feature arrows
    for (var i = 0; i < withinEdges.length; i++) {
      renderArrow(svg, withinEdges[i].source, withinEdges[i].target, false,
        isBlockedEdge(withinEdges[i].sourceId, withinEdges[i].targetId, features));
    }

    // Render cross-feature arrows (if not simplified)
    if (renderCrossFeature) {
      for (var i = 0; i < crossFeatureEdges.length; i++) {
        renderArrow(svg, crossFeatureEdges[i].source, crossFeatureEdges[i].target, true, false);
      }
    }

    // Update SVG size to match container
    var container = svg.parentElement;
    svg.setAttribute('width', container.scrollWidth);
    svg.setAttribute('height', container.scrollHeight);
  }

  function renderArrow(svg, source, target, isCrossFeature, isBlocked) {
    var path = document.createElementNS('http://www.w3.org/2000/svg', 'path');

    // Calculate path from right edge of source to left edge of target
    var x1 = source.right;
    var y1 = source.y;
    var x2 = target.left;
    var y2 = target.y;

    // If target is to the left of source, use bottom/top edges
    if (x2 <= x1) {
      x1 = source.x;
      y1 = source.bottom;
      x2 = target.x;
      y2 = target.top;
    }

    // Cubic bezier curve
    var midX = (x1 + x2) / 2;
    var d = 'M ' + x1 + ' ' + y1 +
            ' C ' + midX + ' ' + y1 + ', ' + midX + ' ' + y2 + ', ' + x2 + ' ' + y2;

    path.setAttribute('d', d);

    if (isBlocked) {
      path.setAttribute('class', 'edge-blocked');
      path.setAttribute('marker-end', 'url(#arrowhead)');
    } else if (isCrossFeature) {
      path.setAttribute('class', 'edge-cross-feature');
      path.setAttribute('marker-end', 'url(#arrowhead-dashed)');
    } else {
      path.setAttribute('class', 'edge-within');
      path.setAttribute('marker-end', 'url(#arrowhead)');
    }

    svg.appendChild(path);
  }

  function isBlockedEdge(sourceId, targetId, features) {
    // Check if either source or target task is blocked
    for (var i = 0; i < features.length; i++) {
      var tasks = features[i].tasks;
      if (!tasks) continue;
      var keys = Object.keys(tasks);
      for (var j = 0; j < keys.length; j++) {
        var t = tasks[keys[j]];
        if ((t.id === sourceId || t.id === targetId) && t.status === 'blocked') {
          return true;
        }
      }
    }
    return false;
  }

  // ---------------------------------------------------------------------------
  // Filter controls
  // ---------------------------------------------------------------------------

  function applyFilters() {
    var rows = document.querySelectorAll('.feature-row');
    rows.forEach(function (row) {
      if (statusFilters.length === 0 && priorityFilters.length === 0) {
        row.classList.remove('hidden');
        return;
      }

      var cards = row.querySelectorAll('.task-card');
      var match = false;
      cards.forEach(function (card) {
        var cardStatus = card.dataset.status;
        var cardPriority = card.dataset.priority;
        var statusOk = statusFilters.length === 0 || statusFilters.indexOf(cardStatus) >= 0;
        var priorityOk = priorityFilters.length === 0 || priorityFilters.indexOf(cardPriority) >= 0;
        if (statusOk && priorityOk) match = true;
      });

      if (match) {
        row.classList.remove('hidden');
      } else {
        row.classList.add('hidden');
      }
    });
  }

  function updateFilterCount(btn, count) {
    var badge = btn.querySelector('.filter-badge-count');
    if (!badge) return;
    if (count > 0) {
      badge.textContent = count;
      badge.style.display = '';
    } else {
      badge.style.display = 'none';
    }
  }

  // ---------------------------------------------------------------------------
  // Task card click handler
  // ---------------------------------------------------------------------------

  function handleTaskCardClick(e) {
    var now = Date.now();
    if (now - lastClickTime < CLICK_DEBOUNCE_MS) return;
    lastClickTime = now;

    var card = e.target.closest('.task-card');
    if (!card) return;

    var taskId = card.dataset.taskId;
    if (!taskId) return;

    // Call openDetailPanel if available (from detail-panel.js)
    if (typeof window.openDetailPanel === 'function') {
      window.openDetailPanel(taskId);
    }
  }

  // ---------------------------------------------------------------------------
  // Highlight task card (called from detail-panel.js or activity.js)
  // ---------------------------------------------------------------------------

  window.highlightTaskCard = function (taskId) {
    var card = document.querySelector('.task-card[data-task-id="' + CSS.escape(taskId) + '"]');
    if (!card) return;

    // Expand parent row if collapsed
    var row = card.closest('.feature-row');
    if (row && row.classList.contains('collapsed')) {
      var slug = row.dataset.feature;
      toggleRow(slug);
    }

    card.scrollIntoView({ behavior: 'smooth', block: 'center' });
    card.classList.add('highlighted');

    setTimeout(function () {
      card.classList.remove('highlighted');
    }, 2000);
  };

  // ---------------------------------------------------------------------------
  // Refresh swimlane (called externally)
  // ---------------------------------------------------------------------------

  window.refreshSwimlane = function () {
    featureCache = {};
    renderPage();
  };

  // ---------------------------------------------------------------------------
  // Event binding
  // ---------------------------------------------------------------------------

  function bindGlobalEvents() {
    var container = document.getElementById('swimlane-container');
    if (!container) return;

    // Delegate: task card clicks
    container.addEventListener('click', function (e) {
      var card = e.target.closest('.task-card');
      if (card) {
        handleTaskCardClick(e);
        return;
      }

      // Delegate: row header toggle
      var header = e.target.closest('.feature-row-header');
      if (header) {
        var slug = header.dataset.toggle;
        if (slug) toggleRow(slug);
        return;
      }
    });

    // Keyboard support for task cards
    container.addEventListener('keydown', function (e) {
      if (e.key === 'Enter' || e.key === ' ') {
        var card = e.target.closest('.task-card');
        if (card) {
          e.preventDefault();
          handleTaskCardClick({ target: card });
        }
      }
    });

    // Bind filter controls
    bindFilterControls();

    // Re-render arrows on scroll/resize
    var resizeTimer;
    window.addEventListener('resize', function () {
      clearTimeout(resizeTimer);
      resizeTimer = setTimeout(function () {
        renderDependencyArrows(getCurrentFeatures());
      }, 150);
    });

    container.addEventListener('scroll', function () {
      // Arrows are position:absolute within the container, so scroll doesn't break them
      // But we may need to update for virtual scrolling in future
    });
  }

  function bindFilterControls() {
    // Status filter dropdown
    var statusBtn = document.getElementById('status-filter-btn');
    var statusDropdown = document.getElementById('status-filter-dropdown');
    if (statusBtn && statusDropdown) {
      statusBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        closeAllDropdowns();
        statusDropdown.style.display = statusDropdown.style.display === 'none' ? '' : 'none';
      });

      statusDropdown.addEventListener('change', function () {
        statusFilters = getCheckedValues('statusFilter');
        updateFilterCount(statusBtn, statusFilters.length);
        applyFilters();
      });
    }

    // Priority filter dropdown
    var priorityBtn = document.getElementById('priority-filter-btn');
    var priorityDropdown = document.getElementById('priority-filter-dropdown');
    if (priorityBtn && priorityDropdown) {
      priorityBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        closeAllDropdowns();
        priorityDropdown.style.display = priorityDropdown.style.display === 'none' ? '' : 'none';
      });

      priorityDropdown.addEventListener('change', function () {
        priorityFilters = getCheckedValues('priorityFilter');
        updateFilterCount(priorityBtn, priorityFilters.length);
        applyFilters();
      });
    }

    // Close dropdowns on outside click
    document.addEventListener('click', function () {
      closeAllDropdowns();
    });
  }

  function closeAllDropdowns() {
    var dropdowns = document.querySelectorAll('.filter-dropdown');
    dropdowns.forEach(function (dd) {
      dd.style.display = 'none';
    });
  }

  function getCheckedValues(name) {
    var checks = document.querySelectorAll('input[name="' + name + '"]:checked');
    return Array.from(checks).map(function (c) { return c.value; });
  }

  // ---------------------------------------------------------------------------
  // Utilities
  // ---------------------------------------------------------------------------

  function escapeHtml(str) {
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function escapeAttr(str) {
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  function truncateTitle(title, maxLen) {
    if (title.length <= maxLen) return title;
    return title.substring(0, maxLen - 3) + '...';
  }

  // ---------------------------------------------------------------------------
  // Boot
  // ---------------------------------------------------------------------------

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

})();
