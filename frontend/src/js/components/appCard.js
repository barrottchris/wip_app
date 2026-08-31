import { capitalize } from "../utils.js";
import { gitRefresh, startComponent, stopComponent } from "../api.js";
import { openAppDetail, openGitPage, refreshCurrentPage } from "../router.js";

export function renderAppCard(entry) {
  const card = document.createElement("div");
  card.className = "app-card";

  const branch = entry.defaultBranch || "—";
  const lastTouched = entry.lastTouchedAt
    ? new Date(entry.lastTouchedAt).toLocaleDateString()
    : "—";

  const anyRunning = (entry.components || []).some((c) => c.running);
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
      <span>branch: ${branch}</span>
      <span>last touched: ${lastTouched}</span>
      <button class="refresh-git-btn" title="Refresh git status">&#8635;</button>
    </div>
    <div class="connection-pills"></div>
    <div class="app-components"></div>
  `;

  card.querySelector(".app-name").addEventListener("click", (e) => {
    e.stopPropagation();
    openAppDetail(entry.id);
  });

  card.querySelector(".refresh-git-btn").addEventListener("click", async (e) => {
    e.stopPropagation();
    const res = await gitRefresh(entry.id);
    if (!res.ok) {
      alert(`Could not refresh git status: ${await res.text()}`);
      return;
    }
    refreshCurrentPage();
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

  const browseLink = component.running && component.url
    ? `<a class="component-url" href="${component.url}" target="_blank" rel="noreferrer">Open ${component.url}</a>`
    : "";

  const startBtn = document.createElement("button");
  startBtn.textContent = "Start";
  startBtn.onclick = async () => {
    const res = await startComponent(appId, component.name);
    if (res.ok) {
      const data = await res.json();
      if (data && data.url) {
        component.url = data.url;
      }
    }
    refreshCurrentPage();
  };

  const stopBtn = document.createElement("button");
  stopBtn.textContent = "Stop";
  stopBtn.onclick = async () => {
    const res = await stopComponent(appId, component.name);
    if (res.ok) {
      const data = await res.json();
      if (data && data.url) {
        component.url = data.url;
      }
    }
    refreshCurrentPage();
  };

  row.innerHTML = `<span>${component.name}</span>${browseLink}`;
  row.appendChild(startBtn);
  row.appendChild(stopBtn);

  return row;
}
