const FETCH_INTERVAL = 10 * 60 * 1000;
const eventLog = document.getElementById("eventLog");
const updateAvailableText = document.getElementById("update-available-text");

async function fetchLatestVersion() {
  const repoUrl = "https://api.github.com/repos/pommee/goaway/releases";
  const lastFetched = localStorage.getItem("lastFetched");
  const now = new Date().getTime();

  if (lastFetched && now - lastFetched < FETCH_INTERVAL) {
    return;
  }

  try {
    const response = await fetch(repoUrl);
    if (!response.ok) {
      throw new Error(`Failed to fetch releases: ${response.statusText}`);
    }
    const releases = await response.json();

    if (releases.length > 0) {
      const latestVersion = releases[0].name.replace("v", "");
      localStorage.setItem("latestVersion", latestVersion);
      localStorage.setItem("lastFetched", now);
    }
  } catch (error) {
    console.error("Error fetching latest version:", error);
  }
}

function updateVersionText() {
  const currentVersion = GetInstalledVersion();
  const latestVersion = GetLatestVersion();

  if (latestVersion) {
    updateAvailableText.innerHTML = `${currentVersion} -> ${latestVersion}`;
  }
}

function startFetchInterval() {
  fetchLatestVersion();
  setInterval(fetchLatestVersion, FETCH_INTERVAL);
}

document.addEventListener("DOMContentLoaded", function () {
  const updateButton = document.querySelector(".update-available-btn");
  const confirmUpdateButton = document.getElementById("confirmUpdate");
  const cancelUpdateButton = document.getElementById("cancelUpdate");
  const modal = document.getElementById("modal-update");

  updateButton.addEventListener("click", () => {
    modal.style.display = "block";
  });

  confirmUpdateButton.addEventListener("click", () => {
    startSSEConnection();
  });

  cancelUpdateButton.addEventListener("click", () => {
    modal.style.display = "none";
  });

  updateVersionText();
  startFetchInterval();
});

function startSSEConnection() {
  const eventSource = new EventSource(GetServerIP() + "/api/runUpdate");

  eventSource.onmessage = function (event) {
    eventLog.value += event.data + "\n";
    eventLog.scrollTop = eventLog.scrollHeight;
  };

  eventSource.onerror = function () {
    eventSource.close();
    eventLog.value += "Connection closed.\n";
  };
}
