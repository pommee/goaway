const container = document.getElementById("client-cards-container");
const modal = document.getElementById("client-modal");
const modalClose = document.getElementById("modal-close");
const modalClientName = document.getElementById("modal-client-name");
const modalClientIP = document.getElementById("modal-client-ip");
const modalClientLastSeen = document.getElementById("modal-client-last-seen");
const blockClientButton = document.getElementById("block-client");
const unblockClientButton = document.getElementById("unblock-client");
const removeClientButton = document.getElementById("remove-client");

async function getUpstreams() {
  const upstreams = await GetRequest("/upstreams");
  populateUpstreams(upstreams.upstreams);
}

function populateUpstreams(upstreams) {
  container.innerHTML = "";
  console.log(upstreams);

  upstreams.forEach((upstream) => {
    const card = document.createElement("div");
    card.className = "upstream-card";

    const header = document.createElement("h1");
    header.className = "upstream-card-header";
    header.textContent = upstream.name;

    const subheader = document.createElement("p");
    subheader.className = "upstream-card-subheader";
    subheader.textContent = "Upstream: " + upstream.upstream;

    const dnsPing = document.createElement("p");
    dnsPing.className = "upstream-card-footer";
    dnsPing.textContent = "DNS ping: " + upstream.dnsPing;

    const icmpPing = document.createElement("p");
    icmpPing.className = "upstream-card-footer";
    icmpPing.textContent = "ICMP ping: " + upstream.icmpPing;

    card.appendChild(header);
    card.appendChild(subheader);
    card.appendChild(dnsPing);
    card.appendChild(icmpPing);

    card.addEventListener("click", () => openModal(upstream));

    container.appendChild(card);
  });
}

document.addEventListener("DOMContentLoaded", () => {
  getUpstreams();
});
