// App shell: top banner + left nav + a simple client-side router that swaps
// content in #page-content. No framework — plain DOM manipulation, matching
// the rest of this scaffold's "no build step" approach.

const pages = {
  registry: renderRegistryPage,
  archived: renderArchivedPage,
  brainstorm: renderBrainstormPage,
  activity: renderActivityPage,
  settings: renderSettingsPage,
  "add-app": renderAddAppPage,
  "app-detail": renderAppDetailPage,
};

// Which app the detail/edit page is currently showing — set by
// openAppDetail() before navigating there. Simple global state, matching
// this scaffold's "no framework" approach (see addAppState for the same
// pattern used by onboarding).
let selectedAppId = null;

function navigateTo(pageName) {
  document.querySelectorAll(".nav-item").forEach((el) => {
    el.classList.toggle("active", el.dataset.page === pageName);
  });
  const content = document.getElementById("page-content");
  content.innerHTML = "";
  const renderFn = pages[pageName] || renderRegistryPage;
  renderFn(content);
}

document.querySelectorAll(".nav-item").forEach((el) => {
  el.addEventListener("click", () => navigateTo(el.dataset.page));
});

const addAppBtn = document.getElementById("add-app-btn");
if (addAppBtn) {
  addAppBtn.addEventListener("click", () => navigateTo("add-app"));
}

// --- Registry page (the original scaffold content) ---

async function renderRegistryPage(container) {
  container.innerHTML = "Loading...";
  try {
    const res = await fetch("/api/apps");
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const apps = await res.json();
    container.innerHTML = "";
    updateRunningCount(apps);
    apps.forEach((entry) => container.appendChild(renderAppCard(entry)));
  } catch (err) {
    container.innerHTML = `<p>Error loading apps: ${err}</p>`;
  }
}

// Archived view — same data shape as the registry, but apps here are
// hidden from the main list (see next-phase-plan.md: hidden by default,
// dedicated Archived view, not just greyed out inline). Cards are rendered
// simply here rather than reusing renderAppCard's full start/stop/pill
// treatment, since an archived app has one relevant action: unarchive it.
async function renderArchivedPage(container) {
  container.innerHTML = "Loading...";
  try {
    const res = await fetch("/api/apps?archived=true");
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const apps = await res.json();
    container.innerHTML = "";

    if (apps.length === 0) {
      container.innerHTML = `<p class="hint">No archived apps.</p>`;
      return;
    }

    apps.forEach((entry) => container.appendChild(renderArchivedCard(entry)));
  } catch (err) {
    container.innerHTML = `<p>Error loading archived apps: ${err}</p>`;
  }
}

function renderArchivedCard(entry) {
  const card = document.createElement("div");
  card.className = "app-card archived-card";

  card.innerHTML = `
    <div class="app-card-header">
      <span class="app-name">${entry.name}</span>
    </div>
    <p class="app-description">${entry.description || ""}</p>
    <p class="hint">${entry.localPath || ""}</p>
  `;

  const unarchiveBtn = document.createElement("button");
  unarchiveBtn.textContent = "Unarchive";
  unarchiveBtn.addEventListener("click", async () => {
    await fetch(`/api/apps/${entry.id}/unarchive`, { method: "POST" });
    renderArchivedPage(document.getElementById("page-content"));
  });
  card.appendChild(unarchiveBtn);

  return card;
}

function updateRunningCount(apps) {
  const runningCount = apps.reduce((count, entry) => {
    return count + (entry.components || []).filter((c) => c.running).length;
  }, 0);
  const el = document.getElementById("running-count");
  if (el) el.textContent = `${runningCount} running`;
}

