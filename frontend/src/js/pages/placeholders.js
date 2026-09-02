import { listActivity, listApps } from "../api.js";

export function renderBrainstormPage(container) {
  container.innerHTML = `
    <div class="placeholder-page">
      <h2>Brainstorm</h2>
      <p>Idea space (tree-based seed-to-app view) — planned for v1.1, not yet built.</p>
    </div>
  `;
}

export async function renderActivityPage(container) {
  container.innerHTML = '<div class="activity-page"><p>Loading activity...</p></div>';
  try {
    const [events, apps] = await Promise.all([listActivity(), listApps()]);
    renderActivity(container, events, apps);
  } catch (err) {
    container.innerHTML = `<div class="activity-page"><h2>Activity</h2><p class="activity-error">Could not load activity: ${escapeHTML(err.message)}</p></div>`;
  }
}

function renderActivity(container, events, apps) {
  container.innerHTML = `
    <div class="activity-page">
      <div class="activity-heading">
        <div><p class="eyebrow">AUDIT TRAIL</p><h2>Activity</h2><p class="activity-subtitle">A clear record of what changed across your projects.</p></div>
        <span class="activity-count">${events.length} recent events</span>
      </div>
      <div class="activity-filters">
        <select id="activity-app"><option value="">All apps</option>${apps.map((entry) => `<option value="${escapeAttr(entry.id)}">${escapeHTML(entry.name)}</option>`).join("")}</select>
        <select id="activity-type"><option value="">All activity</option><option value="component.start">Starts</option><option value="component.stop">Stops</option><option value="git.refresh">Git refreshes</option><option value="app.updated">App updates</option><option value="app.archived">Archive actions</option></select>
        <select id="activity-outcome"><option value="">All outcomes</option><option value="success">Successful</option><option value="failure">Failed</option></select>
        <input id="activity-branch" type="search" placeholder="Filter branch" />
        <button id="activity-clear" type="button">Clear</button>
      </div>
      <div class="activity-layout">
        <div class="activity-feed">${renderEvents(events)}</div>
        <aside class="activity-detail"><p class="activity-detail-empty">Select an event to inspect its full context.</p></aside>
      </div>
      <button id="activity-more" class="activity-more" type="button" ${events.length < 100 ? "hidden" : ""}>Load more</button>
    </div>
  `;
  let currentEvents = events;
  let offset = events.length;
  const reload = async () => {
    const filters = { appId: document.getElementById("activity-app").value, eventType: document.getElementById("activity-type").value, outcome: document.getElementById("activity-outcome").value, branch: document.getElementById("activity-branch").value.trim() };
    try {
      const updated = await listActivity(filters);
      currentEvents = updated;
      offset = updated.length;
      const feed = container.querySelector(".activity-feed");
      feed.innerHTML = renderEvents(updated);
      bindEventDetails(container, updated);
      container.querySelector(".activity-count").textContent = `${updated.length} recent events`;
    } catch (err) {
      container.querySelector(".activity-feed").innerHTML = `<p class="activity-error">Could not apply filters: ${escapeHTML(err.message)}</p>`;
    }
  };
  ["activity-app", "activity-type", "activity-outcome"].forEach((id) => document.getElementById(id).addEventListener("change", reload));
  document.getElementById("activity-branch").addEventListener("change", reload);
  document.getElementById("activity-clear").addEventListener("click", () => { ["activity-app", "activity-type", "activity-outcome", "activity-branch"].forEach((id) => { document.getElementById(id).value = ""; }); reload(); });
  document.getElementById("activity-more").addEventListener("click", async () => {
    const filters = { limit: "100", offset, appId: document.getElementById("activity-app").value, eventType: document.getElementById("activity-type").value, outcome: document.getElementById("activity-outcome").value, branch: document.getElementById("activity-branch").value.trim() };
    const more = await listActivity(filters);
    currentEvents = [...currentEvents, ...more];
    offset += more.length;
    container.querySelector(".activity-feed").innerHTML = renderEvents(currentEvents);
    bindEventDetails(container, currentEvents);
    if (more.length < 100) container.querySelector("#activity-more").hidden = true;
  });
  bindEventDetails(container, currentEvents);
}

function renderEvents(events) {
  if (!events.length) return '<div class="activity-empty"><strong>No activity yet</strong><span>Actions performed through WIP will appear here.</span></div>';
  let lastDay = "";
  return events.map((event) => {
    const day = new Date(event.occurredAt).toLocaleDateString(undefined, { weekday: "long", month: "short", day: "numeric" });
    const heading = day === lastDay ? "" : `<h3 class="activity-day">${day}</h3>`;
    lastDay = day;
    return `${heading}<button class="activity-event ${event.outcome === "failure" ? "is-failure" : ""}" data-event-id="${event.id}" type="button"><span class="activity-event-mark"></span><span class="activity-event-copy"><strong>${escapeHTML(event.summary)}</strong><span>${escapeHTML(event.appName)}${event.branch ? ` / ${escapeHTML(event.branch)}` : ""}</span></span><time>${new Date(event.occurredAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</time><span class="activity-outcome">${event.outcome === "failure" ? "Failed" : "Done"}</span></button>`;
  }).join("");
}

function bindEventDetails(container, events) {
  container.querySelectorAll(".activity-event").forEach((button) => button.addEventListener("click", () => {
    const event = events.find((item) => String(item.id) === button.dataset.eventId);
    container.querySelectorAll(".activity-event").forEach((item) => item.classList.remove("selected"));
    button.classList.add("selected");
    container.querySelector(".activity-detail").innerHTML = `<p class="eyebrow">EVENT DETAILS</p><h3>${escapeHTML(event.summary)}</h3><dl><dt>App</dt><dd>${escapeHTML(event.appName)}</dd><dt>When</dt><dd>${new Date(event.occurredAt).toLocaleString()}</dd><dt>Branch</dt><dd>${escapeHTML(event.branch || "Not recorded")}</dd><dt>Build / check</dt><dd>${escapeHTML(event.build || "Not recorded")}</dd><dt>Lifecycle</dt><dd>${escapeHTML(event.lifecycleStatus || "Not recorded")}</dd><dt>Runtime</dt><dd>${escapeHTML(event.runtimeStatus || "Not recorded")}</dd><dt>Changes</dt><dd>${escapeHTML(event.changes || "Not recorded")}</dd></dl>${event.detail ? `<p class="activity-detail-error">${escapeHTML(event.detail)}</p>` : ""}`;
  }));
}

function escapeHTML(value) { return String(value || "").replace(/[&<>'"]/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[character]); }
function escapeAttr(value) { return escapeHTML(value); }