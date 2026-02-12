const apiBase = "/api/v1";

const state = {
  containers: [],
  images: [],
};

async function fetchJSON(url, options = {}) {
  const res = await fetch(url, options);
  if (!res.ok) {
    const payload = await res.json().catch(() => ({ error: "request failed" }));
    throw new Error(payload.error || "request failed");
  }
  return res.json();
}

function renderStats(d) {
  const stats = document.getElementById("stats");
  stats.innerHTML = `
    <div class="stat-card"><div class="stat-title">Containers</div><div class="stat-value">${d.containersTotal}</div></div>
    <div class="stat-card"><div class="stat-title">Running</div><div class="stat-value">${d.containersUp}</div></div>
    <div class="stat-card"><div class="stat-title">Stopped</div><div class="stat-value">${d.containersDown}</div></div>
    <div class="stat-card"><div class="stat-title">Images</div><div class="stat-value">${d.imagesTotal}</div></div>
  `;
  const mode = document.getElementById("modeBadge");
  mode.textContent = d.simulated ? "Simulation Mode" : "Docker Live Mode";
}

function renderContainers() {
  const body = document.getElementById("containersBody");
  body.innerHTML = state.containers
    .map(
      (c) => `
      <tr>
        <td>${c.id}</td>
        <td>${c.name}</td>
        <td>${c.image}</td>
        <td><span class="status ${c.state}">${c.status}</span></td>
        <td>
          <div class="row-actions">
            <button onclick="actionContainer('${c.id}','start')">Start</button>
            <button class="danger" onclick="actionContainer('${c.id}','stop')">Stop</button>
            <button onclick="actionContainer('${c.id}','restart')">Restart</button>
          </div>
        </td>
      </tr>
    `,
    )
    .join("");
}

function renderImages() {
  const body = document.getElementById("imagesBody");
  body.innerHTML = state.images
    .map(
      (img) => `
      <tr>
        <td>${img.id}</td>
        <td>${img.repoTags}</td>
        <td>${img.sizeMB}</td>
      </tr>
    `,
    )
    .join("");
}

async function loadDashboard() {
  const d = await fetchJSON(`${apiBase}/dashboard`);
  renderStats(d);
}

async function loadContainers() {
  state.containers = await fetchJSON(`${apiBase}/containers`);
  renderContainers();
}

async function loadImages() {
  state.images = await fetchJSON(`${apiBase}/images`);
  renderImages();
}

async function actionContainer(id, action) {
  try {
    await fetchJSON(`${apiBase}/containers/${id}/${action}`, { method: "POST" });
    await Promise.all([loadDashboard(), loadContainers()]);
  } catch (err) {
    alert(err.message);
  }
}

function setupTabs() {
  document.querySelectorAll(".nav-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      document.querySelectorAll(".nav-btn").forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      const tab = btn.dataset.tab;
      document.getElementById("containersTab").classList.toggle("hidden", tab !== "containers");
      document.getElementById("imagesTab").classList.toggle("hidden", tab !== "images");
    });
  });
}

function setupEvents() {
  document.getElementById("refreshContainers").addEventListener("click", async () => {
    await Promise.all([loadDashboard(), loadContainers()]);
  });
  document.getElementById("refreshImages").addEventListener("click", loadImages);
}

async function init() {
  setupTabs();
  setupEvents();
  await Promise.all([loadDashboard(), loadContainers(), loadImages()]);
}

init().catch((err) => {
  alert(`Failed to initialize: ${err.message}`);
});
