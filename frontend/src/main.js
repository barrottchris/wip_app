// App shell: top banner + left nav + a simple client-side router that swaps
// content in #page-content. No framework — plain DOM manipulation, matching
// the rest of this scaffold's "no build step" approach.

const pages = {
  registry: renderRegistryPage,
  brainstorm: renderBrainstormPage,
  activity: renderActivityPage,
  settings: renderSettingsPage,
};

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

  card.innerHTML = `
    <div class="app-card-header">
      <span class="app-name">${entry.name}</span>
      <span class="app-status">${entry.status}</span>
    </div>
    <p class="app-description">${entry.description || ""}</p>
    <div class="app-meta">
      <span>branch: ${branch}</span>
      <span>last touched: ${lastTouched}</span>
    </div>
    <div class="app-components"></div>
  `;

  const componentsEl = card.querySelector(".app-components");
  (entry.components || []).forEach((component) => {
    componentsEl.appendChild(renderComponentRow(entry.id, component));
  });

  return card;
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
