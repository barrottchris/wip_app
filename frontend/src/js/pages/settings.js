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
    <p class="hint">
      Optional setup for future GitHub integrations. WIP uses this information to identify
      your GitHub account and to authenticate repo-related actions such as checking status,
      creating/connecting repos, and other GitHub workflows later on.
    </p>
    <ol class="hint" style="margin:0.5rem 0 1rem 1.25rem; padding-left:1rem;">
      <li>Open GitHub and create a personal access token with at least <strong>repo</strong> access.</li>
      <li>Paste your GitHub username here.</li>
      <li>Paste the token in the field below. It is stored securely and never shown again.</li>
      <li>Leave both blank if you are not using GitHub features yet.</li>
    </ol>
    <label>Username</label>
    <input type="text" id="github-username" value="${settings.githubUsername || ""}" />
    <label>Personal access token</label>
    <input type="password" id="github-token" placeholder="${settings.githubTokenIsSet ? "•••••••• (already set)" : "Not set"}" />
    <p class="hint">This is only for GitHub connectivity and repo access. Leave it blank if you are not setting up GitHub integration yet.</p>

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