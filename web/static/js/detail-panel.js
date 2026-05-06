/**
 * Detail Panel — Agent Task Dashboard
 *
 * Slide-over panel that shows full task details when a task card is clicked.
 * Fetches data from the API, renders metadata, acceptance criteria,
 * execution record sections, and clickable dependency links.
 *
 * Exposes:
 *   window.openDetailPanel(taskId, featureSlug)
 *   window.navigateToTask(taskId, featureSlug)
 */
(function () {
  'use strict';

  // ---------------------------------------------------------------------------
  // Constants
  // ---------------------------------------------------------------------------

  var CLOSE_ANIMATION_MS = 200;
  var HIGHLIGHT_DURATION_MS = 2000;
  var HIGHLIGHT_FADE_MS = 300;
  var SCROLL_DELAY_MS = 100;

  // ---------------------------------------------------------------------------
  // State
  // ---------------------------------------------------------------------------

  var currentTaskId = null;
  var currentFeatureSlug = null;
  var triggerCard = null;      // task card that opened the panel
  var isClosing = false;
  var focusTrapHandler = null;

  // ---------------------------------------------------------------------------
  // DOM references
  // ---------------------------------------------------------------------------

  function getOverlay() {
    return document.getElementById('detail-overlay');
  }

  function getPanel() {
    return document.getElementById('detail-panel');
  }

  function getPanelContent() {
    return document.getElementById('detail-panel-content');
  }

  function getTaskIdEl() {
    return document.getElementById('detail-panel-task-id');
  }

  function getCloseBtn() {
    return document.getElementById('detail-panel-close');
  }

  // ---------------------------------------------------------------------------
  // Open panel
  // ---------------------------------------------------------------------------

  /**
   * Open the detail panel for a given task.
   * @param {string} taskId - The task ID (e.g., "1.1")
   * @param {string} [featureSlug] - The feature slug the task belongs to
   */
  window.openDetailPanel = function (taskId, featureSlug) {
    if (isClosing) return;

    var overlay = getOverlay();
    var panel = getPanel();
    var content = getPanelContent();
    var taskIdEl = getTaskIdEl();

    if (!overlay || !panel || !content || !taskIdEl) return;

    // Track the triggering card for focus return
    triggerCard = document.querySelector(
      '.task-card[data-task-id="' + CSS.escape(taskId) + '"]'
    );

    // If featureSlug not provided, try to derive from the trigger card
    if (!featureSlug && triggerCard) {
      var row = triggerCard.closest('.feature-row');
      if (row) {
        featureSlug = row.dataset.feature || '';
      }
    }

    currentTaskId = taskId;
    currentFeatureSlug = featureSlug || '';

    // Update header
    taskIdEl.textContent = 'Task ' + taskId;

    // Show loading spinner
    content.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;padding:48px 0;">' +
                        '<div class="spinner"></div>' +
                        '</div>';

    // Show panel and overlay
    overlay.style.display = '';
    panel.style.display = '';

    // Remove closing class if present
    overlay.classList.remove('closing');
    panel.classList.remove('closing');

    // Set focus to panel
    panel.focus();

    // Bind close handlers
    bindCloseHandlers();

    // Set up focus trap
    setupFocusTrap(panel);

    // Fetch task detail
    fetchTaskDetail(taskId, featureSlug);
  };

  // ---------------------------------------------------------------------------
  // Fetch task detail from API
  // ---------------------------------------------------------------------------

  function getProjectId() {
    var dataEl = document.getElementById('__INITIAL_DATA__');
    if (!dataEl) return null;
    try {
      var data = JSON.parse(dataEl.textContent);
      return data.projectId || null;
    } catch (e) {
      return null;
    }
  }

  function fetchTaskDetail(taskId, featureSlug) {
    var projectId = getProjectId();
    if (!projectId) {
      renderError('Unable to determine project ID');
      return;
    }

    if (!featureSlug) {
      // Try to find the feature slug from the page data
      featureSlug = findFeatureSlugForTask(taskId);
    }

    if (!featureSlug) {
      renderError('Unable to determine feature for task ' + taskId);
      return;
    }

    currentFeatureSlug = featureSlug;

    var url = '/api/projects/' + encodeURIComponent(projectId) +
              '/features/' + encodeURIComponent(featureSlug) +
              '/tasks/' + encodeURIComponent(taskId);

    fetch(url)
      .then(function (res) {
        if (!res.ok) {
          throw new Error('HTTP ' + res.status);
        }
        return res.json();
      })
      .then(function (data) {
        renderTaskDetail(data);
      })
      .catch(function (err) {
        renderError('Unable to load task details: ' + err.message);
      });
  }

  /**
   * Search the page for which feature contains the given task ID.
   */
  function findFeatureSlugForTask(taskId) {
    var card = document.querySelector(
      '.task-card[data-task-id="' + CSS.escape(taskId) + '"]'
    );
    if (card) {
      var row = card.closest('.feature-row');
      if (row && row.dataset.feature) {
        return row.dataset.feature;
      }
    }
    return null;
  }

  // ---------------------------------------------------------------------------
  // Render task detail content
  // ---------------------------------------------------------------------------

  function renderTaskDetail(data) {
    var content = getPanelContent();
    if (!content) return;

    var html = '';

    // Title
    html += '<div class="detail-panel-title">' + escapeHtml(data.title || '') + '</div>';

    // Badges: status + priority + scope
    html += '<div class="detail-panel-badges">';
    html += renderStatusBadge(data.status);
    html += renderPriorityBadge(data.priority);
    if (data.scope) {
      html += '<span class="badge" style="background:var(--secondary);color:var(--secondary-foreground);">' +
              escapeHtml(data.scope) + '</span>';
    }
    html += '</div>';

    // Metadata section
    html += '<div>';
    html += '<div class="section-header">Details</div>';
    html += '<div class="metadata-table">';
    html += renderMetadataRow('Est. Time', data.estimatedTime || '--');
    html += renderMetadataRow('Breaking', data.breaking ?
      '<span style="color:var(--status-skipped);">Yes &#x26A0;&#xFE0F;</span>' : 'No');
    html += renderMetadataRow('Dependencies', renderDependencyChips(data.dependencies || []));
    html += renderMetadataRow('File Path', data.file ? '<span class="mono">' + escapeHtml(data.file) + '</span>' : '--', true);
    html += renderMetadataRow('Record', data.record ? '<span class="mono">' + escapeHtml(data.record) + '</span>' : '--', true);
    html += '</div>';
    html += '</div>';

    // Acceptance Criteria section
    var criteria = data.acceptanceCriteria || [];
    if (criteria.length > 0) {
      html += '<div>';
      html += '<div class="section-header">Acceptance Criteria</div>';
      html += '<div class="criteria-list">';
      for (var i = 0; i < criteria.length; i++) {
        html += '<div class="criteria-item">';
        html += '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="flex-shrink:0;margin-top:2px;color:var(--muted-foreground);">' +
                '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>' +
                '<polyline points="22 4 12 14.01 9 11.01"/>' +
                '</svg>';
        html += '<span>' + escapeHtml(criteria[i]) + '</span>';
        html += '</div>';
      }
      html += '</div>';
      html += '</div>';
    }

    // Execution Record section
    html += '<div>';
    html += '<div class="section-header">Execution Record</div>';
    if (data.executionRecord) {
      html += renderExecutionRecord(data.executionRecord);
    } else {
      html += '<div class="no-record">';
      html += '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">' +
              '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>' +
              '<polyline points="14 2 14 8 20 8"/>' +
              '</svg>';
      html += '<span>No execution record</span>';
      html += '</div>';
    }
    html += '</div>';

    content.innerHTML = html;

    // Bind dependency chip clicks
    var chips = content.querySelectorAll('.dep-chip[data-task-id]');
    chips.forEach(function (chip) {
      chip.addEventListener('click', function () {
        var depTaskId = chip.dataset.taskId;
        var depFeature = chip.dataset.featureSlug || '';
        navigateToTask(depTaskId, depFeature);
      });
    });
  }

  function renderStatusBadge(status) {
    var label = (status || 'pending').replace(/_/g, ' ');
    var cls = 'status-' + (status || 'pending').replace(/_/g, '-');
    return '<span class="badge ' + cls + '" style="gap:6px;">' +
           '<span class="status-dot"></span>' +
           '<span class="sr-only">' + escapeHtml(label) + '</span>' +
           escapeHtml(label) +
           '</span>';
  }

  function renderPriorityBadge(priority) {
    if (!priority) return '';
    var cls = 'priority-' + priority.toLowerCase();
    return '<span class="priority-badge ' + cls + '">' + escapeHtml(priority) + '</span>';
  }

  function renderMetadataRow(label, valueHtml, isMono) {
    var valueClass = 'metadata-value' + (isMono ? '' : '');
    return '<div class="metadata-row">' +
           '<span class="metadata-label">' + escapeHtml(label) + '</span>' +
           '<span class="' + valueClass + '">' + valueHtml + '</span>' +
           '</div>';
  }

  function renderDependencyChips(deps) {
    if (!deps || deps.length === 0) {
      return '<span style="color:var(--muted-foreground);">None</span>';
    }
    var html = '<div style="display:flex;flex-wrap:wrap;gap:6px;">';
    for (var i = 0; i < deps.length; i++) {
      var depId = deps[i];
      // Try to find which feature this dependency belongs to
      var featureSlug = findFeatureSlugForDep(depId);
      html += '<span class="dep-chip" data-task-id="' + escapeAttr(depId) +
              '" data-feature-slug="' + escapeAttr(featureSlug) +
              '" role="button" tabindex="0" aria-label="Navigate to task ' + escapeAttr(depId) + '">' +
              escapeHtml(depId) +
              '</span>';
    }
    html += '</div>';
    return html;
  }

  /**
   * Find which feature a dependency task belongs to by scanning the page.
   */
  function findFeatureSlugForDep(taskId) {
    var card = document.querySelector(
      '.task-card[data-task-id="' + CSS.escape(taskId) + '"]'
    );
    if (card) {
      var row = card.closest('.feature-row');
      if (row && row.dataset.feature) {
        return row.dataset.feature;
      }
    }
    return '';
  }

  function renderExecutionRecord(record) {
    var html = '<div class="execution-record">';

    if (record.summary) {
      html += '<div style="margin-bottom:12px;">';
      html += '<div style="font-weight:500;margin-bottom:4px;">Summary</div>';
      html += '<div>' + escapeHtml(record.summary) + '</div>';
      html += '</div>';
    }

    if (record.files && record.files.length > 0) {
      html += '<div style="margin-bottom:12px;">';
      html += '<div style="font-weight:500;margin-bottom:4px;">Files</div>';
      html += '<div style="display:flex;flex-direction:column;gap:4px;">';
      for (var i = 0; i < record.files.length; i++) {
        html += '<code style="font-family:var(--font-mono);font-size:13px;background:var(--muted);padding:2px 6px;border-radius:4px;">' +
                escapeHtml(record.files[i]) + '</code>';
      }
      html += '</div>';
      html += '</div>';
    }

    if (record.decisions) {
      html += '<div style="margin-bottom:12px;">';
      html += '<div style="font-weight:500;margin-bottom:4px;">Decisions</div>';
      html += '<div>' + escapeHtml(record.decisions) + '</div>';
      html += '</div>';
    }

    if (record.testResults) {
      html += '<div style="margin-bottom:12px;">';
      html += '<div style="font-weight:500;margin-bottom:4px;">Test Results</div>';
      html += '<div>' + escapeHtml(record.testResults) + '</div>';
      html += '</div>';
    }

    // Fallback: if nothing matched but raw exists, render raw
    if (!record.summary && !record.files && !record.decisions && !record.testResults && record.raw) {
      html += '<pre>' + escapeHtml(record.raw) + '</pre>';
    }

    html += '</div>';
    return html;
  }

  // ---------------------------------------------------------------------------
  // Error rendering
  // ---------------------------------------------------------------------------

  function renderError(message) {
    var content = getPanelContent();
    if (!content) return;

    content.innerHTML =
      '<div class="error-alert">' +
      '<div class="error-alert-title">' + escapeHtml(message) + '</div>' +
      '<button class="btn btn-outline error-alert-action" id="detail-retry-btn">Retry</button>' +
      '</div>';

    var retryBtn = document.getElementById('detail-retry-btn');
    if (retryBtn) {
      retryBtn.addEventListener('click', function () {
        // Show spinner again
        content.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;padding:48px 0;">' +
                            '<div class="spinner"></div>' +
                            '</div>';
        fetchTaskDetail(currentTaskId, currentFeatureSlug);
      });
    }
  }

  // ---------------------------------------------------------------------------
  // Close panel
  // ---------------------------------------------------------------------------

  function closePanel() {
    if (isClosing) return;
    isClosing = true;

    var overlay = getOverlay();
    var panel = getPanel();

    if (!overlay || !panel) {
      isClosing = false;
      return;
    }

    // Add closing animation class
    overlay.classList.add('closing');
    panel.classList.add('closing');

    // After animation, hide elements
    setTimeout(function () {
      overlay.style.display = 'none';
      panel.style.display = 'none';
      overlay.classList.remove('closing');
      panel.classList.remove('closing');

      // Clear content
      var content = getPanelContent();
      if (content) content.innerHTML = '';

      var taskIdEl = getTaskIdEl();
      if (taskIdEl) taskIdEl.textContent = '';

      // Unbind close handlers
      unbindCloseHandlers();
      removeFocusTrap();

      // Return focus to trigger card
      if (triggerCard) {
        triggerCard.focus();
        triggerCard = null;
      }

      currentTaskId = null;
      currentFeatureSlug = null;
      isClosing = false;
    }, CLOSE_ANIMATION_MS);
  }

  // ---------------------------------------------------------------------------
  // Close event handlers
  // ---------------------------------------------------------------------------

  var boundClosePanel = null;
  var boundOverlayClick = null;
  var boundKeydown = null;

  function bindCloseHandlers() {
    // Avoid double-binding
    unbindCloseHandlers();

    boundClosePanel = closePanel;
    boundOverlayClick = function (e) {
      // Only close if clicking directly on the overlay (not children)
      if (e.target === getOverlay()) {
        closePanel();
      }
    };
    boundKeydown = function (e) {
      if (e.key === 'Escape') {
        var panel = getPanel();
        // Only close if panel is visible
        if (panel && panel.style.display !== 'none') {
          e.preventDefault();
          closePanel();
        }
      }
    };

    var closeBtn = getCloseBtn();
    if (closeBtn) {
      closeBtn.addEventListener('click', boundClosePanel);
    }

    var overlay = getOverlay();
    if (overlay) {
      overlay.addEventListener('click', boundOverlayClick);
    }

    document.addEventListener('keydown', boundKeydown);
  }

  function unbindCloseHandlers() {
    var closeBtn = getCloseBtn();
    if (closeBtn && boundClosePanel) {
      closeBtn.removeEventListener('click', boundClosePanel);
    }

    var overlay = getOverlay();
    if (overlay && boundOverlayClick) {
      overlay.removeEventListener('click', boundOverlayClick);
    }

    if (boundKeydown) {
      document.removeEventListener('keydown', boundKeydown);
    }

    boundClosePanel = null;
    boundOverlayClick = null;
    boundKeydown = null;
  }

  // ---------------------------------------------------------------------------
  // Focus trap
  // ---------------------------------------------------------------------------

  function setupFocusTrap(panel) {
    removeFocusTrap();

    focusTrapHandler = function (e) {
      if (e.key !== 'Tab') return;

      var focusable = panel.querySelectorAll(
        'button, [tabindex]:not([tabindex="-1"]), a, input, select, textarea'
      );

      if (focusable.length === 0) return;

      var first = focusable[0];
      var last = focusable[focusable.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first || document.activeElement === panel) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    panel.addEventListener('keydown', focusTrapHandler);
  }

  function removeFocusTrap() {
    var panel = getPanel();
    if (panel && focusTrapHandler) {
      panel.removeEventListener('keydown', focusTrapHandler);
    }
    focusTrapHandler = null;
  }

  // ---------------------------------------------------------------------------
  // Navigate to task (dependency click)
  // ---------------------------------------------------------------------------

  /**
   * Navigate to a task in the swimlane. Closes the panel first, then scrolls
   * to the target task and highlights it.
   * @param {string} taskId - Target task ID
   * @param {string} featureSlug - Feature slug of the target task
   */
  window.navigateToTask = function (taskId, featureSlug) {
    // Close panel first
    var overlay = getOverlay();
    var panel = getPanel();

    if (panel && panel.style.display !== 'none') {
      // Start closing animation
      if (overlay) overlay.classList.add('closing');
      panel.classList.add('closing');

      // At SCROLL_DELAY_MS into close, start scrolling
      setTimeout(function () {
        scrollToTask(taskId, featureSlug);
      }, SCROLL_DELAY_MS);

      // Complete close after animation
      setTimeout(function () {
        if (overlay) {
          overlay.style.display = 'none';
          overlay.classList.remove('closing');
        }
        if (panel) {
          panel.style.display = 'none';
          panel.classList.remove('closing');
        }

        var content = getPanelContent();
        if (content) content.innerHTML = '';

        var taskIdEl = getTaskIdEl();
        if (taskIdEl) taskIdEl.textContent = '';

        unbindCloseHandlers();
        removeFocusTrap();

        currentTaskId = null;
        currentFeatureSlug = null;
        isClosing = false;
      }, CLOSE_ANIMATION_MS);

      isClosing = true;
    } else {
      // Panel not open, just scroll directly
      scrollToTask(taskId, featureSlug);
    }
  };

  function scrollToTask(taskId, featureSlug) {
    var card = document.querySelector(
      '.task-card[data-task-id="' + CSS.escape(taskId) + '"]'
    );

    if (card) {
      // Expand parent row if collapsed
      var row = card.closest('.feature-row');
      if (row && row.classList.contains('collapsed')) {
        var slug = row.dataset.feature;
        if (slug && typeof window.toggleFeatureRow === 'function') {
          window.toggleFeatureRow(slug);
        }
      }

      card.scrollIntoView({ behavior: 'smooth', block: 'center' });

      // Add highlight
      card.classList.add('highlighted');

      setTimeout(function () {
        card.classList.remove('highlighted');
      }, HIGHLIGHT_DURATION_MS);

      card.focus();
    } else if (featureSlug) {
      // Cross-feature: the task might not be visible on this page
      // Check if feature row exists but is collapsed/hidden
      var featureRow = document.querySelector(
        '.feature-row[data-feature="' + CSS.escape(featureSlug) + '"]'
      );

      if (featureRow) {
        // Expand the feature row if collapsed
        if (featureRow.classList.contains('collapsed')) {
          if (typeof window.toggleFeatureRow === 'function') {
            window.toggleFeatureRow(featureSlug);
          }
        }

        // Scroll to the feature row first
        featureRow.scrollIntoView({ behavior: 'smooth', block: 'start' });

        // After a delay, try to find and highlight the task card
        setTimeout(function () {
          var taskCard = featureRow.querySelector(
            '.task-card[data-task-id="' + CSS.escape(taskId) + '"]'
          );
          if (taskCard) {
            taskCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
            taskCard.classList.add('highlighted');
            setTimeout(function () {
              taskCard.classList.remove('highlighted');
            }, HIGHLIGHT_DURATION_MS);
            taskCard.focus();
          }
        }, 300);
      }
    }
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

})();
