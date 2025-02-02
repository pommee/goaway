const container = document.getElementById("client-cards-container");
const modalUpdateCustom = document.getElementById("modal-update-custom");
const modalClose = document.getElementById("modal-close");
const newListBtn = document.getElementById("newListBtn");
const updateCustomBtn = document.getElementById("updateCustomBtn");
const saveListBtn = document.getElementById("saveListBtn");
const listsTextArea = document.getElementById("listsTextArea");

const modalNewList = document.getElementById("modal-new-list");
const modalNewListClose = document.getElementById("modal-new-list-close");
const saveNewListBtn = document.getElementById("saveNewListBtn");
const newListName = document.getElementById("newListName");
const newListURL = document.getElementById("newListURL");

newListBtn.addEventListener("click", () => {
  modalNewList.style.display = "flex";
});

modalNewListClose.onclick = () => {
  modalNewList.style.display = "none";
};

window.onclick = (event) => {
  if (event.target === modalNewList) {
    modalNewList.style.display = "none";
  }
};

saveNewListBtn.addEventListener("click", async () => {
  const name = newListName.value.trim();
  const url = newListURL.value.trim();

  if (name && url) {
    try {
      const response = await fetch(`/addList?name=${encodeURIComponent(name)}&url=${encodeURIComponent(url)}`, {
        method: "GET",
      });

      if (response.ok) {
        showInfoNotification("New list added successfully!");
        newListName.value = "";
        newListURL.value = "";
        modalNewList.style.display = "none";
        getLists();
      } else {
        const errorData = await response.json();
        showErrorNotification(errorData.error);
      }
    } catch (error) {
      showErrorNotification("Error adding new list.");
    }
  } else {
    showErrorNotification("Please provide both name and URL.");
  }
});

async function getLists() {
  const lists = await GetRequest("/lists");
  populateLists(lists);
}

function populateLists(listsData) {
  container.innerHTML = "";
  const lists = listsData.lists;

  for (const key in lists) {
    const list = lists[key];
    const blockedCount = list.blocked_count;
    const lastUpdated = new Date(list.lastUpdated * 1000).toLocaleString();

    const card = document.createElement("div");
    card.className = "list-card";

    const header = document.createElement("h3");
    header.className = "list-card-header";
    header.textContent = key;

    const subheader = document.createElement("p");
    subheader.className = "list-card-subheader";
    subheader.textContent = `Blocked: ${blockedCount}`;

    const updated = document.createElement("p");
    updated.className = "list-card-subheader";
    updated.textContent = `Last Updated: ${lastUpdated}`;

    card.appendChild(header);
    card.appendChild(subheader);
    card.appendChild(updated);
    card.addEventListener("click", () => showListDetails(key));
    container.appendChild(card);
  }
}

function showListDetails(listName) {
  const listDetailsModal = document.getElementById("list-details-modal");
  const listDetailsContent = document.getElementById("list-details-content");
  listDetailsContent.innerHTML = `
    <span id="list-details-close" class="list-details-close">&times;</span>
    <h2>${listName}</h2>
    <button id="removeListBtn" class="remove-btn">Remove List</button>
  `;

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

  document.getElementById("removeListBtn").addEventListener("click", async () => {
    try {
      const response = await fetch(`/list?name=${listName}`, {
        method: "DELETE",
      });

      if (response.ok) {
        showInfoNotification("List removed successfully!");
        listDetailsModal.style.display = "none";
        getLists();
      } else {
        const errorData = await response.json();
        showErrorNotification(errorData.error);
      }
    } catch (error) {
      showErrorNotification("Error removing list.");
    }
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


updateCustomBtn.addEventListener("click", () => updateCustomList());

function updateCustomList() {
  saveListBtn.addEventListener("click", () => saveCustom());
  modalUpdateCustom.style.display = "flex";
}

function closeModal() {
  modalUpdateCustom.style.display = "none";
}

modalClose.onclick = closeModal;

window.onclick = (event) => {
  if (event.target === modalUpdateCustom) {
    closeModal();
  }
};

async function saveCustom() {
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
    response = await PostRequest("/custom", JSON.stringify({ domains: newLists }));
    closeModal();
    showInfoNotification("Updated custom list!");
  }
}

async function removeList(list) {
  response = await DeleteRequest("/lists?list=" + list);
  console.log(response);
}

document.addEventListener("DOMContentLoaded", () => {
  getLists();
});
