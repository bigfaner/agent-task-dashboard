/**
 * Activity Sidebar — Agent Task Dashboard
 *
 * Renders the collapsible activity sidebar showing recent task status
 * change events across all features in the current project.
 *
 * Events are derived from the initial template data (embedded by Go template).
 * The sidebar supports expand/collapse, event click-to-navigate, and
 * auto-collapse when the detail panel opens.
 */
(function () {
  'use strict';

  // ---------------------------------------------------------------------------
  // Constants
  // ---------------------------------------------------------------------------

  var MAX_EVENTS = 50;
  var TITLE_MAX_LENGTH = 40;

  // Status-to-event-type color mapping
  var EVENT_COLORS = {
    claimed: 'var(--status-in-progress)',
    completed: 'var(--status-completed)',
    blocked: 'var(--status-blocked)',
    skipped: 'var(--status-skipped)'
  };

  // ---------------------------------------------------------------------------
  // State
  // ---------------------------------------------------------------------------

  var projectData = null;
  var activityEvents = [];
  var blockedCount = 0;
  var isExpanded = true;
  var wasExpandedBeforePanel = true;

  // DOM references
  var sidebar = null;
  var eventListEl = null;
  var toggleBtn = null;

  // ---------------------------------------------------------------------------
  // Initialization
  // ---------------------------------------------------------------------------

  function init() {
    var dataEl = document.getElementById('__INITIAL_DATA__');
    if (!dataEl) return;

    try {
      projectData = JSON.parse(dataEl.textContent);
    } catch (e) {
      console.error('Activity: failed to parse initial data', e);
      return;
    }

    if (!projectData) return;

    activityEvents = projectData.activityEvents || [];
    blockedCount = projectData.blockedCount || 0;

    sidebar = document.getElementById('activity-sidebar');
    eventListEl = document.getElementById('event-list');
    toggleBtn = document.getElementById('sidebar-toggle-btn');

    if (!sidebar || !eventListEl || !toggleBtn) return;

    render();
    bindEvents();
  }

  // ---------------------------------------------------------------------------
  // Rendering
  // ---------------------------------------------------------------------------

  function render() {
    if (activityEvents.length === 0) {
      renderEmptyState();
    } else {
      renderEvents();
    }
  }

  function renderEmptyState() {
    eventListEl.innerHTML =
      '<div class="activity-empty-state">' +
      '<svg class="activity-empty-icon" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">' +
      '<path d="M22 12h-4l-3 9L9 3l-3 9H2"/>' +
      '</svg>' +
      '<span class="activity-empty-text">No activity yet</span>' +
      '</div>';
  }

  function renderEvents() {
    var html = '';
    var events = activityEvents.slice(0, MAX_EVENTS);

    for (var i = 0; i < events.length; i++) {
      html += renderEventItem(events[i]);
    }

    eventListEl.innerHTML = html;
  }

  function renderEventItem(event) {
    var taskId = escapeHtml(event.taskId || '');
    var title = escapeHtml(truncateText(event.taskTitle || '', TITLE_MAX_LENGTH));
    var feature = escapeHtml(event.feature || '');
    var eventType = escapeHtml(event.eventType || '');
    var timestamp = formatTimestamp(event.timestamp);
    var colorClass = 'event-type-' + (event.eventType || '');

    return (
      '<div class="event-item" data-task-id="' + escapeAttr(event.taskId || '') +
      '" data-feature="' + escapeAttr(event.feature || '') + '">' +
      '<div class="event-timestamp">' + timestamp + '</div>' +
      '<div class="event-task-line">' +
      '<span class="status-dot" style="background:' + (EVENT_COLORS[event.eventType] || 'var(--status-pending)') + ';"></span>' +
      '<span class="event-task-id">' + taskId + '</span>' +
      '<span class="event-title">' + title + '</span>' +
      '</div>' +
      '<div class="event-type ' + colorClass + '">' + eventType + '</div>' +
      '<div class="event-feature">' + feature + '</div>' +
      '</div>'
    );
  }

  // ---------------------------------------------------------------------------
  // Collapsed state rendering
  // ---------------------------------------------------------------------------

  function renderCollapsedContent() {
    var existing = sidebar.querySelector('.sidebar-collapsed-content');
    if (existing) existing.remove();

    var html = '<button class="sidebar-toggle-expand" id="sidebar-expand-btn" aria-label="Expand activity sidebar">' +
      '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>' +
      '</button>';

    if (blockedCount > 0) {
      html += '<div class="blocked-badge">' + blockedCount + '</div>';
    }

    var div = document.createElement('div');
    div.className = 'sidebar-collapsed-content';
    div.innerHTML = html;
    sidebar.appendChild(div);

    // Bind expand button
    var expandBtn = div.querySelector('#sidebar-expand-btn');
    if (expandBtn) {
      expandBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        expand();
      });
    }
  }

  function removeCollapsedContent() {
    var el = sidebar.querySelector('.sidebar-collapsed-content');
    if (el) el.remove();
  }

  // ---------------------------------------------------------------------------
  // Expand / Collapse
  // ---------------------------------------------------------------------------

  function collapse() {
    if (!isExpanded) return;
    isExpanded = false;

    sidebar.classList.remove('expanded');
    sidebar.classList.add('collapsed');

    // Hide expanded content, show collapsed
    var header = sidebar.querySelector('.sidebar-header');
    if (header) header.style.display = 'none';
    eventListEl.style.display = 'none';

    renderCollapsedContent();
  }

  function expand() {
    if (isExpanded) return;
    isExpanded = true;

    sidebar.classList.remove('collapsed');
    sidebar.classList.add('expanded');

    // Show expanded content, hide collapsed
    var header = sidebar.querySelector('.sidebar-header');
    if (header) header.style.display = '';
    eventListEl.style.display = '';

    removeCollapsedContent();
  }

  // ---------------------------------------------------------------------------
  // Event binding
  // ---------------------------------------------------------------------------

  function bindEvents() {
    // Collapse button in header
    toggleBtn.addEventListener('click', function () {
      collapse();
    });

    // Event item clicks
    eventListEl.addEventListener('click', function (e) {
      var item = e.target.closest('.event-item');
      if (!item) return;

      var taskId = item.dataset.taskId;
      if (!taskId) return;

      navigateToTask(taskId);
    });

    // Listen for detail panel open/close to auto-collapse/restore sidebar
    bindDetailPanelIntegration();
  }

  function navigateToTask(taskId) {
    if (typeof window.highlightTaskCard === 'function') {
      window.highlightTaskCard(taskId);
    }
  }

  // ---------------------------------------------------------------------------
  // Detail panel integration
  // ---------------------------------------------------------------------------

  function bindDetailPanelIntegration() {
    // Watch for detail panel visibility changes via MutationObserver
    var panel = document.getElementById('detail-panel');
    if (!panel) return;

    var observer = new MutationObserver(function (mutations) {
      for (var i = 0; i < mutations.length; i++) {
        var mutation = mutations[i];
        if (mutation.attributeName === 'style') {
          var isVisible = panel.style.display !== 'none';
          if (isVisible && isExpanded) {
            wasExpandedBeforePanel = true;
            collapse();
          } else if (!isVisible && wasExpandedBeforePanel) {
            expand();
          }
        }
      }
    });

    observer.observe(panel, { attributes: true, attributeFilter: ['style'] });
  }

  // ---------------------------------------------------------------------------
  // Public API
  // ---------------------------------------------------------------------------

  // Expose for external callers (e.g., detail-panel.js, swimlane.js)
  window.collapseActivitySidebar = function () {
    collapse();
  };

  window.expandActivitySidebar = function () {
    expand();
  };

  window.getActivityState = function () {
    return {
      isExpanded: isExpanded,
      blockedCount: blockedCount,
      eventCount: activityEvents.length
    };
  };

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

  function truncateText(text, maxLen) {
    if (text.length <= maxLen) return text;
    return text.substring(0, maxLen - 3) + '...';
  }

  function formatTimestamp(ts) {
    if (!ts) return '';

    var date = new Date(ts);
    if (isNaN(date.getTime())) return '';

    var now = new Date();
    var diffMs = now.getTime() - date.getTime();
    var diffSec = Math.floor(diffMs / 1000);
    var diffMin = Math.floor(diffSec / 60);
    var diffHour = Math.floor(diffMin / 60);
    var diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return 'just now';
    if (diffMin < 60) return diffMin + 'm ago';
    if (diffHour < 24) return diffHour + 'h ago';
    if (diffDay < 7) return diffDay + 'd ago';

    // Format as HH:mm
    var hours = date.getHours();
    var minutes = date.getMinutes();
    return (hours < 10 ? '0' : '') + hours + ':' + (minutes < 10 ? '0' : '') + minutes;
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
