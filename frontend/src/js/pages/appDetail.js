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

    <div id="folder-section" class="detail-section folder-panel">
      <h2>Folder</h2>
      <div class="selected-folder-summary">
        <span class="selected-folder-label">Current path</span>
        <div id="current-path-display" class="selected-path-hint">${escapeAttr(entry.localPath)}</div>
      </div>
      <button id="change-folder-btn">Change folder</button>
      <div id="edit-folder-picker" style="display:none; margin-top:0.75rem;"></div>
    </div>

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

    <div id="archive-confirm" class="archive-confirm" style="display:none;">
      <p>Archive "${escapeAttr(entry.name)}"? It will be hidden from the registry but can be restored from the Archived view.</p>
      <label class="archive-delete-option">
        <input type="checkbox" id="archive-delete-folder-checkbox" />
        Also delete the folder from disk. This cannot be undone.
      </label>
      <div class="archive-confirm-actions">
        <button id="archive-confirm-btn" class="danger-btn">Archive app</button>
        <button id="archive-cancel-btn" type="button">Cancel</button>
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
    addComponentRow(componentRowsEl, { name: "", buildCommands: [], startCommand: "", stopCommand: "", runMode: "native" });
  });

  wrapper.querySelector("#save-components-btn").addEventListener("click", async () => {
    const components = readComponentsFromDOM(componentRowsEl);
    const statusMsgEl = wrapper.querySelector("#components-status-msg");

    if (components.some((c) => !c.name)) {
      statusMsgEl.className = "status-pill status-error";
      statusMsgEl.textContent = "Every component needs a name.";
      return;
    }

    const res = await updateComponents(entry.id, components);
    if (res.ok) {
      statusMsgEl.className = "status-pill status-success";
      statusMsgEl.textContent = "Saved";
    } else {
      statusMsgEl.className = "status-pill status-error";
      statusMsgEl.textContent = `Error: ${await res.text()}`;
    }
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

  wrapper.querySelector("#archive-btn").addEventListener("click", () => {
    const confirmEl = wrapper.querySelector("#archive-confirm");
    confirmEl.style.display = "block";
  });

  wrapper.querySelector("#archive-cancel-btn").addEventListener("click", () => {
    const confirmEl = wrapper.querySelector("#archive-confirm");
    confirmEl.style.display = "none";
  });

  wrapper.querySelector("#archive-confirm-btn").addEventListener("click", async () => {
    const confirmEl = wrapper.querySelector("#archive-confirm");
    const alsoDelete = wrapper.querySelector("#archive-delete-folder-checkbox").checked;
    confirmEl.style.display = "none";
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
      const statusEl = wrapper.querySelector("#edit-status-msg");
      statusEl.className = "status-pill status-success";
      statusEl.textContent = "Saved";
      setTimeout(() => navigateTo("registry"), 350);
    } else {
      const errText = await putRes.text();
      const statusEl = wrapper.querySelector("#edit-status-msg");
      statusEl.className = "status-pill status-error";
      statusEl.textContent = `Error: ${errText}`;
    }
  });
}