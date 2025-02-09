const container = document.getElementById("client-cards-container");
const modal = document.getElementById("client-modal");
const modalClose = document.getElementById("modal-close");
const newUpstreamBtn = document.getElementById("newUpstreamBtn");
const saveUpstreamBtn = document.getElementById("saveUpstreamBtn");
const upstreamsTextArea = document.getElementById("upstreamsTextArea");

async function getUpstreams() {
  const upstreams = await GetRequest("/upstreams");
  populateUpstreams(upstreams);
}

let currentPreferredUpstream = null;

function populateUpstreams(upstreamsData) {
  container.innerHTML = "";
  const upstreams = upstreamsData.upstreams;
  const preferredUpstream = upstreamsData.preferredUpstream;

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

    const setPreferredBtn = document.createElement("button");
    setPreferredBtn.className = "set-preferred-btn";

    if (upstream.upstream === preferredUpstream) {
      setPreferredBtn.textContent = "Preferred";
      setPreferredBtn.disabled = true;
    } else {
      setPreferredBtn.textContent = "Set as Preferred";
      setPreferredBtn.disabled = false;

      setPreferredBtn.addEventListener("click", async () => {
        const response = await GetRequest(`/preferredUpstream?upstream=${upstream.upstream}`);
        showInfoNotification(response.message);

        if (currentPreferredUpstream) {
          const previousPreferredCard = document.querySelector(`[data-upstream="${currentPreferredUpstream}"]`);
          const previousPreferredBtn = previousPreferredCard.querySelector(".set-preferred-btn");
          previousPreferredBtn.textContent = "Set as Preferred";
          previousPreferredBtn.disabled = false;
        }

        currentPreferredUpstream = upstream.upstream;
        setPreferredBtn.textContent = "Preferred";
        setPreferredBtn.disabled = true;
      });
    }

    card.setAttribute("data-upstream", upstream.upstream);

    card.appendChild(header);
    card.appendChild(subheader);
    card.appendChild(dnsPing);
    card.appendChild(icmpPing);
    card.appendChild(setPreferredBtn);

    container.appendChild(card);
  });
}

newUpstreamBtn.addEventListener("click", () => addNewUpstream());

function addNewUpstream() {
  saveUpstreamBtn.addEventListener("click", () => saveUpstream());
  modal.style.display = "flex";
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

async function saveUpstream() {
  const validUpstreamPattern = /^[0-9.:]+$/; // Only allow numbers, dots, and colons
  let containsInvalidUpstream = false;

  const newUpstreams = upstreamsTextArea.value
    .split("\n")
    .map((upstream) => upstream.trim())
    .filter((upstream) => {
      if (upstream === "") return false;
      if (!validUpstreamPattern.test(upstream)) {
        showErrorNotification(`Invalid upstream: "${upstream}" contains invalid characters.`);
        containsInvalidUpstream = true;
        return false;
      }
      return true;
    });

  if (!containsInvalidUpstream) {
    response = await PostRequest("/upstreams", JSON.stringify({ upstreams: newUpstreams }));
    closeModal();
    showInfoNotification("Added ", newUpstreams + " as new upstreams!");
  }
}

async function removeUpstream(upstream) {
  response = await DeleteRequest("/upstreams?upstream=" + upstream);
}

document.addEventListener("DOMContentLoaded", () => {
  getUpstreams();
});
