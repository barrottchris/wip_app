import { browse } from "../api.js";

export async function renderFolderPicker(containerEl, path, options = {}) {
  const {
    selectButtonLabel = "Select this folder",
    onBrowse,
    onSelect,
  } = options;

  containerEl.innerHTML = "Loading...";
  try {
    const listing = await browse(path || "");
    onBrowse?.(listing.currentPath);

    containerEl.innerHTML = "";

    const currentRow = document.createElement("div");
    currentRow.className = "folder-current-row";
    currentRow.innerHTML = `<span>${listing.currentPath}</span>`;

    const selectBtn = document.createElement("button");
    selectBtn.textContent = selectButtonLabel;
    selectBtn.addEventListener("click", () => onSelect?.(listing.currentPath));
    currentRow.appendChild(selectBtn);
    containerEl.appendChild(currentRow);

    const list = document.createElement("div");
    list.className = "folder-list";

    if (listing.parentPath) {
      const upRow = document.createElement("div");
      upRow.className = "folder-row";
      upRow.textContent = ".. (up)";
      upRow.addEventListener("click", () => renderFolderPicker(containerEl, listing.parentPath, options));
      list.appendChild(upRow);
    }

    (listing.directories || []).forEach((dir) => {
      const row = document.createElement("div");
      row.className = "folder-row";
      row.textContent = dir.name;
      row.addEventListener("click", () => renderFolderPicker(containerEl, dir.path, options));
      list.appendChild(row);
    });

    containerEl.appendChild(list);
  } catch (err) {
    containerEl.innerHTML = "Could not browse this location.";
  }
}