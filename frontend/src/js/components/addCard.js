import { capitalize } from "../utils.js";
import { openFolder, startComponent, stopComponent } from "../api.js";
import { openAppDetail, openGitPage, refreshCurrentPage } from "../router.js";
import { notifyRuntimeChanged } from "../runtimeEvents.js";

export function renderAppCard(entry) {
  const card = document.createElement("div");
  card.className = "app-card";

  const directory = entry.localPath || "Not available";
  const lastEdited = entry.lastTouchedAt
    ? new Date(entry.lastTouchedAt).toLocaleDateString()
    : "Not available";

  const anyRunning = (entry.components || []).some((c) => c.running);
  // "Running" reflects live process state (computed, never stored).
  // Otherwise fall back to the user-set lifecycle status (active/paused/
  // abandoned/shipped) — these are deliberately different concepts.
  const statusBadge = anyRunning
    ? `<span class="app-status status-running">Running</span>`
    : `<span class="app-status status-${entry.status}">${capitalize(entry.status)}</span>`;

  card.innerHTML = `
    <div class="app-card-header">
      <span class="app-name">${entry.name}</span>
      ${statusBadge}
    </div>
    <p class="app-description">${entry.description || ""}</p>
    <div class="app-meta">
      <a href="#" class="folder-link" title="Open folder in File Explorer"></a>
      <span>last edited: ${lastEdited}</span>
    </div>
    <div class="connection-pills"></div>
    <div class="app-components"></div>
  `;

  const folderLink = card.querySelector(".folder-link");
  folderLink.textContent = `Folder directory: ${directory}`;
  if (entry.localPath) {
    folderLink.addEventListener("click", async (e) => {
      e.preventDefault();
      e.stopPropagation();
      const res = await openFolder(entry.localPath);
      if (!res.ok) alert(`Could not open folder: ${await res.text()}`);
    });
  } else {
    folderLink.removeAttribute("href");
    folderLink.classList.add("folder-link-unavailable");
  }

  card.querySelector(".app-name").addEventListener("click", (e) => {
    e.stopPropagation();
    openAppDetail(entry.id);
  });

  const pillsEl = card.querySelector(".connection-pills");
  pillsEl.appendChild(renderConnectionPill("Git", entry.gitConnected, false, () => openGitPage(entry.id)));
  pillsEl.appendChild(renderConnectionPill("Jira", entry.jiraConnected, entry.jiraComingSoon));
  pillsEl.appendChild(renderConnectionPill("Confluence", entry.confluenceConnected, entry.confluenceComingSoon));

  const componentsEl = card.querySelector(".app-components");
  (entry.components || []).forEach((component) => {
    componentsEl.appendChild(renderComponentRow(entry.id, component));
  });

  return card;
}

// renderConnectionPill: a small clickable tag showing whether an
// integration is connected — the label itself states the status (e.g.
// "Git: Connected" / "Git: Not connected") rather than relying on color
// alone. Jira/Confluence are always shown greyed out with a "coming soon"
// state since those integrations aren't built yet (see mvp-scope.md) —
// clicking them does nothing, rather than pretending there's somewhere to go.
// onClick is only wired up when the integration actually has somewhere to
// go (currently just Git, which opens its own dedicated page).
export function renderConnectionPill(label, connected, comingSoon, onClick) {
  const pill = document.createElement("span");
  pill.className = "connection-pill";
  if (comingSoon) {
    pill.classList.add("pill-coming-soon");
    pill.textContent = `${label}: Coming soon`;
    pill.title = `${label} integration is planned but not available yet`;
  } else {
    pill.classList.add(connected ? "pill-connected" : "pill-disconnected");
    pill.textContent = `${label}: ${connected ? "Connected" : "Not connected"}`;
    pill.title = connected
      ? `${label} is connected — click to view details`
      : `${label} is not connected — click to configure`;
    if (onClick) {
      pill.addEventListener("click", (e) => {
        e.stopPropagation();
        onClick();
      });
    }
  }
  return pill;
}

function renderComponentRow(appId, component) {
  const row = document.createElement("div");
  row.className = "component-row";

  const startBtn = document.createElement("button");
  startBtn.textContent = "Start";
  startBtn.onclick = async () => {
    const res = await startComponent(appId, component.name);
    if (res.ok) notifyRuntimeChanged();
    refreshCurrentPage();
  };

  const stopBtn = document.createElement("button");
  stopBtn.textContent = "Stop";
  stopBtn.onclick = async () => {
    const res = await stopComponent(appId, component.name);
    if (res.ok) notifyRuntimeChanged();
    refreshCurrentPage();
  };

  row.innerHTML = `<span>${component.name}</span>`;
  row.appendChild(startBtn);
  row.appendChild(stopBtn);

  return row;
}