function renderAppCard(entry) {
  const card = document.createElement("div");
  card.className = "app-card";

  const branch = entry.defaultBranch || "—";
  const lastTouched = entry.lastTouchedAt
    ? new Date(entry.lastTouchedAt).toLocaleDateString()
    : "—";

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
      <span>branch: ${branch}</span>
      <span>last touched: ${lastTouched}</span>
    </div>
    <div class="connection-pills"></div>
    <div class="app-components"></div>
  `;

  card.querySelector(".app-name").addEventListener("click", (e) => {
    e.stopPropagation();
    openAppDetail(entry.id);
  });

  const pillsEl = card.querySelector(".connection-pills");
  pillsEl.appendChild(renderConnectionPill("Git", entry.gitConnected, false, entry.id));
  pillsEl.appendChild(renderConnectionPill("Jira", entry.jiraConnected, entry.jiraComingSoon, entry.id));
  pillsEl.appendChild(renderConnectionPill("Confluence", entry.confluenceConnected, entry.confluenceComingSoon, entry.id));

  const componentsEl = card.querySelector(".app-components");
  (entry.components || []).forEach((component) => {
    componentsEl.appendChild(renderComponentRow(entry.id, component));
  });

  return card;
}

function capitalize(s) {
  return (s || "").charAt(0).toUpperCase() + (s || "").slice(1);
}

// renderConnectionPill: a small clickable tag showing whether an
// integration is connected — the label itself states the status (e.g.
// "Git: Connected" / "Git: Not connected") rather than relying on color
// alone. Jira/Confluence are always shown greyed out with a "coming soon"
// state since those integrations aren't built yet (see mvp-scope.md) —
// clicking them does nothing, rather than pretending there's somewhere to go.
function renderConnectionPill(label, connected, comingSoon, appId) {
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
    pill.addEventListener("click", (e) => {
      e.stopPropagation();
      openAppDetail(appId, "git-section");
    });
  }
  return pill;
}

// pendingScrollTarget: set by a pill click before navigating, so the detail
// page can scroll straight to the relevant section once it's rendered
// (e.g. clicking the Git pill from the registry should land you on the
// Git section of that app's detail page, not just the top of the page).
let pendingScrollTarget = null;

function openAppDetail(appId, scrollTargetId) {
  selectedAppId = appId;
  pendingScrollTarget = scrollTargetId || null;
  navigateTo("app-detail");
}

function renderComponentRow(appId, component) {
  const row = document.createElement("div");
  row.className = "component-row";

  const startBtn = document.createElement("button");
  startBtn.textContent = "Start";
  startBtn.onclick = async () => {
    await postAction(appId, "start", component.name);
    navigateTo("registry");
  };

  const stopBtn = document.createElement("button");
  stopBtn.textContent = "Stop";
  stopBtn.onclick = async () => {
    await postAction(appId, "stop", component.name);
    navigateTo("registry");
  };

  row.innerHTML = `<span>${component.name}</span>`;
  row.appendChild(startBtn);
  row.appendChild(stopBtn);

  return row;
}

async function postAction(appId, action, componentName) {
  await fetch(`/api/apps/${appId}/${action}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ component: componentName }),
  });
}

// --- Settings page ---

async function renderSettingsPage(container) {
  container.innerHTML = "Loading settings...";
  try {
    const res = await fetch("/api/settings");
    const settings = await res.json();
    container.innerHTML = "";
    container.appendChild(buildSettingsForm(settings));
  } catch (err) {
    container.innerHTML = `<p>Error loading settings: ${err}</p>`;
  }
}

