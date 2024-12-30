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

function aggregateDataByHour(data) {
  // Create an object to store hourly data for the last 24 hours
  const now = new Date();
  const hourlyData = {};

  // Initialize last 24 hours with zero values
  for (let i = 23; i >= 0; i--) {
    const hour = new Date(now.getTime() - i * 3600000);
    hour.setMinutes(0, 0, 0);
    hourlyData[hour.getTime()] = {
      timestamp: hour,
      blocked: 0,
      nonBlocked: 0,
    };
  }

  // Aggregate the actual data
  data.forEach((entry) => {
    const timestamp = new Date(entry.timestamp);
    timestamp.setMinutes(0, 0, 0);
    const timeKey = timestamp.getTime();

    if (hourlyData[timeKey]) {
      if (entry.blocked) {
        hourlyData[timeKey].blocked++;
      } else {
        hourlyData[timeKey].nonBlocked++;
      }
    }
  });

  return hourlyData;
}

function updateTimeline(data) {
  if (!chart) {
    const aggregatedData = aggregateDataByHour(data.details);
    const chartData = Object.values(aggregatedData);
    const ctx = document.getElementById("requestChart").getContext("2d");
    chart = new Chart(ctx, {
      type: "bar",
      data: {
        datasets: [
          {
            label: "blocked",
            data: chartData,
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
            data: chartData,
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
    const aggregatedData = aggregateDataByHour(data.details);
    const chartData = Object.values(aggregatedData);
    chart.data.datasets.forEach((dataset) => {
      dataset.data = chartData;
    });
    chart.update();
  }
}

document.addEventListener("DOMContentLoaded", function () {
  getQueries();
  setInterval(function () {
    getQueries();
  }, 1000);
});
