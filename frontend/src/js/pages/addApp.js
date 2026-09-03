import { createApp, gitInit, gitStatus } from "../api.js";
import { navigateTo } from "../router.js";
import { renderFolderPicker } from "../components/folderPicker.js";

// Local state for the in-progress add-app form, reset each time the page
// is rendered. Kept simple/module-scoped since this is a single-page-at-a-
// time hand-rolled router, not a component framework.
let addAppState = {
  mode: "existing", // "existing" | "new"
  selectedPath: null,
  browsePath: "",
};

export async function renderAddAppPage(container) {
  addAppState = { mode: "existing", selectedPath: null, browsePath: "" };

  const wrapper = document.createElement("div");
  wrapper.className = "add-app-page";
  wrapper.innerHTML = `
    <h2>Add app</h2>

    <div class="mode-toggle">
      <button class="mode-btn active" data-mode="existing">Existing folder</button>
      <button class="mode-btn" data-mode="new">Create new</button>
    </div>

    <label>Name</label>
    <input type="text" id="new-app-name" placeholder="My App" />

    <label>Description</label>
    <input type="text" id="new-app-description" placeholder="What is this app?" />

    <div class="folder-selection-panel">
      <div class="folder-selection-header">
        <div>
          <h3>Folder selection</h3>
          <p>Choose the folder this app will live in.</p>
        </div>
      </div>

      <div class="selected-folder-summary">
        <span class="selected-folder-label">Selected folder</span>
        <div id="selected-path-hint" class="selected-path-hint">Not selected yet</div>
      </div>

      <div class="folder-picker-shell">
        <label id="folder-label">Choose an existing folder</label>
        <div class="folder-path-search">
          <input type="text" id="folder-path-input" placeholder="C:\\path\\to\\folder" />
          <button type="button" id="find-folder-btn">Find folder</button>
        </div>
        <p id="folder-search-status" class="hint"></p>
        <div id="folder-picker"></div>
      </div>
    </div>

    <div id="git-prompt" style="display:none;" class="git-prompt">
      <p>This folder isn't under git yet. Initialize it?</p>
      <button id="git-init-yes">Yes, run git init</button>
      <button id="git-init-skip">Skip for now</button>
    </div>

    <button id="create-app-btn" disabled>Add app</button>
    <span id="create-status" class="hint"></span>
  `;
  container.appendChild(wrapper);

  const pickerEl = wrapper.querySelector("#folder-picker");
  const pickerOptions = {
    get selectButtonLabel() {
      return addAppState.mode === "existing" ? "Use this folder" : "Create here";
    },
    onBrowse: (path) => {
      addAppState.browsePath = path;
      wrapper.querySelector("#folder-path-input").value = path;
      wrapper.querySelector("#folder-search-status").textContent = "";
    },
    onSelect: (path) => selectFolder(wrapper, path),
  };

  wrapper.querySelectorAll(".mode-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      wrapper.querySelectorAll(".mode-btn").forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      addAppState.mode = btn.dataset.mode;
      addAppState.selectedPath = null;
      wrapper.querySelector("#folder-label").textContent =
        addAppState.mode === "existing"
          ? "Choose an existing folder"
          : "Choose where to create the new folder";
      wrapper.querySelector("#selected-path-hint").textContent = "";
      wrapper.querySelector("#folder-path-input").value = addAppState.browsePath;
      wrapper.querySelector("#folder-search-status").textContent = "";
      wrapper.querySelector("#git-prompt").style.display = "none";
      updateCreateButtonState(wrapper);
      renderFolderPicker(pickerEl, addAppState.browsePath, pickerOptions);
    });
  });

  wrapper.querySelector("#new-app-name").addEventListener("input", () => updateCreateButtonState(wrapper));

  wrapper.querySelector("#find-folder-btn").addEventListener("click", async () => {
    const path = wrapper.querySelector("#folder-path-input").value.trim();
    const statusEl = wrapper.querySelector("#folder-search-status");
    if (!path) {
      statusEl.className = "hint folder-search-error";
      statusEl.textContent = "Enter a folder path first.";
      return;
    }

    statusEl.className = "hint";
    statusEl.textContent = "Checking folder...";
    await renderFolderPicker(pickerEl, path, pickerOptions);
    if (pickerEl.textContent === "Could not browse this location.") {
      statusEl.className = "hint folder-search-error";
      statusEl.textContent = "Folder not found or unavailable.";
    }
  });

  wrapper.querySelector("#git-init-yes").addEventListener("click", async () => {
    await gitInit(addAppState.selectedPath);
    wrapper.querySelector("#git-prompt").style.display = "none";
  });
  wrapper.querySelector("#git-init-skip").addEventListener("click", () => {
    wrapper.querySelector("#git-prompt").style.display = "none";
  });

  wrapper.querySelector("#create-app-btn").addEventListener("click", async () => {
    const name = wrapper.querySelector("#new-app-name").value.trim();
    const description = wrapper.querySelector("#new-app-description").value.trim();
    if (!name || !addAppState.selectedPath) return;

    const res = await createApp({
      name,
      description,
      localPath: addAppState.selectedPath,
      createNew: addAppState.mode === "new",
    });

    if (res.ok) {
      navigateTo("registry");
    } else {
      const errText = await res.text();
      wrapper.querySelector("#create-status").textContent = `Error: ${errText}`;
    }
  });

  await renderFolderPicker(pickerEl, "", pickerOptions);
}

function updateCreateButtonState(wrapper) {
  const name = wrapper.querySelector("#new-app-name").value.trim();
  const btn = wrapper.querySelector("#create-app-btn");
  btn.disabled = !(name && addAppState.selectedPath);
}

async function selectFolder(wrapper, path) {
  addAppState.selectedPath = path;
  wrapper.querySelector("#selected-path-hint").textContent = path;
  updateCreateButtonState(wrapper);

  if (addAppState.mode === "existing") {
    try {
      const { hasGit } = await gitStatus(path);
      wrapper.querySelector("#git-prompt").style.display = hasGit ? "none" : "block";
    } catch (err) {
      // Non-critical — leave the prompt hidden if the check itself fails.
    }
  } else {
    wrapper.querySelector("#git-prompt").style.display = "none";
  }
}