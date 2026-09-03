import { listApps, saveRegistryOrder } from "../api.js";
import { renderAppCard } from "../components/appCard.js";
import { refreshCurrentPage } from "../router.js";
import { runtimeChangedEvent } from "../runtimeEvents.js";

const runtimePollIntervalMs = 2000;

async function refreshRunningCount() {
  try {
    updateRunningCount(await listApps());
  } catch (err) {
    // Keep the last known count when a background refresh cannot reach the backend.
  }
}

window.addEventListener(runtimeChangedEvent, refreshRunningCount);
setInterval(refreshRunningCount, runtimePollIntervalMs);

export async function renderRegistryPage(container) {
  container.innerHTML = "Loading...";
  try {
    const apps = await listApps();
    container.innerHTML = "";
    container.classList.add("registry-list");
    updateRunningCount(apps);
    apps.forEach((entry) => {
      const card = renderAppCard(entry);
      card.draggable = true;
      card.addEventListener("dragstart", () => {
        card.classList.add("is-dragging");
      });
      card.addEventListener("dragend", () => {
        card.classList.remove("is-dragging");
        container.querySelectorAll(".is-drag-over").forEach((item) => item.classList.remove("is-drag-over"));
      });
      card.addEventListener("dragover", (event) => {
        event.preventDefault();
        const dragging = container.querySelector(".is-dragging");
        if (!dragging || dragging === card) return;
        const after = event.clientY > card.getBoundingClientRect().top + card.offsetHeight / 2;
        animateReorder(container, dragging, after ? card.nextSibling : card);
        container.querySelectorAll(".is-drag-over").forEach((item) => item.classList.remove("is-drag-over"));
        card.classList.add("is-drag-over");
      });
      card.addEventListener("drop", async (event) => {
        event.preventDefault();
        card.classList.remove("is-drag-over");
        const ids = [...container.querySelectorAll(".app-card")].map((item) => item.dataset.appId);
        const response = await saveRegistryOrder(ids);
        if (!response.ok) {
          alert(`Could not save registry order: ${await response.text()}`);
          await renderRegistryPage(container);
        }
      });
      container.appendChild(card);
    });
  } catch (err) {
    container.innerHTML = `<p>Error loading apps: ${err}</p>`;
  }
}

function animateReorder(container, dragging, anchor) {
  const before = new Map([...container.querySelectorAll(".app-card")].map((item) => [item, item.getBoundingClientRect().top]));
  if (anchor === dragging || anchor?.previousElementSibling === dragging) return;
  container.insertBefore(dragging, anchor);
  for (const item of container.querySelectorAll(".app-card")) {
    if (item === dragging) continue;
    const delta = before.get(item) - item.getBoundingClientRect().top;
    if (!delta) continue;
    item.style.transform = `translateY(${delta}px)`;
    requestAnimationFrame(() => {
      item.style.transform = "";
    });
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
