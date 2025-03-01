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

const ANIMATION_STAGGER_DELAY = 80;
const ANIMATION_BASE_DELAY = 150;

const modalAnimDuration = 300;

function initAnimations() {
  const style = document.createElement("style");
  document.head.appendChild(style);
}

function addButtonAnimations() {
  const buttons = document.querySelectorAll("button");
  buttons.forEach((button) => {
    button.addEventListener("mousedown", () => {
      button.style.transform = "scale(0.95)";
    });

    button.addEventListener("mouseup", () => {
      button.style.transform = "";
    });

    button.addEventListener("mouseleave", () => {
      button.style.transform = "";
    });
  });
}

newListBtn.addEventListener("click", () => {
  modalNewList.style.display = "flex";
  modalNewList.style.backgroundColor = "rgba(0, 0, 0, 0)";
  setTimeout(() => {
    modalNewList.style.backgroundColor = "rgba(0, 0, 0, 0.7)";
  }, 10);
});

modalNewListClose.onclick = () => {
  closeModalWithAnimation(modalNewList);
};

function closeModalWithAnimation(modal) {
  modal.style.backgroundColor = "rgba(0, 0, 0, 0)";
  const content =
    modal.querySelector(".modal-content") ||
    modal.querySelector(".list-details-content");
  content.style.opacity = "0";
  content.style.transform = "scale(0.9)";

  setTimeout(() => {
    modal.style.display = "none";
    content.style.opacity = "";
    content.style.transform = "";
  }, modalAnimDuration);
}

window.onclick = (event) => {
  if (event.target === modalNewList) {
    closeModalWithAnimation(modalNewList);
  } else if (event.target === modalUpdateCustom) {
    closeModalWithAnimation(modalUpdateCustom);
  } else if (event.target === document.getElementById("list-details-modal")) {
    closeModalWithAnimation(document.getElementById("list-details-modal"));
  }
};

