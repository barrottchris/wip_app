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

    const currentPath = document.createElement("span");
    currentPath.className = "folder-current-path";
    currentPath.textContent = listing.currentPath || "Current folder";
    currentRow.appendChild(currentPath);

    const selectBtn = document.createElement("button");
    selectBtn.className = "folder-select-btn";
    selectBtn.textContent = selectButtonLabel;
    selectBtn.addEventListener("click", () => onSelect?.(listing.currentPath));
    currentRow.appendChild(selectBtn);
    containerEl.appendChild(currentRow);

    const list = document.createElement("div");
    list.className = "folder-list";

    if (listing.parentPath) {
      const upButton = document.createElement("button");
      upButton.type = "button";
      upButton.className = "folder-nav-btn";
      upButton.innerHTML = '<span class="folder-nav-icon" aria-hidden="true">↑</span><span>Up a level</span>';
      upButton.addEventListener("click", () => renderFolderPicker(containerEl, listing.parentPath, options));
      list.appendChild(upButton);
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