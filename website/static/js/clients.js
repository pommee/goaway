const container = document.getElementById("client-cards-container");
const modal = document.getElementById("client-modal");
const modalClose = document.getElementById("modal-close");
const modalClientName = document.getElementById("modal-client-name");
const modalClientIP = document.getElementById("modal-client-ip");
const modalClientLastSeen = document.getElementById("modal-client-last-seen");
const blockClientButton = document.getElementById("block-client");
const unblockClientButton = document.getElementById("unblock-client");
const removeClientButton = document.getElementById("remove-client");

async function getClients() {
  const clients = await GetRequest("/clients");
  await populateClientsTable(clients);
}

async function populateClientsTable(data) {
  container.innerHTML = "";

  data.clients.forEach((client) => {
    const card = document.createElement("div");
    card.className = "client-card";

    const header = document.createElement("h1");
    header.className = "client-card-header";
    header.textContent = client.Name;

    const subheader = document.createElement("h4");
    subheader.className = "client-card-subheader";
    subheader.textContent = client.IP;

    const footer = document.createElement("p");
    footer.className = "client-card-footer";
    footer.textContent = `Last seen: ${formatTimestamp(client.lastSeen)}`;

    card.appendChild(header);
    card.appendChild(subheader);
    card.appendChild(footer);

    card.addEventListener("click", async () => await openModal(client));

    container.appendChild(card);
  });
}

document.addEventListener("DOMContentLoaded", () => {
  getClients();
});

async function openModal(client) {
  const clientDetailsReq = await GetRequest(`/clientDetails?clientIP=${client.IP}`);
  const clientDetails = clientDetailsReq.details

  modal.style.display = "flex";

  modalClientName.textContent = `Name: ${client.Name}`;
  modalClientIP.textContent = `IP: ${client.IP}`;
  modalClientLastSeen.textContent = `Last Seen: ${formatTimestamp(client.lastSeen)}`;
  
  document.getElementById("modal-total-requests").textContent = clientDetails.TotalRequests;
  document.getElementById("modal-unique-domains").textContent = clientDetails.UniqueDomains;
  document.getElementById("modal-blocked-requests").textContent = clientDetails.BlockedRequests;
  document.getElementById("modal-cached-requests").textContent = clientDetails.CachedRequests;
  document.getElementById("modal-avg-response-time").textContent = `${clientDetails.AvgResponseTimeMs.toFixed(2)} ms`;
  document.getElementById("modal-most-queried").textContent = clientDetails.MostQueriedDomain || "N/A";

  const domainListContainer = document.getElementById("modal-all-domains");
  domainListContainer.innerHTML = "";

  if (clientDetails.AllDomains.length > 0) {
      clientDetails.AllDomains.forEach((domain) => {
          const domainItem = document.createElement("p");
          domainItem.textContent = domain;
          domainItem.className = "domain-item";
          domainListContainer.appendChild(domainItem);
      });
  } else {
      domainListContainer.innerHTML = "<p>No domains queried.</p>";
  }
}


function closeModal() {
  modal.style.display = "none";
}

modalClose.onclick = closeModal;

window.onclick = (event) => {
  if (event.target === modal) {
    closeModal();
  }
};
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
