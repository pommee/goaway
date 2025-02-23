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

  if (window.location.pathname !== "/login.html") {
    updateHeader(serverStatus);
  }
}

function storeVersion(version) {
  localStorage.setItem("version", version);
}

function GetVersion() {
  return localStorage.getItem("version");
}

function updateHeader(serverStatus) {
  storeVersion(serverStatus.version);
  serverVersion.innerText = `v${serverStatus.version}`;
  cpuUsageElement.innerText = `CPU: ${serverStatus.cpuUsage.toFixed(1)}%`;
  cpuTempElement.innerText = `CPU temp: ${serverStatus.cpuTemp.toFixed(1)}°`;
  memoryUsageElement.innerText = `Mem: ${serverStatus.usedMemPercentage.toFixed(
    1
  )}%`;
  dbUsageElement.innerText = `Size: ${serverStatus.dbSize.toFixed(1)}MB`;
}

document.addEventListener("DOMContentLoaded", async function () {
  await getServerStatus();
  quote();
  setInterval(async function () {
    await getServerStatus();
  }, 2500);
});

function quote() {
  quoteElement.innerText = quotes[Math.floor(Math.random() * quotes.length)];
}
