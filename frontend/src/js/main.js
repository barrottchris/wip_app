import { renderAddAppPage } from "./pages/addApp.js";
import { renderAppDetailPage } from "./pages/appDetail.js";
import { renderAppGitPage } from "./pages/appGit.js";
import { renderArchivedPage } from "./pages/archived.js";
import { renderBrainstormPage, renderActivityPage } from "./pages/placeholders.js";
import { renderRegistryPage } from "./pages/registry.js";
import { renderSettingsPage } from "./pages/settings.js";
import { navigateTo, setPages } from "./router.js";

const pages = {
  registry: renderRegistryPage,
  archived: renderArchivedPage,
  brainstorm: renderBrainstormPage,
  activity: renderActivityPage,
  settings: renderSettingsPage,
  "add-app": renderAddAppPage,
  "app-detail": renderAppDetailPage,
  "app-git": renderAppGitPage,
};

setPages(pages);

const addAppBtn = document.getElementById("add-app-btn");
if (addAppBtn) {
  addAppBtn.addEventListener("click", () => navigateTo("add-app"));
}

document.querySelectorAll(".nav-item").forEach((el) => {
  el.addEventListener("click", () => navigateTo(el.dataset.page));
});

window.addEventListener("DOMContentLoaded", () => navigateTo("registry"));
