import { capitalize } from "../utils.js";
import { getApp, gitRefresh, openComponentTerminal, startComponent, stopComponent } from "../api.js";
import { openAppDetail, openGitPage, refreshCurrentPage } from "../router.js";

async function refreshCard(card, appId) {
  try {
    const updated = await getApp(appId);
    const replacement = renderAppCard(updated);
    if (card && card.isConnected) {
      card.replaceWith(replacement);
    }
  } catch (err) {
    // Keep the current card in place if the refresh fails; the user can retry.
  }
}

export function renderAppCard(entry) {
  const card = document.createElement("div");
  card.className = "app-card";

  const branch = entry.defaultBranch || "—";
  const lastTouched = entry.lastTouchedAt
    ? new Date(entry.lastTouchedAt).toLocaleDateString()
    : "—";

  const anyRunning = (entry.components || []).some((c) => c.running);
  const appStatusBadge = `<span class="app-status status-${entry.status}">${capitalize(entry.status)}</span>`;
  const runtimeStatusBadge = anyRunning && entry.status !== "running" ? '<span class="app-runtime-status">Running</span>' : "";
  const runningUrl = (entry.components || []).find((c) => c.running && c.url)?.url || null;

  card.innerHTML = `
    <div class="app-card-header">
      <div class="app-card-main">
        <span class="app-name">${entry.name}</span>
        <p class="app-description">${entry.description || ""}</p>
        <div class="app-meta">
          <span>branch: ${branch}</span>
          <span>last touched: ${lastTouched}</span>
          <button class="refresh-git-btn" title="Refresh git status">&#8635;</button>
        </div>
        <div class="connection-pills"></div>
        <div class="app-components"></div>
      </div>
      <div class="app-card-terminal-slot"></div>
      <div class="app-card-status-panel">
        ${appStatusBadge}
        ${runtimeStatusBadge}
        ${runningUrl ? `<div class="app-url-label">Access app here</div><a class="app-url-pill" href="${runningUrl}" target="_blank" rel="noreferrer">${runningUrl}</a>` : ""}
      </div>
    </div>
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
    await refreshCard(card, entry.id);
  });

  const pillsEl = card.querySelector(".connection-pills");
  pillsEl.appendChild(renderConnectionPill("Git", entry.gitConnected, false, () => openGitPage(entry.id)));
  pillsEl.appendChild(renderConnectionPill("Jira", entry.jiraConnected, entry.jiraComingSoon));
  pillsEl.appendChild(renderConnectionPill("Confluence", entry.confluenceConnected, entry.confluenceComingSoon));

  const componentsEl = card.querySelector(".app-components");
  const terminalSlot = card.querySelector(".app-card-terminal-slot");
  let firstRunningComponent = null;
  for (const component of entry.components || []) {
    if (component.running || (Array.isArray(component.logs) && component.logs.length > 0)) {
      firstRunningComponent = component;
      break;
    }
  }

  if (firstRunningComponent) {
    const livePanel = createTerminalPanel(entry.id, firstRunningComponent);
    terminalSlot.appendChild(livePanel);
  }

  const controlsRow = document.createElement("div");
  controlsRow.className = "controls-row";
  const controlsWrap = document.createElement("div");
  controlsWrap.className = "controls-wrap";
  controlsWrap.textContent = "App";
  const buttonRow = document.createElement("div");
  buttonRow.className = "component-actions";

  (entry.components || []).forEach((component) => {
    const startBtn = document.createElement("button");
    startBtn.textContent = "Start";
    startBtn.onclick = async () => {
      const res = await startComponent(entry.id, component.name);
      if (!res.ok) {
        const text = await res.text();
        alert(`Could not start ${component.name}: ${text || "unknown server error"}`);
        return;
      }
      await refreshCard(card, entry.id);
    };

    const stopBtn = document.createElement("button");
    stopBtn.textContent = "Stop";
    stopBtn.onclick = async () => {
      const res = await stopComponent(entry.id, component.name);
      if (!res.ok) {
        const text = await res.text();
        alert(`Could not stop ${component.name}: ${text || "unknown server error"}`);
        return;
      }
      await refreshCard(card, entry.id);
    };

    buttonRow.appendChild(startBtn);
    buttonRow.appendChild(stopBtn);
  });

  controlsWrap.appendChild(buttonRow);
  controlsRow.appendChild(controlsWrap);
  componentsEl.appendChild(controlsRow);

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

function createTerminalPanel(appId, component) {
  const logs = Array.isArray(component.logs) ? component.logs.slice(-8) : [];
  const terminalPanel = document.createElement("div");
  terminalPanel.className = "component-terminal-panel";

  const terminalHeader = document.createElement("div");
  terminalHeader.className = "component-terminal-header";

  const terminalTitle = document.createElement("span");
  terminalTitle.className = "component-terminal-title";
  terminalTitle.textContent = "Live output";
  terminalHeader.appendChild(terminalTitle);

  const popoutBtn = document.createElement("button");
  popoutBtn.className = "component-popout-btn";
  popoutBtn.textContent = "Open terminal";
  popoutBtn.onclick = async () => {
    const res = await openComponentTerminal(appId, component.name);
    if (!res.ok) {
      const text = await res.text();
      alert(`Could not open terminal for ${component.name}: ${text || "unknown server error"}`);
      return;
    }
  };
  terminalHeader.appendChild(popoutBtn);
  terminalPanel.appendChild(terminalHeader);

  if (logs.length > 0) {
    const logBlock = document.createElement("div");
    logBlock.className = "component-logs";
    logs.forEach((entry) => {
      const line = document.createElement("div");
      line.className = "component-log-line";
      line.textContent = entry;
      logBlock.appendChild(line);
    });
    terminalPanel.appendChild(logBlock);
  } else if (component.running) {
    const idle = document.createElement("div");
    idle.className = "component-log-line muted";
    idle.textContent = "Process is running with no output yet.";
    terminalPanel.appendChild(idle);
  }

  return terminalPanel;
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
    await refreshCard(row.closest(".app-card"), appId);
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
    await refreshCard(row.closest(".app-card"), appId);
  };

  buttonsWrap.appendChild(startBtn);
  buttonsWrap.appendChild(stopBtn);

  leftWrap.appendChild(infoWrap);
  leftWrap.appendChild(buttonsWrap);

  const logs = Array.isArray(component.logs) ? component.logs.slice(-8) : [];
  if (logs.length > 0 || component.running) {
    const middle = document.createElement("div");
    middle.className = "component-middle-panel";

    const terminalPanel = document.createElement("div");
    terminalPanel.className = "component-terminal-panel";

    const terminalHeader = document.createElement("div");
    terminalHeader.className = "component-terminal-header";

    const terminalTitle = document.createElement("span");
    terminalTitle.className = "component-terminal-title";
    terminalTitle.textContent = "Live output";
    terminalHeader.appendChild(terminalTitle);

    const popoutBtn = document.createElement("button");
    popoutBtn.className = "component-popout-btn";
    popoutBtn.textContent = "Open terminal";
    popoutBtn.onclick = async () => {
      const res = await openComponentTerminal(appId, component.name);
      if (!res.ok) {
        const text = await res.text();
        alert(`Could not open terminal for ${component.name}: ${text || "unknown server error"}`);
        return;
      }
      await refreshCard(row.closest(".app-card"), appId);
    };
    terminalHeader.appendChild(popoutBtn);
    terminalPanel.appendChild(terminalHeader);

    if (logs.length > 0) {
      const logBlock = document.createElement("div");
      logBlock.className = "component-logs";
      logs.forEach((entry) => {
        const line = document.createElement("div");
        line.className = "component-log-line";
        line.textContent = entry;
        logBlock.appendChild(line);
      });
      terminalPanel.appendChild(logBlock);
    } else if (component.running) {
      const idle = document.createElement("div");
      idle.className = "component-log-line muted";
      idle.textContent = "Process is running with no output yet.";
      terminalPanel.appendChild(idle);
    }

    middle.appendChild(terminalPanel);

    row.appendChild(leftWrap);
    row.appendChild(middle);
    return row;
  }

  row.appendChild(leftWrap);
  return row;
}
