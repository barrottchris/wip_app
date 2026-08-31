import { listArchivedApps, unarchiveApp } from "../api.js";
import { refreshCurrentPage } from "../router.js";

// Archived view — same data shape as the registry, but apps here are
// hidden from the main list (see next-phase-plan.md: hidden by default,
// dedicated Archived view, not just greyed out inline). Cards are rendered
// simply here rather than reusing renderAppCard's full start/stop/pill
// treatment, since an archived app has one relevant action: unarchive it.
export async function renderArchivedPage(container) {
  container.innerHTML = "Loading...";
  try {
    const apps = await listArchivedApps();
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
    await unarchiveApp(entry.id);
    refreshCurrentPage();
  });
  card.appendChild(unarchiveBtn);

  return card;
}