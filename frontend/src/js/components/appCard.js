import { capitalize } from "../utils.js";
import { gitRefresh, openComponentTerminal, startComponent, stopComponent } from "../api.js";
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

  const leftWrap = document.createElement("div");
  leftWrap.className = "component-left-wrap";

  const infoWrap = document.createElement("div");
  infoWrap.className = "component-info";

  const nameEl = document.createElement("span");
  nameEl.textContent = component.name;
  infoWrap.appendChild(nameEl);

  const buttonsWrap = document.createElement("div");
  buttonsWrap.className = "component-actions";

  const startBtn = document.createElement("button");
  startBtn.textContent = "Start";
  startBtn.onclick = async () => {
    const res = await startComponent(appId, component.name);
    if (!res.ok) {
      const text = await res.text();
      alert(`Could not start ${component.name}: ${text || "unknown server error"}`);
      return;
    }

    const data = await res.json();
    if (data && data.url) component.url = data.url;
    refreshCurrentPage();
  };

  const terminalBtn = document.createElement("button");
  terminalBtn.textContent = "Terminal";
  terminalBtn.onclick = async () => {
    const res = await openComponentTerminal(appId, component.name);
    if (!res.ok) {
      const text = await res.text();
      alert(`Could not open terminal for ${component.name}: ${text || "unknown server error"}`);
      return;
    }
    refreshCurrentPage();
  };

  const stopBtn = document.createElement("button");
  stopBtn.textContent = "Stop";
  stopBtn.onclick = async () => {
    const res = await stopComponent(appId, component.name);
    if (!res.ok) {
      const text = await res.text();
      alert(`Could not stop ${component.name}: ${text || "unknown server error"}`);
      return;
    }

    const data = await res.json();
    if (data && data.url) component.url = data.url;
    refreshCurrentPage();
  };

  buttonsWrap.appendChild(startBtn);
  buttonsWrap.appendChild(terminalBtn);
  buttonsWrap.appendChild(stopBtn);

  leftWrap.appendChild(infoWrap);
  leftWrap.appendChild(buttonsWrap);

  const logs = Array.isArray(component.logs) ? component.logs.slice(-8) : [];
  if (logs.length > 0 || component.running) {
    const rightSide = document.createElement("div");
    rightSide.className = "component-right-panel";

    const header = document.createElement("div");
    header.className = "component-status-header";
    header.textContent = component.running ? "Running" : "Last output";
    rightSide.appendChild(header);

    if (component.running && component.url) {
      const link = document.createElement("a");
      link.className = "component-url";
      link.href = component.url;
      link.target = "_blank";
      link.rel = "noreferrer";
      link.textContent = `Open ${component.url}`;
      rightSide.appendChild(link);
    }

    if (logs.length > 0) {
      const logBlock = document.createElement("div");
      logBlock.className = "component-logs";
      logs.forEach((entry) => {
        const line = document.createElement("div");
        line.className = "component-log-line";
        line.textContent = entry;
        logBlock.appendChild(line);
      });
      rightSide.appendChild(logBlock);
    } else if (component.running) {
      const idle = document.createElement("div");
      idle.className = "component-log-line muted";
      idle.textContent = "Process is running with no output yet.";
      rightSide.appendChild(idle);
    }

    row.appendChild(leftWrap);
    row.appendChild(rightSide);
    return row;
  }

  row.appendChild(leftWrap);
  return row;
}
