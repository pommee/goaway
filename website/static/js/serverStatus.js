const serverVersion = document.getElementById("server-version");
const cpuUsageElement = document.getElementById("cpu-usage");
const cpuTempElement = document.getElementById("cpu-temp");
const memoryUsageElement = document.getElementById("mem-usage");
const dbUsageElement = document.getElementById("db-usage");
const quoteElement = document.getElementsByClassName("top-section-text")[0];

const quotes = [
  "Block party!",
  "No ads!",
  "Bye-bye, spam!",
  "Adios, ads!",
  "Get lost!",
  "Bye, trackers!",
  "Stop, right there!",
  "Catch you later!",
  "Ad free zone!",
  "Blockzilla strikes!",
  "Ad-ocalypse now!",
  "Nope, not today!",
  "Buzz off, ads!",
  "Ads? Not here!",
  "Shh... no ads.",
  "Ad blocker engaged!",
  "Gone in a click!",
  "Ads begone!",
  "Block mode: ON!",
  "Spam, who?",
  "No entry for ads!",
  "Bye-bye bandwidth hogs!",
  "Ad-free vibes!",
  "Don't block me!",
  "Not in my house!",
  "Get out, ads!",
  "Bye-bye popups!",
  "Say no to ads!",
  "Ad block, power!",
  "Stay adless!",
];

async function getServerStatus() {
  const serverStatus = await GetRequest("/server");

  saveInstalledVersion(serverStatus.version);
  if (window.location.pathname !== "/login.html") {
    updateHeader(serverStatus);
  }
}

function saveInstalledVersion(version) {
  localStorage.setItem("installedVersion", version);
}

function GetInstalledVersion() {
  return localStorage.getItem("installedVersion");
}

function GetLatestVersion() {
  return localStorage.getItem("latestVersion");
}

function updateHeader(serverStatus) {
  saveInstalledVersion(serverStatus.version);
  serverVersion.innerText = `v${serverStatus.version}`;
  cpuUsageElement.innerText = `CPU: ${serverStatus.cpuUsage.toFixed(1)}%`;
  cpuTempElement.innerText = `CPU temp: ${serverStatus.cpuTemp.toFixed(1)}°`;
  memoryUsageElement.innerText = `Mem: ${serverStatus.usedMemPercentage.toFixed(
    1
  )}%`;
  dbUsageElement.innerText = `Size: ${serverStatus.dbSize.toFixed(1)}MB`;
}

function checkIfUpdateAvailable() {
  function semverCompare(a, b) {
    const pa = a.split(".").map((n) => parseInt(n, 10));
    const pb = b.split(".").map((n) => parseInt(n, 10));

    for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
      const numA = pa[i] ?? 0;
      const numB = pb[i] ?? 0;
      if (numA > numB) return 1;
      if (numA < numB) return -1;
    }
    return 0;
  }
  const installedVersion = GetInstalledVersion();
  const latestVersion = GetLatestVersion();
  const updateAvailable = semverCompare(latestVersion, installedVersion);

  if (updateAvailable === 1) {
    let indicator = document.getElementById("update-available-indicator");
    indicator.classList.remove("hidden");
  }
}

document.addEventListener("DOMContentLoaded", async function () {
  try {
    checkIfUpdateAvailable();
  } catch {}
  await getServerStatus();
  quote();
  setInterval(async function () {
    await getServerStatus();
  }, 2500);
});

function quote() {
  quoteElement.innerText = quotes[Math.floor(Math.random() * quotes.length)];
}
