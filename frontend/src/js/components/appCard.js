import { capitalize } from "../utils.js";
import { getApp, openComponentTerminal, openFolder, startComponent, stopComponent } from "../api.js";
import { openAppDetail, openGitPage, refreshCurrentPage } from "../router.js";
import { notifyRuntimeChanged } from "../runtimeEvents.js";

const collapsedAppIds = new Set();

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
  card.dataset.appId = entry.id;
  if (collapsedAppIds.has(entry.id)) card.classList.add("is-collapsed");

  const directory = entry.localPath || "Not available";
  const lastEdited = entry.lastTouchedAt
    ? new Date(entry.lastTouchedAt).toLocaleDateString()
    : "Not available";

  const anyRunning = (entry.components || []).some((c) => c.running);
  const appStatusBadge = `<span class="app-status status-${entry.status}">${capitalize(entry.status)}</span>`;
  const runtimeStatusBadge = anyRunning && entry.status !== "running" ? '<span class="app-runtime-status">Running</span>' : "";
  const runningUrl = (entry.components || []).find((c) => c.running && c.url)?.url || null;

  card.innerHTML = `
    <div class="app-card-header">
      <div class="app-card-main">
        <div class="app-name-row">
          <span class="app-name">${entry.name}</span>
        </div>
        <p class="app-description">${entry.description || ""}</p>
        <div class="app-meta">
          <a href="#" class="folder-link" title="Open folder in File Explorer"></a>
          <span>last edited: ${lastEdited}</span>
        </div>
        <div class="connection-pills"></div>
        <div class="app-components"></div>
      </div>
      <div class="app-card-terminal-slot"></div>
      <div class="app-card-status-panel">
        <div class="app-status-row">
          ${appStatusBadge}
          ${runtimeStatusBadge}
          <button class="card-collapse-toggle" type="button" aria-expanded="${!collapsedAppIds.has(entry.id)}">Collapse</button>
        </div>
        <div class="app-url-slot"></div>
      </div>
    </div>
  `;

  updateAppUrl(card, runningUrl);

  const collapseButton = card.querySelector(".card-collapse-toggle");
  const updateCollapseButton = () => {
    const collapsed = card.classList.contains("is-collapsed");
    collapseButton.textContent = collapsed ? "Expand" : "Collapse";
    collapseButton.setAttribute("aria-expanded", String(!collapsed));
    collapseButton.setAttribute("aria-label", `${collapsed ? "Expand" : "Collapse"} ${entry.name} card`);
  };
  collapseButton.addEventListener("click", (event) => {
    event.stopPropagation();
    const collapsed = card.classList.toggle("is-collapsed");
    if (collapsed) collapsedAppIds.add(entry.id);
    else collapsedAppIds.delete(entry.id);
    updateCollapseButton();
  });
  updateCollapseButton();

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
  pillsEl.appendChild(renderConnectionPill("Git", entry.gitConnected, false, () => openGitPage(entry.id), entry.gitDetails));
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
    const livePanel = createTerminalPanel(card, entry.id, firstRunningComponent);
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
      notifyRuntimeChanged();
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
      notifyRuntimeChanged();
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

export function renderConnectionPill(label, connected, comingSoon, onClick, details) {
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

    if (connected && details) {
      pill.tabIndex = 0;
      const wrapper = document.createElement("span");
      wrapper.className = "connection-pill-wrap";
      wrapper.appendChild(pill);

      const popover = document.createElement("span");
      popover.className = "git-details-popover";
      popover.setAttribute("role", "status");
      const rows = [
        ["Repository", details.repositoryName],
        ["Branch", details.branch],
        ["Last update", details.lastUpdate ? new Date(details.lastUpdate).toLocaleString() : "Not available"],
      ];
      rows.forEach(([labelText, value]) => {
        const row = document.createElement("span");
        row.className = "git-details-row";
        const key = document.createElement("strong");
        key.textContent = labelText;
        const text = document.createElement("span");
        text.textContent = value || "Not available";
        row.append(key, text);
        popover.appendChild(row);
      });
      if (details.repoUrl) {
        const row = document.createElement("span");
        row.className = "git-details-row";
        const key = document.createElement("strong");
        key.textContent = "Repository URL";
        const link = document.createElement("a");
        link.href = details.repoUrl;
        link.target = "_blank";
        link.rel = "noreferrer";
        link.textContent = details.repoUrl;
        link.addEventListener("click", (e) => e.stopPropagation());
        row.append(key, link);
        popover.appendChild(row);
      }
      wrapper.appendChild(popover);
      return wrapper;
    }
  }
  return pill;
}

