import { escapeAttr } from "../utils.js";

// addComponentRow builds one editable component block — name, an
// add/remove list of build commands, a single run command, a stop command,
// and run mode — each field on its own line rather than crammed into one
// row. Used both for existing components (pre-filled) and the "+ Add
// component" button (blank block).
export function addComponentRow(containerEl, component) {
  const block = document.createElement("div");
  block.className = "component-edit-block";

  block.innerHTML = `
    <label>Component name</label>
    <input type="text" class="comp-name" placeholder="e.g. Frontend" value="${escapeAttr(component.name)}" />

    <label>Build commands</label>
    <div class="build-commands-list"></div>
    <button type="button" class="add-build-cmd-btn">+ Add build command</button>

    <label>Run command</label>
    <input type="text" class="comp-run" placeholder="Command that starts and keeps this running" value="${escapeAttr(component.runCommand)}" />

    <label>Stop command</label>
    <input type="text" class="comp-stop" placeholder="Command to stop it" value="${escapeAttr(component.stopCommand)}" />

    <label>Run mode</label>
    <select class="comp-runmode">
      <option value="native">Native</option>
      <option value="docker">Docker</option>
    </select>

    <button type="button" class="remove-component-btn">Remove component</button>
  `;

  block.querySelector(".comp-runmode").value = component.runMode || "native";

  const buildListEl = block.querySelector(".build-commands-list");
  (component.buildCommands || []).forEach((cmd) => addBuildCommandRow(buildListEl, cmd));

  block.querySelector(".add-build-cmd-btn").addEventListener("click", () => {
    addBuildCommandRow(buildListEl, "");
  });

  block.querySelector(".remove-component-btn").addEventListener("click", () => block.remove());

  containerEl.appendChild(block);
}

// addBuildCommandRow adds one build-command input (with its own remove
// button) to a component's build-commands list.
function addBuildCommandRow(listEl, value) {
  const row = document.createElement("div");
  row.className = "build-cmd-row";
  row.innerHTML = `
    <input type="text" class="build-cmd-input" placeholder="e.g. npm install" value="${escapeAttr(value)}" />
    <button type="button" class="remove-build-cmd-btn" title="Remove this build command">✕</button>
  `;
  row.querySelector(".remove-build-cmd-btn").addEventListener("click", () => row.remove());
  listEl.appendChild(row);
}

// readComponentsFromDOM collects the current state of every component
// block back into plain objects — used when saving (see pages/appDetail.js).
export function readComponentsFromDOM(containerEl) {
  return Array.from(containerEl.querySelectorAll(".component-edit-block")).map((block) => ({
    name: block.querySelector(".comp-name").value.trim(),
    buildCommands: Array.from(block.querySelectorAll(".build-cmd-input"))
      .map((i) => i.value.trim())
      .filter((v) => v),
    runCommand: block.querySelector(".comp-run").value.trim(),
    stopCommand: block.querySelector(".comp-stop").value.trim(),
    runMode: block.querySelector(".comp-runmode").value,
  }));
}