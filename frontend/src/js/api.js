// All backend communication lives here. Pages import from this module
// instead of calling fetch() directly, so there's exactly one place that
// knows the API's shape — matching the same reasoning as the Go side
// having one store.go per concern instead of scattered SQL everywhere.

export async function listApps() {
  const res = await fetch("/api/apps");
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export async function listArchivedApps() {
  const res = await fetch("/api/apps?archived=true");
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export async function getApp(id) {
  const res = await fetch(`/api/apps/${id}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

// createApp/updateApp/updateComponents return the raw Response rather than
// parsed JSON, since callers need to distinguish ok vs. error and show the
// error text — a plain throw-on-failure helper would lose that detail.
export function createApp(payload) {
  return fetch("/api/apps", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export function updateApp(id, payload) {
  return fetch(`/api/apps/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export function updateComponents(id, components) {
  return fetch(`/api/apps/${id}/components`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ components }),
  });
}

export function archiveApp(id, deleteFolder) {
  return fetch(`/api/apps/${id}/archive`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ deleteFolder }),
  });
}

export function unarchiveApp(id) {
  return fetch(`/api/apps/${id}/unarchive`, { method: "POST" });
}

export function startComponent(appId, componentName) {
  return postComponentAction(appId, "start", componentName);
}

export function stopComponent(appId, componentName) {
  return postComponentAction(appId, "stop", componentName);
}

function postComponentAction(appId, action, componentName) {
  return fetch(`/api/apps/${appId}/${action}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ component: componentName }),
  });
}

export function gitRefresh(id) {
  return fetch(`/api/apps/${id}/git-refresh`, { method: "POST" });
}

export function gitInit(path) {
  return fetch("/api/git-init", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path }),
  });
}

export async function gitStatus(path) {
  const res = await fetch(`/api/git-status?path=${encodeURIComponent(path)}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export async function browse(path) {
  const res = await fetch(`/api/browse?path=${encodeURIComponent(path || "")}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export async function getSettings() {
  const res = await fetch("/api/settings");
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export function updateSettings(payload) {
  return fetch("/api/settings", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}