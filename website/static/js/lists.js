const container = document.getElementById("client-cards-container");
const modal = document.getElementById("client-modal");
const modalClose = document.getElementById("modal-close");
const newListBtn = document.getElementById("newListBtn");
const saveListBtn = document.getElementById("saveListBtn");
const listName = document.getElementById("listName");
const listsTextArea = document.getElementById("listsTextArea");

async function getLists() {
  const lists = await GetRequest("/lists");
  populateLists(lists);
  createSourceButtons(lists);
}

function populateLists(listsData) {
  container.innerHTML = "";
  const lists = listsData.lists;

  for (const key in lists) {
    console.log(`${key}: ${lists[key]}`);
    const card = document.createElement("div");
    card.className = "list-card";

    const header = document.createElement("h3");
    header.className = "list-card-header";
    header.textContent = key;

    const subheader = document.createElement("p");
    subheader.className = "list-card-subheader";
    subheader.textContent = "Blocked: " + lists[key];

    card.appendChild(header);
    card.appendChild(subheader);
    card.addEventListener("click", () => showListDetails(key));
    container.appendChild(card);
  }
}

function showListDetails(listName) {
  const listDetailsModal = document.getElementById("list-details-modal");
  const listDetailsContent = document.getElementById("list-details-content");
  listDetailsContent.innerHTML = `<span id="list-details-close" class="list-details-close">&times;</span><h2>${listName}</h2>`;

  fetch(`/getDomainsForList?list=${listName}`)
    .then(response => response.json())
    .then(data => {
      const domains = data.domains;
      const table = document.createElement("table");
      table.id = "domains-table";
      table.className = "display";
      const thead = document.createElement("thead");
      thead.innerHTML = `
        <tr>
          <th>Domain</th>
          <th>Toggle</th>
        </tr>
      `;
      const tbody = document.createElement("tbody");
      domains.forEach(domain => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
          <td>${domain}</td>
          <td><button class="toggle-button blocked-true" data-blocked="true" data-domain="${domain}">Whitelist</button></td>
        `;
        tbody.appendChild(tr);
      });
      table.appendChild(thead);
      table.appendChild(tbody);
      listDetailsContent.appendChild(table);
    });

  listDetailsModal.style.display = "flex";
}

document.getElementById("list-details-close").onclick = () => {
  document.getElementById("list-details-modal").style.display = "none";
};

window.onclick = (event) => {
  if (event.target === document.getElementById("list-details-modal")) {
    document.getElementById("list-details-modal").style.display = "none";
  }
};

function createSourceButtons(listsData) {
  const sourcesContainer = document.getElementById("sources-container");
  sourcesContainer.innerHTML = "";
  const lists = listsData.lists;

  for (const key in lists) {
    const button = document.createElement("button");
    button.className = "source-btn";
    button.textContent = key;
    button.addEventListener("click", () => {
      listName.value = key;
      modal.style.display = "flex";
    });
    sourcesContainer.appendChild(button);
  }
}

newListBtn.addEventListener("click", () => addNewList());

function addNewList() {
  saveListBtn.addEventListener("click", () => saveList());
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

async function saveList() {
  const validListPattern = /^[a-z0-9.:]+$/; // Only allow numbers, dots, and colons
  let containsInvalidList = false;

  const newLists = listsTextArea.value
    .split("\n")
    .map((list) => list.trim())
    .filter((list) => {
      if (list === "") return false;
      if (!validListPattern.test(list)) {
        showErrorNotification(`Invalid list: "${list}" contains invalid characters.`);
        containsInvalidList = true;
        return false;
      }
      return true;
    });

  if (!containsInvalidList) {
    response = await PostRequest("/lists", JSON.stringify({ list: listName.value, domains: newLists }));
    closeModal();
    showInfoNotification("Added ", newLists + " as new lists!");
  }
}

async function removeList(list) {
  response = await DeleteRequest("/lists?list=" + list);
  console.log(response);
}

document.addEventListener("DOMContentLoaded", () => {
  getLists();
});
