const cpuUsageElement = document.getElementById("cpu-usage");
const cpuTempElement = document.getElementById("cpu-temp");
const memoryUsageElement = document.getElementById("mem-usage");

function getServerStatus() {
  fetch("http://localhost:8080/server")
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
  setInterval(function () {
    getServerStatus();
  }, 2500);
});
