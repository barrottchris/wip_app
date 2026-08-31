import { listApps } from "../api.js";
import { renderAppCard } from "../components/appCard.js";
import { refreshCurrentPage } from "../router.js";

export async function renderRegistryPage(container) {
  container.innerHTML = "Loading...";
  try {
    const apps = await listApps();
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

export async function refreshRegistryPage(container) {
  await renderRegistryPage(container);
  refreshCurrentPage();
}
