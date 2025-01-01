const container = document.getElementById("client-cards-container");

function getClients() {
  fetch(GetServerIP() + "/clients")
    .then(function (response) {
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then(function (data) {
      populateClientsTable(data);
    })
    .catch(function (error) {
      console.error("Failed to fetch clients:", error);
    });
}

function formatTimestamp(timestamp) {
  const date = new Date(timestamp);

  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  const seconds = String(date.getSeconds()).padStart(2, "0");

  return `${year}/${month}/${day} ${hours}:${minutes}:${seconds}`;
}

function populateClientsTable(data) {
  container.innerHTML = "";

  data.clients.forEach((client) => {
    const card = document.createElement("div");
    card.className = "client-card";

    const header = document.createElement("h1");
    header.className = "client-card-header";
    header.textContent = client.Name;

    const subheader = document.createElement("h3");
    subheader.className = "client-card-subheader";
    subheader.textContent = client.IP;

    const footer = document.createElement("p");
    footer.className = "client-card-footer";
    footer.textContent = `Last seen: ${formatTimestamp(client.lastSeen)}`;

    card.appendChild(header);
    card.appendChild(subheader);
    card.appendChild(footer);

    container.appendChild(card);
  });
}

document.addEventListener("DOMContentLoaded", () => {
  getClients();
});
