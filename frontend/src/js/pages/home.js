import { navigateTo } from "../router.js";

export function renderHomePage(container) {
  container.innerHTML = `
    <div class="home-page">
      <div class="home-hero">
        <section class="home-intro">
          <p class="home-kicker">A place for work in progress</p>
          <h1>Your Work-in-Progress,<br /><span>Managed in One Place.</span></h1>
          <p class="home-summary">
            WIP keeps your half-built apps findable, runnable, and moving forward.
            See what exists, resume it without hunting for commands, and keep the work
            that matters in one managed place.
          </p>
          <div class="home-actions">
            <button id="home-registry-btn" class="home-primary-btn" type="button">Open registry</button>
            <button id="home-add-app-btn" type="button">Add an app</button>
          </div>
        </section>

        <section class="home-visual" data-image-slot="future-problem-illustration" aria-label="CSS illustration of scattered projects becoming one organized workspace">
          <div class="visual-caption">Many projects. One place to pick up.</div>
          <div class="visual-board visual-board-left"><span>notes</span><span>copy-final</span><span>maybe-this-one</span></div>
          <div class="visual-board visual-board-top"><span>start again?</span><span>localhost:3000</span></div>
          <div class="visual-board visual-board-bottom"><span>branch: wip</span><span>TODO / TODO / TODO</span></div>
          <div class="visual-thread thread-one"></div>
          <div class="visual-thread thread-two"></div>
          <div class="visual-thread thread-three"></div>
          <div class="visual-hub"><strong>WIP</strong><span>pick up where<br />you left off</span></div>
        </section>
      </div>

      <section class="home-principles" aria-label="What WIP keeps together">
        <article><strong>Find it</strong><span>One registry for the projects that are still alive.</span></article>
        <article><strong>Run it</strong><span>Start and stop each app without interrupting the others.</span></article>
        <article><strong>Return to it</strong><span>Git state and commands stay close to the work.</span></article>
      </section>

      <p class="home-roadmap"><span>Planned connections</span> GitHub <b>+</b> Jira <b>+</b> Confluence</p>
    </div>
  `;

  container.querySelector("#home-registry-btn").addEventListener("click", () => navigateTo("registry"));
  container.querySelector("#home-add-app-btn").addEventListener("click", () => navigateTo("add-app"));
}