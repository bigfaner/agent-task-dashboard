import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../../helpers.js';

test.describe('API E2E Tests — agent-task-dashboard', () => {

  // ── GET /api/projects ────────────────────────────────────────────

  // Traceability: TC-028 → Story 6 / AC-1, AC-2 + Spec 5.5
  test('TC-028: GET /api/projects returns all configured projects', async () => {
    const res = await curl('GET', `${apiBaseUrl()}/api/projects`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(Array.isArray(data.projects)).toBe(true);
    expect(data.meta).toBeDefined();
    expect(data.meta.lastUpdated).toBeTruthy();
    // Each project should have required fields
    const projects = data.projects;
    for (const p of projects) {
      expect(p.id).toBeTruthy();
      expect(p.name).toBeTruthy();
      expect(typeof p.featureCount).toBe('number');
      expect(typeof p.completedTasks).toBe('number');
      expect(typeof p.totalTasks).toBe('number');
      expect(typeof p.completionPct).toBe('number');
      expect(p.lastUpdated).toBeTruthy();
      expect(p.healthStatus).toBeTruthy();
    }
  });

  // ── GET /api/projects/:id ────────────────────────────────────────

  // Traceability: TC-029 → Spec 5.5 Endpoints
  test('TC-029: GET /api/projects/:id returns single project with features', async () => {
    // First get the list to find a valid project ID
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    expect(listRes.status).toBe(200);
    const listData = JSON.parse(listRes.body);
    expect(listData.projects.length).toBeGreaterThanOrEqual(1);
    const projectId = listData.projects[0].id;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.id).toBe(projectId);
    expect(data.name).toBeTruthy();
    expect(Array.isArray(data.features)).toBe(true);
    expect(data.meta).toBeDefined();
    expect(data.meta.lastUpdated).toBeTruthy();
  });

  // ── GET /api/projects/:id/features ───────────────────────────────

  // Traceability: TC-030 → Story 6 / AC-3 + Spec 5.5 Endpoints
  test('TC-030: GET /api/projects/:id/features returns feature list', async () => {
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(Array.isArray(data.features)).toBe(true);
    expect(data.meta).toBeDefined();
    // Each feature should have slug and status
    for (const f of data.features) {
      expect(f.slug).toBeTruthy();
      expect(f.status).toBeTruthy();
    }
  });

  // ── GET /api/projects/:id/features/:slug/tasks ───────────────────

  // Traceability: TC-031 → Story 6 / AC-3, AC-4 + Spec 5.5 Endpoints
  test('TC-031: GET /api/projects/:id/features/:slug/tasks returns task list', async () => {
    // Get a project with features
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    // Get features
    const featRes = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features`);
    const featData = JSON.parse(featRes.body);
    if (featData.features.length === 0) {
      // No features to test; skip gracefully
      return;
    }
    const slug = featData.features[0].slug;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features/${encodeURIComponent(slug)}/tasks`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(Array.isArray(data.tasks)).toBe(true);
    expect(data.meta).toBeDefined();
    // Each task should have required metadata fields
    for (const t of data.tasks) {
      expect(t.id).toBeTruthy();
      expect(t.title).toBeTruthy();
      expect(t.status).toBeTruthy();
      expect(t.priority).toBeTruthy();
    }
  });

  // ── Performance ──────────────────────────────────────────────────

  // Traceability: TC-032 → Story 6 / AC-4 + Spec Performance Requirements
  test('TC-032: API response time is under 200ms for single project', async () => {
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    const start = Date.now();
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}`);
    const elapsed = Date.now() - start;
    expect(res.status).toBe(200);
    expect(elapsed).toBeLessThan(200);
  });

  // ── GET /api/projects/:id/features/:slug ─────────────────────────

  // Traceability: TC-033 → Spec 5.5 Endpoints
  test('TC-033: GET /api/projects/:id/features/:slug returns feature with tasks', async () => {
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    const featRes = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features`);
    const featData = JSON.parse(featRes.body);
    if (featData.features.length === 0) return;
    const slug = featData.features[0].slug;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features/${encodeURIComponent(slug)}`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.slug).toBe(slug);
    expect(data.tasks).toBeDefined();
    expect(data.meta).toBeDefined();
  });

  // ── GET /api/projects/:id/features/:slug/tasks/:taskId ───────────

  // Traceability: TC-034 → Spec 5.5 Endpoints
  test('TC-034: GET /api/projects/:id/features/:slug/tasks/:taskId returns task details', async () => {
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    const featRes = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features`);
    const featData = JSON.parse(featRes.body);
    if (featData.features.length === 0) return;
    const slug = featData.features[0].slug;
    // Get tasks to find a valid task ID
    const tasksRes = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features/${encodeURIComponent(slug)}/tasks`);
    const tasksData = JSON.parse(tasksRes.body);
    if (tasksData.tasks.length === 0) return;
    const taskId = tasksData.tasks[0].id;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features/${encodeURIComponent(slug)}/tasks/${encodeURIComponent(taskId)}`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.id).toBe(taskId);
    expect(data.title).toBeTruthy();
    expect(data.status).toBeTruthy();
    expect(data.priority).toBeTruthy();
    expect(data.scope).toBeDefined();
    expect(typeof data.breaking).toBe('boolean');
    expect(data.dependencies).toBeDefined();
    expect(data.meta).toBeDefined();
  });

  // ── GET /api/projects/:id/features/:slug/dependencies ────────────

  // Traceability: TC-035 → Spec 5.5 Endpoints + Story 6 Agent Flow
  test('TC-035: GET /api/projects/:id/features/:slug/dependencies returns dependency graph', async () => {
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    const featRes = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features`);
    const featData = JSON.parse(featRes.body);
    if (featData.features.length === 0) return;
    const slug = featData.features[0].slug;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features/${encodeURIComponent(slug)}/dependencies`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(Array.isArray(data.nodes)).toBe(true);
    expect(Array.isArray(data.edges)).toBe(true);
    expect(data.meta).toBeDefined();
    // Nodes should have task info
    for (const node of data.nodes) {
      expect(node.id).toBeTruthy();
      expect(node.title).toBeTruthy();
      expect(node.status).toBeTruthy();
    }
    // Edges should have source and target
    for (const edge of data.edges) {
      expect(edge.source).toBeTruthy();
      expect(edge.target).toBeTruthy();
      expect(typeof edge.crossFeature).toBe('boolean');
    }
  });

  // ── Error responses ──────────────────────────────────────────────

  // Traceability: TC-036 → Story 6 / AC-5 + Spec 5.5 Error Responses
  test('TC-036: API returns 404 for non-existent project', async () => {
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/non-existent-project-xyz`);
    expect(res.status).toBe(404);
    const data = JSON.parse(res.body);
    expect(data.error).toBe('not_found');
    expect(data.message).toBeTruthy();
  });

  // Traceability: TC-037 → Story 6 / AC-5 + Spec 5.5 Error Responses
  test('TC-037: API returns 404 for non-existent feature', async () => {
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features/non-existent-feature-xyz`);
    expect(res.status).toBe(404);
    const data = JSON.parse(res.body);
    expect(data.error).toBe('not_found');
    expect(data.message).toBeTruthy();
  });

  // Traceability: TC-038 → Story 6 / AC-5 + Spec 5.5 Error Responses
  test('TC-038: API returns 404 for non-existent task', async () => {
    const listRes = await curl('GET', `${apiBaseUrl()}/api/projects`);
    const listData = JSON.parse(listRes.body);
    const projectId = listData.projects[0].id;
    const featRes = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features`);
    const featData = JSON.parse(featRes.body);
    if (featData.features.length === 0) return;
    const slug = featData.features[0].slug;
    const res = await curl('GET', `${apiBaseUrl()}/api/projects/${encodeURIComponent(projectId)}/features/${encodeURIComponent(slug)}/tasks/non-existent-task-xyz`);
    expect(res.status).toBe(404);
    const data = JSON.parse(res.body);
    expect(data.error).toBe('not_found');
    expect(data.message).toBeTruthy();
  });
});