function updateAppUrl(card, url) {
  const urlSlot = card.querySelector(".app-url-slot");
  if (!urlSlot) return;

  urlSlot.replaceChildren();
  if (!url) return;

  const label = document.createElement("div");
  label.className = "app-url-label";
  label.textContent = "Access app here";
  const link = document.createElement("a");
  link.className = "app-url-pill";
  link.href = url;
  link.target = "_blank";
  link.rel = "noreferrer";
  link.textContent = url;
  urlSlot.append(label, link);
}

function createTerminalPanel(card, appId, component) {
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
    const terminal = window.open(`/terminal.html?app=${encodeURIComponent(appId)}&component=${encodeURIComponent(component.name)}`, "wip-terminal", "popup,width=900,height=600");
    if (terminal) terminal.document.title = `${component.name} - WIP Terminal`;
  };
  terminalHeader.appendChild(popoutBtn);
  terminalPanel.appendChild(terminalHeader);

  renderTerminalOutput(terminalPanel, component);
  if (component.running) pollTerminalOutput(card, terminalPanel, appId, component.name);

  return terminalPanel;
}

function renderTerminalOutput(terminalPanel, component) {
  let logBlock = terminalPanel.querySelector(".component-logs");
  terminalPanel.querySelector(".component-log-empty")?.remove();

  const logs = Array.isArray(component.logs) ? component.logs.slice(-8) : [];
  if (logs.length > 0) {
    const wasAtBottom = !logBlock ||
      logBlock.scrollHeight - logBlock.scrollTop <= logBlock.clientHeight + 4;
    if (!logBlock) {
      logBlock = document.createElement("div");
      logBlock.className = "component-logs";
      terminalPanel.appendChild(logBlock);
    }
    logBlock.replaceChildren();
    logs.forEach((entry) => {
      const line = document.createElement("div");
      line.className = "component-log-line";
      line.textContent = entry;
      logBlock.appendChild(line);
    });
    if (wasAtBottom) logBlock.scrollTop = logBlock.scrollHeight;
  } else if (logBlock) {
    logBlock.remove();
  } else if (component.running) {
    const idle = document.createElement("div");
    idle.className = "component-log-line muted component-log-empty";
    idle.textContent = "Process is running with no output yet.";
    terminalPanel.appendChild(idle);
  }
}

function pollTerminalOutput(card, terminalPanel, appId, componentName) {
  setTimeout(async () => {
    if (!terminalPanel.isConnected) return;

    try {
      const app = await getApp(appId);
      const component = (app.components || []).find((item) => item.name === componentName);
      if (!component) return;

      renderTerminalOutput(terminalPanel, component);
      const runningUrl = (app.components || []).find((item) => item.running && item.url)?.url || null;
      updateAppUrl(card, runningUrl);
      if (component.running && terminalPanel.isConnected) {
        pollTerminalOutput(card, terminalPanel, appId, componentName);
      }
    } catch (err) {
      if (terminalPanel.isConnected) pollTerminalOutput(card, terminalPanel, appId, componentName);
    }
  }, 1000);
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
      const terminal = window.open(`/terminal.html?app=${encodeURIComponent(appId)}&component=${encodeURIComponent(component.name)}`, "wip-terminal", "popup,width=900,height=600");
      if (terminal) terminal.document.title = `${component.name} - WIP Terminal`;
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
