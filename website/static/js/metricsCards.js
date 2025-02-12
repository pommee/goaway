const cardQueriesClients = document.querySelector("#card-queries p");
const cardQueries = document.querySelector("#card-queries .card-text");
const cardBlocked = document.querySelector("#card-blocked .card-text");
const cardBlockedPercentage = document.querySelector("#card-blocked-percentage .card-text");
const cardBlockedDomains = document.querySelector("#card-blocked-domains .card-text");

async function getStatus() {
  metrics = await GetRequest("/metrics");
  if (metrics && Object.keys(metrics).length > 0) {
    updateDashboardCards(metrics);
  } else {
    console.log("No data available");
  }
}

function updateDashboardCards(status) {
  if (cardQueries.innerText !== status.total.toLocaleString()) {
    applyAnimation(cardQueries);
    cardQueries.innerText = status.total.toLocaleString();
    cardQueriesClients.innerText = `Total queries (${status.clients} clients)`
  }

  if (cardBlocked.innerText !== status.blocked.toLocaleString()) {
    applyAnimation(cardBlocked);
    cardBlocked.innerText = status.blocked.toLocaleString();
  }

  if (cardBlockedPercentage.innerText !== status.percentageBlocked.toFixed(1) + "%") {
    applyAnimation(cardBlockedPercentage);
    cardBlockedPercentage.innerText = status.percentageBlocked.toFixed(1) + "%";
  }

  if (cardBlockedDomains.innerText !== status.domainBlockLen.toLocaleString()) {
    applyAnimation(cardBlockedDomains);
    cardBlockedDomains.innerText = status.domainBlockLen.toLocaleString();
  }
}

function applyAnimation(cardElement) {
  cardElement.classList.add("update-animation");

  setTimeout(function () {
    cardElement.classList.remove("update-animation");
  }, 500);
}

document.addEventListener("DOMContentLoaded", function () {
  getStatus();
  setInterval(function () {
    getStatus();
  }, 1000);
});
