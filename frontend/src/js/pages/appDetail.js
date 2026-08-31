import { archiveApp, getApp, updateApp, updateComponents } from "../api.js";
import { renderConnectionPill } from "../components/appCard.js";
import { addComponentRow, readComponentsFromDOM } from "../components/componentEditor.js";
import { renderFolderPicker } from "../components/folderPicker.js";
import { navigateTo, openGitPage, selectedAppId } from "../router.js";
import { escapeAttr } from "../utils.js";

export async function renderAppDetailPage(container) {
  if (!selectedAppId) {
    container.innerHTML = `<p>No app selected. <a href="#" id="back-to-registry">Back to registry</a></p>`;
    container.querySelector("#back-to-registry").addEventListener("click", (e) => {
      e.preventDefault();
      navigateTo("registry");
    });
    return;
  }

  container.innerHTML = "Loading...";
  let entry;
  try {
    entry = await getApp(selectedAppId);
  } catch (err) {
    container.innerHTML = `<p>Could not load app.</p>`;
    return;
  }

  container.innerHTML = "";
  const editState = { selectedPath: entry.localPath, browsing: false };

  const wrapper = document.createElement("div");
  wrapper.className = "add-app-page";
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

    <div id="components-section" class="detail-section">
      <h2>Components</h2>
      <p class="hint">Each component can have multiple build steps, a single run command, and a stop command.</p>
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

  const componentRowsEl = wrapper.querySelector("#component-rows");
  (entry.components || []).forEach((component) => addComponentRow(componentRowsEl, component));

  wrapper.querySelector("#add-component-btn").addEventListener("click", () => {
    addComponentRow(componentRowsEl, { name: "", buildCommands: [], runCommand: "", stopCommand: "", runMode: "native" });
  });

  wrapper.querySelector("#save-components-btn").addEventListener("click", async () => {
    const components = readComponentsFromDOM(componentRowsEl);
    const statusMsgEl = wrapper.querySelector("#components-status-msg");

    if (components.some((c) => !c.name)) {
      statusMsgEl.textContent = "Every component needs a name.";
      return;
    }

    const res = await updateComponents(entry.id, components);
    statusMsgEl.textContent = res.ok ? "Saved." : `Error: ${await res.text()}`;
  });

  wrapper.querySelector("#edit-status").value = entry.status;

  const pillsEl = wrapper.querySelector(".detail-pills");
  pillsEl.appendChild(renderConnectionPill("Git", entry.gitConnected, false, () => openGitPage(entry.id)));
  pillsEl.appendChild(renderConnectionPill("Jira", entry.jiraConnected, entry.jiraComingSoon));
  pillsEl.appendChild(renderConnectionPill("Confluence", entry.confluenceConnected, entry.confluenceComingSoon));

  wrapper.querySelector("#change-folder-btn").addEventListener("click", async () => {
    editState.browsing = !editState.browsing;
    const pickerEl = wrapper.querySelector("#edit-folder-picker");
    pickerEl.style.display = editState.browsing ? "block" : "none";
    if (editState.browsing) {
      await renderFolderPicker(pickerEl, editState.selectedPath, {
        selectButtonLabel: "Select this folder",
        onSelect: (path) => {
          editState.selectedPath = path;
          wrapper.querySelector("#current-path-display").textContent = path;
          pickerEl.style.display = "none";
          editState.browsing = false;
        },
      });
    }
  });

  wrapper.querySelector("#cancel-edit-btn").addEventListener("click", () => navigateTo("registry"));

  wrapper.querySelector("#archive-btn").addEventListener("click", async () => {
    const confirmed = confirm(`Archive "${entry.name}"? It will be hidden from the registry but can be restored from the Archived view.`);
    if (!confirmed) return;

    const alsoDelete = confirm("Also delete the folder from disk? This cannot be undone.");
    await archiveApp(entry.id, alsoDelete);
    navigateTo("registry");
  });

  wrapper.querySelector("#save-edit-btn").addEventListener("click", async () => {
    const name = wrapper.querySelector("#edit-name").value.trim();
    const description = wrapper.querySelector("#edit-description").value.trim();
    const status = wrapper.querySelector("#edit-status").value;
    if (!name || !editState.selectedPath) return;

    const putRes = await updateApp(entry.id, { name, description, localPath: editState.selectedPath, status });

    if (putRes.ok) {
      navigateTo("registry");
    } else {
      const errText = await putRes.text();
      wrapper.querySelector("#edit-status-msg").textContent = `Error: ${errText}`;
    }
  });
}