function buildSettingsForm(settings) {
  const wrapper = document.createElement("div");
  wrapper.className = "settings-page";

  wrapper.innerHTML = `
    <h2>Storage</h2>
    <label>Managed apps folder</label>
    <input type="text" id="managed-root" value="${settings.managedRoot || ""}" />
    <p class="hint">Where WIP looks for and organizes tracked apps.</p>

    <h2>GitHub</h2>
    <label>Username</label>
    <input type="text" id="github-username" value="${settings.githubUsername || ""}" />
    <label>Personal access token</label>
    <input type="password" id="github-token" placeholder="${settings.githubTokenIsSet ? "•••••••• (already set)" : "Not set"}" />
    <p class="hint">Used to bring apps under git and check repo status. Never displayed once saved.</p>

    <button id="save-settings-btn">Save settings</button>
    <span id="save-status" class="hint"></span>
  `;

  wrapper.querySelector("#save-settings-btn").onclick = async () => {
    const body = {
      managedRoot: wrapper.querySelector("#managed-root").value,
      githubUsername: wrapper.querySelector("#github-username").value,
      githubToken: wrapper.querySelector("#github-token").value,
    };
    await fetch("/api/settings", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    wrapper.querySelector("#save-status").textContent = "Saved.";
  };

  return wrapper;
}

// --- Add App page (onboarding) ---

// Local state for the in-progress add-app form, reset each time the page
// is rendered. Kept simple/global since this is a single-page-at-a-time
// hand-rolled router, not a component framework.
let addAppState = {
  mode: "existing", // "existing" | "new"
  selectedPath: null,
  browsePath: "",
};

async function renderAddAppPage(container) {
  addAppState = { mode: "existing", selectedPath: null, browsePath: "" };

  const wrapper = document.createElement("div");
  wrapper.className = "add-app-page";
  wrapper.innerHTML = `
    <h2>Add app</h2>

    <div class="mode-toggle">
      <button class="mode-btn active" data-mode="existing">Existing folder</button>
      <button class="mode-btn" data-mode="new">Create new</button>
    </div>

    <label>Name</label>
    <input type="text" id="new-app-name" placeholder="My App" />

    <label>Description</label>
    <input type="text" id="new-app-description" placeholder="What is this app?" />

    <label id="folder-label">Choose an existing folder</label>
    <div id="folder-picker"></div>
    <p class="hint" id="selected-path-hint"></p>

    <div id="git-prompt" style="display:none;" class="git-prompt">
      <p>This folder isn't under git yet. Initialize it?</p>
      <button id="git-init-yes">Yes, run git init</button>
      <button id="git-init-skip">Skip for now</button>
    </div>

    <button id="create-app-btn" disabled>Add app</button>
    <span id="create-status" class="hint"></span>
  `;
  container.appendChild(wrapper);

  wrapper.querySelectorAll(".mode-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      wrapper.querySelectorAll(".mode-btn").forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      addAppState.mode = btn.dataset.mode;
      addAppState.selectedPath = null;
      wrapper.querySelector("#folder-label").textContent =
        addAppState.mode === "existing"
          ? "Choose an existing folder"
          : "Choose where to create the new folder";
      wrapper.querySelector("#selected-path-hint").textContent = "";
      wrapper.querySelector("#git-prompt").style.display = "none";
      updateCreateButtonState(wrapper);
      renderFolderPicker(wrapper, addAppState.browsePath);
    });
  });

  wrapper.querySelector("#new-app-name").addEventListener("input", () => updateCreateButtonState(wrapper));

  wrapper.querySelector("#git-init-yes").addEventListener("click", async () => {
    await fetch("/api/git-init", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path: addAppState.selectedPath }),
    });
    wrapper.querySelector("#git-prompt").style.display = "none";
  });
  wrapper.querySelector("#git-init-skip").addEventListener("click", () => {
    wrapper.querySelector("#git-prompt").style.display = "none";
  });

  wrapper.querySelector("#create-app-btn").addEventListener("click", async () => {
    const name = wrapper.querySelector("#new-app-name").value.trim();
    const description = wrapper.querySelector("#new-app-description").value.trim();
    if (!name || !addAppState.selectedPath) return;

    const res = await fetch("/api/apps", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name,
        description,
        localPath: addAppState.selectedPath,
        createNew: addAppState.mode === "new",
      }),
    });

    if (res.ok) {
      navigateTo("registry");
    } else {
      const errText = await res.text();
      wrapper.querySelector("#create-status").textContent = `Error: ${errText}`;
    }
  });

  await renderFolderPicker(wrapper, "");
}

function updateCreateButtonState(wrapper) {
  const name = wrapper.querySelector("#new-app-name").value.trim();
  const btn = wrapper.querySelector("#create-app-btn");
  btn.disabled = !(name && addAppState.selectedPath);
}

