// Router holds only navigation primitives. It deliberately does NOT import
// the page modules itself — main.js does that and calls setPages() once at
// startup. If router.js imported pages/*.js directly, and pages/*.js also
// need to import navigateTo from here, that's a circular import; this
// avoids it entirely.

let pages = {};
let currentPageName = null;
let registryPollTimer = null;

export function setPages(pageMap) {
  pages = pageMap;
}

// Which app the detail/git/edit pages are currently showing — set by
// openAppDetail()/openGitPage() before navigating there. Simple module-level
// state, matching this scaffold's "no framework" approach.
export let selectedAppId = null;

export function navigateTo(pageName) {
  if (registryPollTimer) {
    clearInterval(registryPollTimer);
    registryPollTimer = null;
  }

  document.querySelectorAll(".nav-item").forEach((el) => {
    el.classList.toggle("active", el.dataset.page === pageName);
  });
  currentPageName = pageName;
  const content = document.getElementById("page-content");
  content.innerHTML = "";
  const renderFn = pages[pageName] || pages["registry"];
  renderFn(content);

  if (pageName === "registry") {
    registryPollTimer = setInterval(() => {
      if (currentPageName === "registry") {
        refreshCurrentPage();
      }
    }, 2000);
  }
}

// Re-render whatever page is currently showing — used after an action
// (start/stop, git refresh, archive) changes data the current page
// displays, without needing every caller to know which page it's on.
export function refreshCurrentPage() {
  if (currentPageName) navigateTo(currentPageName);
}

export function openAppDetail(appId) {
  selectedAppId = appId;
  navigateTo("app-detail");
}

export function openGitPage(appId) {
  selectedAppId = appId;
  navigateTo("app-git");
}