saveNewListBtn.addEventListener("click", async () => {
  const name = newListName.value.trim();
  const url = newListURL.value.trim();

  if (name && url) {
    animateButton(saveNewListBtn);

    try {
      const response = await fetch(
        `/api/addList?name=${encodeURIComponent(name)}&url=${encodeURIComponent(
          url
        )}`,
        {
          method: "GET",
        }
      );

      if (response.ok) {
        showInfoNotification("New list added successfully!");
        newListName.value = "";
        newListURL.value = "";
        closeModalWithAnimation(modalNewList);
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

function animateButton(button) {
  button.classList.add("animating");
  button.style.transform = "scale(0.95)";
  setTimeout(() => {
    button.style.transform = "";
    button.classList.remove("animating");
  }, 300);
}

function populateLists(listsData) {
  container.innerHTML = "";
  const lists = listsData.lists;
  let delay = ANIMATION_BASE_DELAY;

  for (const key in lists) {
    const list = lists[key];
    const blockedCount = list.blocked_count;
    const lastUpdated = new Date(list.lastUpdated * 1000).toLocaleString();

    const card = document.createElement("div");
    card.className = "list-card";
    card.style.animationDelay = `${delay}ms`;

    const cardContent = document.createElement("div");
    cardContent.className = "card-content";

    const statusIndicator = document.createElement("div");
    statusIndicator.className = "card-status";
    if (list.active === false) {
      statusIndicator.classList.add("inactive");
    }

    const header = document.createElement("h3");
    header.className = "list-card-header";
    header.textContent = key;

    const counter = document.createElement("div");
    counter.className = "card-counter";
    counter.innerHTML = `<i class="fa-solid fa-ban"></i> ${blockedCount}`;

    const subheader = document.createElement("p");
    subheader.className = "list-card-subheader";
    subheader.textContent = `Domains blocked in this list`;

    const updated = document.createElement("p");
    updated.className = "list-card-updated";
    updated.innerHTML = `<i class="fa-regular fa-clock"></i> ${lastUpdated}`;

    const cardActions = document.createElement("div");
    cardActions.className = "card-actions";

    const viewButton = document.createElement("button");
    viewButton.className = "card-action-btn";
    viewButton.innerHTML = `<i class="fa-solid fa-eye"></i> View Details`;
    viewButton.addEventListener("click", (event) => {
      event.stopPropagation();
      showListDetails(key);
    });

    cardActions.appendChild(viewButton);

    card.appendChild(statusIndicator);
    cardContent.appendChild(header);
    cardContent.appendChild(counter);
    cardContent.appendChild(subheader);
    cardContent.appendChild(updated);
    card.appendChild(cardContent);
    card.appendChild(cardActions);

    card.addEventListener("click", () => showListDetails(key));
    container.appendChild(card);

    setTimeout(() => {
      card.style.animation = "fadeInUp 0.5s ease-out forwards";
    }, 10);

    delay += ANIMATION_STAGGER_DELAY;
  }
}

function showListDetails(listName) {
  const listDetailsModal = document.getElementById("list-details-modal");
  const listDetailsContent = document.getElementById("list-details-content");

  listDetailsContent.innerHTML = `
    <span id="list-details-close" class="list-details-close">&times;</span>
    <h2>${listName}</h2>
    <div class="modal-actions">
      <button id="toggleListBtn" class="toggle-btn" onclick="toggleBlocklist('${listName}')"><i class="fa-solid fa-power-off"></i> Toggle</button>
      <button id="updateListBtn" class="update-btn"><i class="fa-solid fa-sync"></i> Update</button>
      <button id="removeListBtn" class="remove-btn"><i class="fa-solid fa-trash"></i> Remove List</button>
    </div>
    <div class="domains-loading">
      <div class="spinner"></div>
      <p>Loading domains...</p>
    </div>
  `;

  listDetailsModal.style.display = "flex";
  listDetailsModal.style.backgroundColor = "rgba(0, 0, 0, 0)";
  setTimeout(() => {
    listDetailsModal.style.backgroundColor = "rgba(0, 0, 0, 0.7)";
  }, 10);

  document.getElementById("list-details-close").onclick = () => {
    closeModalWithAnimation(listDetailsModal);
  };

  fetch(`/api/getDomainsForList?list=${listName}`)
    .then((response) => response.json())
    .then((data) => {
      const domains = data.domains;

      const loadingEl = document.querySelector(".domains-loading");
      if (loadingEl) loadingEl.remove();

      const table = document.createElement("table");
      table.id = "domains-table";
      table.className = "display";
      table.style.opacity = "0";
      table.style.transform = "translateY(20px)";

      const thead = document.createElement("thead");
      thead.innerHTML = `
        <tr>
          <th>Domain</th>
          <th>Toggle</th>
        </tr>
      `;

      const tbody = document.createElement("tbody");
      domains.forEach((domain) => {
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

      setTimeout(() => {
        table.style.transition = "all 0.5s ease-out";
        table.style.opacity = "1";
        table.style.transform = "translateY(0)";
      }, 50);
    });

  document
    .getElementById("removeListBtn")
    .addEventListener("click", async () => {
      animateButton(document.getElementById("removeListBtn"));

      try {
        const response = await fetch(`/api/list?name=${listName}`, {
          method: "DELETE",
        });

        if (response.ok) {
          showInfoNotification("List removed successfully!");
          closeModalWithAnimation(listDetailsModal);
          getLists();
        } else {
          const errorData = await response.json();
          showErrorNotification(errorData.error);
        }
      } catch (error) {
        showErrorNotification("Error removing list.");
      }
    });
}

async function toggleBlocklist(blocklistName) {
  await PostRequest(
    "/toggleBlocklist",
    JSON.stringify({ name: blocklistName })
  ).then((response) => {
    showInfoNotification(response.message);
  });
}

updateCustomBtn.addEventListener("click", () => {
  updateCustomList();
  animateButton(updateCustomBtn);
});

function updateCustomList() {
  modalUpdateCustom.style.display = "flex";
  modalUpdateCustom.style.backgroundColor = "rgba(0, 0, 0, 0)";
  setTimeout(() => {
    modalUpdateCustom.style.backgroundColor = "rgba(0, 0, 0, 0.7)";
  }, 10);

  saveListBtn.addEventListener("click", () => {
    saveCustom();
    animateButton(saveListBtn);
  });
}

function closeModal() {
  closeModalWithAnimation(modalUpdateCustom);
}

modalClose.onclick = closeModal;

async function saveCustom() {
  const validListPattern = /^[a-z0-9.:]+$/; // Only allow numbers, dots, and colons
  let containsInvalidList = false;

  const newLists = listsTextArea.value
    .split("\n")
    .map((list) => list.trim())
    .filter((list) => {
      if (list === "") return false;
      if (!validListPattern.test(list)) {
        showErrorNotification(
          `Invalid list: "${list}" contains invalid characters.`
        );
        containsInvalidList = true;
        return false;
      }
      return true;
    });

  if (!containsInvalidList) {
    response = await PostRequest(
      "/custom",
      JSON.stringify({ domains: newLists })
    );
    closeModal();
    showInfoNotification("Updated custom list!");

    getLists();
  }
}

function showNotification(message, type) {
  const notification = document.createElement("div");
  notification.className = `notification ${type}`;
  notification.textContent = message;

  notification.style.transform = "translateY(20px)";
  notification.style.opacity = "0";

  document.body.appendChild(notification);

  setTimeout(() => {
    notification.style.transition = "all 0.3s ease-out";
    notification.style.transform = "translateY(0)";
    notification.style.opacity = "1";
  }, 10);

  setTimeout(() => {
    notification.style.transform = "translateY(-20px)";
    notification.style.opacity = "0";

    setTimeout(() => {
      notification.remove();
    }, 300);
  }, 4000);
}

function showInfoNotification(message) {
  showNotification(message, "info");
}

function showErrorNotification(message) {
  showNotification(message, "error");
}

document.addEventListener("DOMContentLoaded", () => {
  initAnimations();
  addButtonAnimations();
  getLists();
});