// renderFolderPicker fetches a directory listing from the Go backend and
// renders it as a simple clickable list — the browser can't give us a real
// native OS path, so the backend does the browsing instead.
async function renderFolderPicker(wrapper, path) {
  const pickerEl = wrapper.querySelector("#folder-picker");
  pickerEl.innerHTML = "Loading...";

  const res = await fetch(`/api/browse?path=${encodeURIComponent(path || "")}`);
  if (!res.ok) {
    pickerEl.innerHTML = "Could not browse this location.";
    return;
  }
  const listing = await res.json();
  addAppState.browsePath = listing.currentPath;

  pickerEl.innerHTML = "";

  const currentRow = document.createElement("div");
  currentRow.className = "folder-current-row";
  currentRow.innerHTML = `<span>${listing.currentPath}</span>`;
  const selectBtn = document.createElement("button");
  selectBtn.textContent =
    addAppState.mode === "existing" ? "Select this folder" : "Create here";
  selectBtn.addEventListener("click", () => selectFolder(wrapper, listing.currentPath));
  currentRow.appendChild(selectBtn);
  pickerEl.appendChild(currentRow);

  const list = document.createElement("div");
  list.className = "folder-list";

  if (listing.parentPath) {
    const upRow = document.createElement("div");
    upRow.className = "folder-row";
    upRow.textContent = ".. (up)";
    upRow.addEventListener("click", () => renderFolderPicker(wrapper, listing.parentPath));
    list.appendChild(upRow);
  }

  (listing.directories || []).forEach((dir) => {
    const row = document.createElement("div");
    row.className = "folder-row";
    row.textContent = dir.name;
    row.addEventListener("click", () => renderFolderPicker(wrapper, dir.path));
    list.appendChild(row);
  });

  pickerEl.appendChild(list);
}

async function selectFolder(wrapper, path) {
  addAppState.selectedPath = path;
  wrapper.querySelector("#selected-path-hint").textContent = `Selected: ${path}`;
  updateCreateButtonState(wrapper);

  if (addAppState.mode === "existing") {
    const res = await fetch(`/api/git-status?path=${encodeURIComponent(path)}`);
    if (res.ok) {
      const { hasGit } = await res.json();
      wrapper.querySelector("#git-prompt").style.display = hasGit ? "none" : "block";
    }
  } else {
    wrapper.querySelector("#git-prompt").style.display = "none";
  }
}

// --- App Detail / Edit page ---

