import { getApp, gitInit } from "../api.js";
import { navigateTo, selectedAppId } from "../router.js";

export async function renderAppGitPage(container) {
  if (!selectedAppId) {
    container.innerHTML = `<p>No app selected. <a href="#" id="back-to-registry">Back to registry</a></p>`;
    container.querySelector("#back-to-registry").addEventListener("click", (e) => {
      e.preventDefault();
      navigateTo("registry");
    });
    return;
  }

  container.innerHTML = "Loading git info...";
  try {
    const entry = await getApp(selectedAppId);
    container.innerHTML = "";

    const wrapper = document.createElement("div");
    wrapper.className = "add-app-page";
    wrapper.innerHTML = `
      <h2>Git: ${entry.name}</h2>
      <p class="hint">${entry.gitConnected ? "This folder is already under git." : "This folder does not have a git repo yet."}</p>
      <p><strong>Folder:</strong> ${entry.localPath || "—"}</p>
      <p><strong>Default branch:</strong> ${entry.defaultBranch || "—"}</p>
      <p><strong>Last touched:</strong> ${entry.lastTouchedAt ? new Date(entry.lastTouchedAt).toLocaleString() : "—"}</p>
      <div class="detail-section">
        <button id="git-init-btn" ${entry.gitConnected ? "style=\"display:none;\"" : ""}>Initialize git here</button>
        <button id="back-to-app-btn">Back to app</button>
        <button id="back-to-registry-btn">Back to registry</button>
      </div>
      <div id="git-status-msg" class="hint"></div>
    `;

    container.appendChild(wrapper);

    wrapper.querySelector("#git-init-btn")?.addEventListener("click", async () => {
      const res = await gitInit(entry.localPath);
      const msg = wrapper.querySelector("#git-status-msg");
      if (!res.ok) {
        msg.textContent = `Git init failed: ${await res.text()}`;
        return;
      }
      msg.textContent = "Git initialized successfully.";
      navigateTo("app-git");
    });

    wrapper.querySelector("#back-to-app-btn").addEventListener("click", () => {
      navigateTo("app-detail");
    });

    wrapper.querySelector("#back-to-registry-btn").addEventListener("click", () => {
      navigateTo("registry");
    });
  } catch (err) {
    container.innerHTML = `<p>Could not load git info: ${err}</p>`;
  }
}