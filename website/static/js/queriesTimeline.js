let queryChart = null;
let queryTypeChart = null;

function initializeCharts() {
  const queryCtx = document.getElementById("requestChart").getContext("2d");
  const queryTypeCtx = document
    .getElementById("requestTypeChart")
    .getContext("2d");

  const labels = Array.from({ length: 144 }, (_, i) => {
    const hour = Math.floor(i / 6)
      .toString()
      .padStart(2, "0");
    const minute = ((i % 6) * 10).toString().padStart(2, "0");
    return `${hour}:${minute}`;
  });

  queryChart = new Chart(queryCtx, {
    type: "bar",
    data: {
      labels: labels,
      datasets: [
        {
          label: "Blocked",
          backgroundColor: "rgba(173, 0, 0, 0.8)",
          data: Array(144).fill(0),
          stack: "stack0",
        },
        {
          label: "Allowed",
          backgroundColor: "rgba(40, 141, 0, 0.8)",
          data: Array(144).fill(0),
          stack: "stack0",
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: {
          display: true,
          text: "Requests (Last 24 Hours)",
        },
      },
      scales: {
        x: {
          stacked: true,
          title: {
            display: true,
            text: "Time",
          },
          ticks: {
            maxRotation: 45,
            callback: function (val) {
              if (val % 6 === 0) {
                return this.getLabelForValue(val);
              }
              return "";
            },
          },
        },
        y: {
          stacked: true,
          beginAtZero: true,
          title: {
            display: true,
            text: "Queries",
          },
        },
      },
    },
  });

  queryTypeChart = new Chart(queryTypeCtx, {
    type: "doughnut",
    data: {
      datasets: [
        {
          label: "Occurrences",
          backgroundColor: [
            "rgba(255, 99, 132, 0.5)",
            "rgba(54, 162, 235, 0.5)",
            "rgba(255, 206, 86, 0.5)",
            "rgba(75, 192, 192, 0.5)",
          ],
          borderColor: [
            "rgba(255,99,132,1)",
            "rgba(54, 162, 235, 1)",
            "rgba(255, 206, 86, 1)",
            "rgba(75, 192, 192, 1)",
          ],
          borderWidth: 1,
        },
      ],
    },
    options: {
      maintainAspectRatio: true,
      layout: {
        padding: 10,
      },
      plugins: {
        legend: {
          display: true,
          position: "right",
          labels: {
            boxWidth: 10,
            padding: 10,
          },
        },
      },
    },
  });
}

function updateDashboard(data) {
  const now = new Date();
  const twentyFourHoursAgo = new Date(now - 24 * 60 * 60 * 1000);

  const intervalData = Array(144)
    .fill()
    .map(() => ({
      blocked: 0,
      allowed: 0,
      total: 0,
    }));

  let totalQueries = 0;
  let blockedQueries = 0;

  data.queries.forEach((query) => {
    const queryDate = new Date(query.timestamp);
    if (queryDate >= twentyFourHoursAgo && queryDate <= now) {
      const minutesAgo = (now - queryDate) / (1000 * 60);
      const intervalsAgo = Math.floor(minutesAgo / 10);

      let intervalIndex = 143 - intervalsAgo;

      if (intervalIndex >= 0 && intervalIndex < 144) {
        if (query.blocked) {
          intervalData[intervalIndex].blocked++;
          blockedQueries++;
        } else {
          intervalData[intervalIndex].allowed++;
        }
        intervalData[intervalIndex].total++;
        totalQueries++;
      }
    }
  });

  const newLabels = Array.from({ length: 144 }, (_, i) => {
    const minutesAgo = (143 - i) * 10;
    const labelTime = new Date(now - minutesAgo * 60 * 1000);
    const hour = labelTime.getHours().toString().padStart(2, "0");
    const minute = labelTime.getMinutes().toString().padStart(2, "0");
    return `${hour}:${minute}`;
  });

  queryChart.data.labels = newLabels;
  queryChart.data.datasets[0].data = intervalData.map((h) => h.blocked);
  queryChart.data.datasets[1].data = intervalData.map((h) => h.allowed);
  queryChart.update();
}

async function updateDoughnut(data) {
  const intervalData = {};

  data.queries.forEach((query) => {
    intervalData[query.queryType] = query.count;
  });

  queryTypeChart.data.labels = Object.keys(intervalData);
  queryTypeChart.data.datasets[0].data = Object.values(intervalData);
  queryTypeChart.update();
}

async function getQueries() {
  const data = await GetRequest("/queryTimestamps");
  updateDashboard(data);
}

document.addEventListener("DOMContentLoaded", function () {
  initializeCharts();
  getQueries();
  fetchTopBlockedDomains();
  fetchTopClients();
  fetchTypes();
  setInterval(getQueries, 1000);
  setInterval(fetchTopBlockedDomains, 1000);
  setInterval(fetchTopClients, 1000);
  setInterval(fetchTypes, 1000);
});

document.getElementById("logout").addEventListener("click", () => Logout());

function updateTable(data, tableId, dataKey, nameKey, countKey) {
  const tbody = document.getElementById(tableId);
  tbody.innerHTML = "";

  if (!data[dataKey]) {
    const row = document.createElement("tr");
    const cell = document.createElement("td");
    cell.style.padding = "10px";
    cell.textContent = `No ${dataKey}`;
    row.appendChild(cell);
    tbody.appendChild(row);
    return;
  }

  data[dataKey].forEach((item) => {
    const row = document.createElement("tr");
    const nameCell = document.createElement("td");
    const countCell = document.createElement("td");
    const frequencyCell = document.createElement("td");
    const frequencyBarContainer = document.createElement("div");
    const frequencyBar = document.createElement("div");

    nameCell.textContent = item[nameKey];
    countCell.textContent = item[countKey];

    frequencyBarContainer.classList.add("frequency-bar-container");
    frequencyBar.classList.add("frequency-bar");
    frequencyBar.style.width = `${item.frequency}%`;
    frequencyBarContainer.appendChild(frequencyBar);
    frequencyCell.appendChild(frequencyBarContainer);

    row.appendChild(nameCell);
    row.appendChild(countCell);
    row.appendChild(frequencyCell);

    tbody.appendChild(row);
  });
}

async function fetchTypes() {
  data = await GetRequest("/queryTypes");
  updateDoughnut(data);
}

function fetchData(url, tableId, dataKey, nameKey, countKey) {
  fetch(GetServerIP() + url)
    .then((response) => {
      if (!response.ok) throw new Error("Network response was not ok");
      return response.json();
    })
    .then((data) => {
      updateTable(data, tableId, dataKey, nameKey, countKey);
    })
    .catch((error) => {
      console.error(`Failed to fetch data from ${url}:`, error);
    });
}

function fetchTopBlockedDomains() {
  fetchData(
    "/api/topBlockedDomains",
    "blocked-domains-body",
    "domains",
    "name",
    "hits",
  );
}

function fetchTopClients() {
  fetchData(
    "/api/topClients",
    "top-clients-body",
    "clients",
    "client",
    "requestCount",
  );
}