async function renderAppDetailPage(container) {
  if (!selectedAppId) {
    container.innerHTML = `<p>No app selected. <a href="#" id="back-to-registry">Back to registry</a></p>`;
    container.querySelector("#back-to-registry").addEventListener("click", (e) => {
      e.preventDefault();
      navigateTo("registry");
    });
    return;
  }

  container.innerHTML = "Loading...";
  const res = await fetch(`/api/apps/${selectedAppId}`);
  if (!res.ok) {
    container.innerHTML = `<p>Could not load app.</p>`;
    return;
  }
  const entry = await res.json();
  container.innerHTML = "";

  // Local, page-scoped state for the folder picker used when changing the
  // path — kept separate from addAppState so onboarding and editing never
  // interfere with each other.
  const editState = { selectedPath: entry.localPath, browsing: false };

  const wrapper = document.createElement("div");
  wrapper.className = "add-app-page"; // reuse onboarding form styling
  wrapper.innerHTML = `
    <h2>${entry.name}</h2>

    <div class="connection-pills detail-pills"></div>

    <label>Name</label>
    <input type="text" id="edit-name" value="${escapeAttr(entry.name)}" />

    <label>Description</label>
    <input type="text" id="edit-description" value="${escapeAttr(entry.description || "")}" />

    <label>Status</label>
    <select id="edit-status">
      <option value="active">Active (being worked on)</option>
      <option value="paused">Paused</option>
      <option value="abandoned">Abandoned</option>
      <option value="shipped">Shipped</option>
    </select>
    <p class="hint">This is separate from whether it's actually running right now — that's shown live above, based on its components.</p>

    <label>Folder</label>
    <p class="hint" id="current-path-display">${escapeAttr(entry.localPath)}</p>
    <button id="change-folder-btn">Change folder</button>
    <div id="edit-folder-picker" style="display:none; margin-top:0.5rem;"></div>

    <div id="git-section" class="detail-section">
      <h2>Git</h2>
      <p class="hint" id="git-status-text">
        ${entry.gitConnected ? "This folder is under git." : "This folder is not under git yet."}
      </p>
      <p class="hint">Remote: ${entry.repoUrl ? escapeAttr(entry.repoUrl) : "none set"}</p>
      <button id="git-init-btn" style="${entry.gitConnected ? "display:none;" : ""}">Initialize git here</button>
    </div>

    <div id="components-section" class="detail-section">
      <h2>Components</h2>
      <p class="hint">Each component gets its own start/stop command and run mode. Add as many as this app needs.</p>
      <div id="component-rows"></div>
      <button id="add-component-btn">+ Add component</button>
      <div style="margin-top:0.75rem;">
        <button id="save-components-btn">Save components</button>
        <span id="components-status-msg" class="hint"></span>
      </div>
    </div>

    <div style="margin-top:1.25rem;">
      <button id="save-edit-btn">Save changes</button>
      <button id="cancel-edit-btn">Cancel</button>
      <button id="archive-btn" class="danger-btn">Archive</button>
      <span id="edit-status-msg" class="hint"></span>
    </div>
  `;
  container.appendChild(wrapper);

  // Component rows are built from the entry's current components, then
  // edited/added/removed entirely client-side until "Save components" is
  // clicked — which submits the whole list at once (see UpdateComponents).
  const componentRowsEl = wrapper.querySelector("#component-rows");
  (entry.components || []).forEach((component) => addComponentRow(componentRowsEl, component));

  wrapper.querySelector("#add-component-btn").addEventListener("click", () => {
    addComponentRow(componentRowsEl, { name: "", startCommand: "", stopCommand: "", runMode: "native" });
  });

  wrapper.querySelector("#save-components-btn").addEventListener("click", async () => {
    const components = Array.from(componentRowsEl.querySelectorAll(".component-edit-row")).map((row) => ({
      name: row.querySelector(".comp-name").value.trim(),
      startCommand: row.querySelector(".comp-start").value.trim(),
      stopCommand: row.querySelector(".comp-stop").value.trim(),
      runMode: row.querySelector(".comp-runmode").value,
    }));

    const statusMsgEl = wrapper.querySelector("#components-status-msg");

    if (components.some((c) => !c.name)) {
      statusMsgEl.textContent = "Every component needs a name.";
      return;
    }

    const res = await fetch(`/api/apps/${entry.id}/components`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ components }),
    });

    statusMsgEl.textContent = res.ok ? "Saved." : `Error: ${await res.text()}`;
  });

  wrapper.querySelector("#git-init-btn")?.addEventListener("click", async () => {
    await fetch("/api/git-init", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path: entry.localPath }),
    });
    // Re-render the whole page so the pill, status text, and button all
    // reflect the freshly-initialized repo rather than being patched by hand.
    pendingScrollTarget = "git-section";
    renderAppDetailPage(container);
  });

  if (pendingScrollTarget) {
    const target = wrapper.querySelector(`#${pendingScrollTarget}`);
    if (target) target.scrollIntoView({ behavior: "smooth", block: "center" });
    pendingScrollTarget = null;
  }

  wrapper.querySelector("#edit-status").value = entry.status;

  const pillsEl = wrapper.querySelector(".detail-pills");
  pillsEl.appendChild(renderConnectionPill("Git", entry.gitConnected, false, entry.id));
  pillsEl.appendChild(renderConnectionPill("Jira", entry.jiraConnected, entry.jiraComingSoon, entry.id));
  pillsEl.appendChild(renderConnectionPill("Confluence", entry.confluenceConnected, entry.confluenceComingSoon, entry.id));

  wrapper.querySelector("#change-folder-btn").addEventListener("click", async () => {
    editState.browsing = !editState.browsing;
    const pickerEl = wrapper.querySelector("#edit-folder-picker");
    pickerEl.style.display = editState.browsing ? "block" : "none";
    if (editState.browsing) {
      await renderEditFolderPicker(wrapper, editState, editState.selectedPath);
    }
  });

  wrapper.querySelector("#cancel-edit-btn").addEventListener("click", () => navigateTo("registry"));

  wrapper.querySelector("#archive-btn").addEventListener("click", async () => {
    const confirmed = confirm(`Archive "${entry.name}"? It will be hidden from the registry but can be restored from the Archived view.`);
    if (!confirmed) return;

    // Deliberately a second, separate confirmation — deleting the folder
    // is the one genuinely irreversible action here, so it must never be
    // bundled into the softer "archive" confirmation above.
    const alsoDelete = confirm("Also delete the folder from disk? This cannot be undone.");

    await fetch(`/api/apps/${entry.id}/archive`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ deleteFolder: alsoDelete }),
    });
    navigateTo("registry");
  });

  wrapper.querySelector("#save-edit-btn").addEventListener("click", async () => {
    const name = wrapper.querySelector("#edit-name").value.trim();
    const description = wrapper.querySelector("#edit-description").value.trim();
    const status = wrapper.querySelector("#edit-status").value;
    if (!name || !editState.selectedPath) return;

    const putRes = await fetch(`/api/apps/${entry.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, description, localPath: editState.selectedPath, status }),
    });

    if (putRes.ok) {
      navigateTo("registry");
    } else {
      const errText = await putRes.text();
      wrapper.querySelector("#edit-status-msg").textContent = `Error: ${errText}`;
    }
  });
}

function escapeAttr(s) {
  return (s || "").replace(/"/g, "&quot;");
}

// addComponentRow builds one editable component row (name, start/stop
// commands, run mode, remove button) and appends it to containerEl. Used
// both for existing components (pre-filled) and the "+ Add component"
// button (blank row).
function addComponentRow(containerEl, component) {
  const row = document.createElement("div");
  row.className = "component-edit-row";

  row.innerHTML = `
    <input type="text" class="comp-name" placeholder="Name (e.g. Frontend)" value="${escapeAttr(component.name)}" />
    <input type="text" class="comp-start" placeholder="Start command" value="${escapeAttr(component.startCommand)}" />
    <input type="text" class="comp-stop" placeholder="Stop command" value="${escapeAttr(component.stopCommand)}" />
    <select class="comp-runmode">
      <option value="native">Native</option>
      <option value="docker">Docker</option>
    </select>
    <button class="remove-component-btn" title="Remove this component">✕</button>
  `;

  row.querySelector(".comp-runmode").value = component.runMode || "native";
  row.querySelector(".remove-component-btn").addEventListener("click", () => row.remove());

  containerEl.appendChild(row);
}

// A small folder browser reused for the edit page's "change folder" flow —
// same backend endpoint as onboarding's picker, but writes into editState
// instead of the global addAppState so the two never conflict.
async function renderEditFolderPicker(wrapper, editState, path) {
  const pickerEl = wrapper.querySelector("#edit-folder-picker");
  pickerEl.innerHTML = "Loading...";

  const res = await fetch(`/api/browse?path=${encodeURIComponent(path || "")}`);
  if (!res.ok) {
    pickerEl.innerHTML = "Could not browse this location.";
    return;
  }
  const listing = await res.json();

  pickerEl.innerHTML = "";

  const currentRow = document.createElement("div");
  currentRow.className = "folder-current-row";
  currentRow.innerHTML = `<span>${listing.currentPath}</span>`;
  const selectBtn = document.createElement("button");
  selectBtn.textContent = "Select this folder";
  selectBtn.addEventListener("click", () => {
    editState.selectedPath = listing.currentPath;
    wrapper.querySelector("#current-path-display").textContent = listing.currentPath;
    pickerEl.parentElement.style.display = "none";
    editState.browsing = false;
  });
  currentRow.appendChild(selectBtn);
  pickerEl.appendChild(currentRow);

  const list = document.createElement("div");
  list.className = "folder-list";

  if (listing.parentPath) {
    const upRow = document.createElement("div");
    upRow.className = "folder-row";
    upRow.textContent = ".. (up)";
    upRow.addEventListener("click", () => renderEditFolderPicker(wrapper, editState, listing.parentPath));
    list.appendChild(upRow);
  }

  (listing.directories || []).forEach((dir) => {
    const row = document.createElement("div");
    row.className = "folder-row";
    row.textContent = dir.name;
    row.addEventListener("click", () => renderEditFolderPicker(wrapper, editState, dir.path));
    list.appendChild(row);
  });

  pickerEl.appendChild(list);
}

// --- Placeholder pages for nav items not yet built ---

function renderBrainstormPage(container) {
  container.innerHTML = `
    <div class="placeholder-page">
      <h2>Brainstorm</h2>
      <p>Idea space (tree-based seed-to-app view) — planned for v1.1, not yet built.</p>
    </div>
  `;
}

function renderActivityPage(container) {
  container.innerHTML = `
    <div class="placeholder-page">
      <h2>Activity</h2>
      <p>Recent start/stop and git activity across apps — not yet built.</p>
    </div>
  `;
}

// --- Init ---

window.addEventListener("DOMContentLoaded", () => navigateTo("registry"));