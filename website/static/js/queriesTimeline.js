let chart = null;

function getQueries() {
  fetch(GetServerIP() + "/queriesData")
    .then(function (response) {
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then(function (data) {
      updateTimeline(data);
    })
    .catch(function (error) {
      console.error("Failed to fetch logs:", error);
    });
}

function aggregateData(data, hours = null) {
  const now = new Date();
  const resultData = {};

  if (hours !== null) {
    for (let i = hours - 1; i >= 0; i--) {
      const hour = new Date(now.getTime() - i * 3600000);
      hour.setMinutes(0, 0, 0);
      resultData[hour.getTime()] = {
        timestamp: hour,
        blocked: 0,
        nonBlocked: 0,
      };
    }
  }

  data.forEach((entry) => {
    const timestamp = new Date(entry.timestamp);
    timestamp.setMinutes(0, 0, 0);
    const timeKey = timestamp.getTime();

    if (!resultData[timeKey]) {
      resultData[timeKey] = {
        timestamp,
        blocked: 0,
        nonBlocked: 0,
      };
    }

    if (entry.blocked) {
      resultData[timeKey].blocked++;
    } else {
      resultData[timeKey].nonBlocked++;
    }
  });

  return resultData;
}

function updateTimeline(data) {
  const allData = aggregateData(data.details);
  const allChartData = Object.values(allData);

  if (!chart) {
    const initialData = aggregateData(data.details, 60);
    const initialChartData = Object.values(initialData);

    const ctx = document.getElementById("requestChart").getContext("2d");
    chart = new Chart(ctx, {
      type: "bar",
      data: {
        datasets: [
          {
            label: "blocked",
            data: allChartData,
            parsing: {
              xAxisKey: "timestamp",
              yAxisKey: "blocked",
            },
            backgroundColor: "rgba(173, 7, 18, 1)",
            order: 2,
            barPercentage: 1,
          },
          {
            label: "non-blocked",
            data: allChartData,
            parsing: {
              xAxisKey: "timestamp",
              yAxisKey: "nonBlocked",
            },
            backgroundColor: "green",
            order: 1,
            barPercentage: 1,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          x: {
            type: "time",
            time: {
              unit: "hour",
              displayFormats: {
                hour: "HH:00",
              },
            },
            title: {
              display: true,
              text: "Time",
            },
            stacked: true,
            min: initialChartData[0]?.timestamp || null,
            max: initialChartData[initialChartData.length - 1]?.timestamp || null,
          },
          y: {
            beginAtZero: true,
            title: {
              display: true,
              text: "Number of Requests",
            },
            ticks: {
              stepSize: 1,
            },
            stacked: true,
          },
        },
        plugins: {
          zoom: {
            zoom: {
              wheel: {
                enabled: true,
              },
              pinch: {
                enabled: true,
              },
              mode: "x",
            },
            pan: {
              enabled: true,
              mode: "x",
            },
          },
        },
      },
    });
  } else {
    const currentMin = chart.options.scales.x.min;
    const currentMax = chart.options.scales.x.max;

    chart.data.datasets.forEach((dataset) => {
      dataset.data = allChartData;
    });

    // Preserve current zoom/pan state
    chart.options.scales.x.min = currentMin;
    chart.options.scales.x.max = currentMax;

    chart.update();
  }
}

document.addEventListener("DOMContentLoaded", function () {
  getQueries();
  setInterval(function () {
    getQueries();
  }, 1000);
});

document.getElementById("logout").addEventListener("click", () => Logout());
