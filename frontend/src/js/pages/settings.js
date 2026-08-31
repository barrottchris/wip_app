import { getSettings, updateSettings } from "../api.js";

export async function renderSettingsPage(container) {
  container.innerHTML = "Loading settings...";
  try {
    const settings = await getSettings();
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
    await updateSettings(body);
    wrapper.querySelector("#save-status").textContent = "Saved.";
  };

  return wrapper;
}