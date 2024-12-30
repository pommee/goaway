const cpuUsageElement = document.getElementById("cpu-usage");
const cpuTempElement = document.getElementById("cpu-temp");
const memoryUsageElement = document.getElementById("mem-usage");
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

function getServerStatus() {
  fetch(GetServerIP() + "/server")
    .then(function (response) {
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then(function (data) {
      updateHeader(data);
    })
    .catch(function (error) {
      console.error("Failed to fetch logs:", error);
    });
}

function updateHeader(serverStatus) {
  cpuUsageElement.innerText = "CPU: " + serverStatus.cpuUsage.toFixed(1) + "%";
  cpuTempElement.innerText =
    "CPU temp: " + serverStatus.cpuTemp.toFixed(1) + "°";
  memoryUsageElement.innerText =
    "Mem: " + serverStatus.usedMemPercentage.toFixed(1) + "%";
}

document.addEventListener("DOMContentLoaded", function () {
  getServerStatus();
  quote();
  setInterval(function () {
    getServerStatus();
  }, 2500);
});

function quote() {
  quoteElement.innerText = quotes[Math.floor(Math.random() * quotes.length)];
}
