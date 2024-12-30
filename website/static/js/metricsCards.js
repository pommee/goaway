// Store the card elements to update
const cardQueries = document.querySelector("#card-queries .card-text");
const cardBlocked = document.querySelector("#card-blocked .card-text");
const cardBlockedPercentage = document.querySelector(
  "#card-blocked-percentage .card-text",
);
const cardBlockedDomains = document.querySelector(
  "#card-blocked-domains .card-text",
);

function getStatus() {
  fetch(GetServerIP() + "/metrics")
    .then((response) => {
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }

      return response.text(); // Get response as text
    })
    .then((text) => {
      if (!text) {
        return;
      }

      try {
        const data = JSON.parse(text);
        if (data && Object.keys(data).length > 0) {
          updateDashboardCards(data);
        } else {
          console.log("No data available");
        }
      } catch (e) {
        throw new Error("Failed to parse JSON");
      }
    })
    .catch((error) => {
      console.error("Failed to fetch logs:", error);
    });
}

function updateDashboardCards(status) {
  if (cardQueries.innerText !== status.total.toLocaleString()) {
    applyAnimation(cardQueries);
    cardQueries.innerText = status.total.toLocaleString();
  }

  if (cardBlocked.innerText !== status.blocked.toLocaleString()) {
    applyAnimation(cardBlocked);
    cardBlocked.innerText = status.blocked.toLocaleString();
  }

  if (
    cardBlockedPercentage.innerText !==
    status.percentageBlocked.toFixed(1) + "%"
  ) {
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
