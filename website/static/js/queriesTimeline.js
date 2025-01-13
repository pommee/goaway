let queryChart = null;

function initializeCharts() {
  const queryCtx = document.getElementById('requestChart').getContext('2d');
  
  const labels = Array.from({length: 96}, (_, i) => {
    const hour = Math.floor(i / 4).toString().padStart(2, '0');
    const minute = (i % 4 * 15).toString().padStart(2, '0');
    return `${hour}:${minute}`;
  });

  queryChart = new Chart(queryCtx, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [
        {
          label: 'Blocked',
          backgroundColor: 'rgba(173, 0, 0, 0.8)',
          data: Array(96).fill(0),
          stack: 'stack0',
        },
        {
          label: 'Allowed',
          backgroundColor: 'rgba(40, 141, 0, 0.8)',
          data: Array(96).fill(0),
          stack: 'stack0',
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: {
          display: true,
          text: 'Requests last 24h'
        },
      },
      scales: {
        x: {
          stacked: true,
          title: {
            display: true,
            text: 'Time',
          },
          ticks: {
            maxRotation: 45,
            callback: function(val) {
              return this.getLabelForValue(val);
            }
          }
        },
        y: {
          stacked: true,
          beginAtZero: true,
          title: {
            display: true,
            text: 'Queries'
          }
        }
      }
    }
  });
}

function updateDashboard(data) {
  const now = new Date();
  const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  
  const intervalData = Array(96).fill().map(() => ({
    blocked: 0,
    allowed: 0,
    total: 0
  }));

  let totalQueries = 0;
  let blockedQueries = 0;

  data.queries.forEach(query => {
    const queryDate = new Date(query.timestamp);
    if (queryDate >= startOfDay && queryDate <= now) {
      const hour = queryDate.getHours();
      const minute = queryDate.getMinutes();
      const intervalIndex = (hour * 4) + Math.floor(minute / 15);
      
      if (query.blocked) {
        intervalData[intervalIndex].blocked++;
        blockedQueries++;
      } else {
        intervalData[intervalIndex].allowed++;
      }
      intervalData[intervalIndex].total++;
      totalQueries++;
    }
  });

  queryChart.data.datasets[0].data = intervalData.map(h => h.blocked);
  queryChart.data.datasets[1].data = intervalData.map(h => h.allowed);
  queryChart.update();
}

function getQueries() {
  fetch(GetServerIP() + "/queryTimestamps")
    .then(response => {
      if (!response.ok) throw new Error("Network response was not ok");
      return response.json();
    })
    .then(data => {
      updateDashboard(data);
    })
    .catch(error => {
      console.error("Failed to fetch logs:", error);
    });
}

document.addEventListener('DOMContentLoaded', function() {
  initializeCharts();
  getQueries();
  setInterval(getQueries, 1000);
});

document.getElementById("logout").addEventListener("click", () => Logout());