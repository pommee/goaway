// Table initialization and configuration
const TABLE_CONFIG = {
  paging: true,
  pageLength: 10,
  columns: [
    { data: null, render: (data) => data },
    {
      data: null,
      render: (data) => `<button class="update-block-status" data-domain="${data}">unblock</button>`,
    },
  ],
};

// DOM Elements
const elements = {
  tableContainer: () => document.getElementById("domains-table-container"),
  modal: () => document.getElementById("add-domain-modal"),
  domainInput: () => document.getElementById("domain-name"),
};

// API endpoints
const API = {
  base: GetServerIP(),
  domains: "/domains",
  updateStatus: (domain, blocked) => `/updateBlockStatus?domain=${domain}&blocked=${blocked}`,
};

// Table management
class DomainTable {
  constructor() {
    this.table = null;
    this.isLoading = false;
  }

  init() {
    if (this.table) {
      this.table.destroy();
    }
    this.table = $("#domains-table").DataTable(TABLE_CONFIG);
    this.attachEventHandlers();
  }

  attachEventHandlers() {
    $("#domains-table").on("click", ".update-block-status", async (e) => {
      const button = $(e.target);
      const domain = button.data("domain");
      await this.unblockDomain(domain, button);
    });
  }

  async unblockDomain(domain, button) {
    try {
      button.prop("disabled", true);
      const response = await fetch(`${API.base}${API.updateStatus(domain, false)}`);
      const data = await response.json();

      if (data.domain) {
        this.table.row(button.closest("tr")).remove().draw();
      }
    } catch (error) {
      showWarningNotification("Failed to unblock domain");
      button.prop("disabled", false);
    }
  }

  async addDomain(domain) {
    try {
      const response = await fetch(`${API.base}${API.updateStatus(domain, true)}`);
      const data = await response.json();

      if (data.error) {
        showInfoNotification(data.error);
        return;
      }

      if (data.domain) {
        this.table.row
          .add([data.domain, `<button class="update-block-status" data-domain="${data.domain}">unblock</button>`])
          .draw();
        showInfoNotification("Added domain:", data.domain);
      }
    } catch (error) {
      showErrorNotification("Failed to add domain");
    }
  }
}

class LoadingManager {
  static show() {
    elements.tableContainer().innerHTML = `
      <div class="loading-spinner-container">
        <div class="loading-spinner"></div>
      </div>
    `;
  }

  static hide() {
    elements.tableContainer().innerHTML = `
      <table id="domains-table" class="display">
        <thead>
          <tr>
            <th>Domains</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody></tbody>
      </table>
    `;
  }
}

class ModalManager {
  static open() {
    elements.modal().style.display = "block";
  }

  static close() {
    elements.modal().style.display = "none";
    elements.domainInput().value = "";
  }
}

class DomainManager {
  constructor() {
    this.domainTable = new DomainTable();
  }

  async initialize() {
    try {
      LoadingManager.show();
      const response = await fetch(`${API.base}${API.domains}`);
      if (!response.ok) throw new Error("Network response failed");

      const data = await response.json();
      LoadingManager.hide();
      this.domainTable.init();
      this.domainTable.table.rows.add(data.domains).draw();
    } catch (error) {
      LoadingManager.hide();
      showErrorNotification("Failed to fetch domains");
    }
  }

  setupEventListeners() {
    document.getElementById("confirm-add-domain-btn").addEventListener("click", () => {
      const domain = elements.domainInput().value;
      if (domain) {
        this.domainTable.addDomain(domain);
        ModalManager.close();
      } else {
        showWarningNotification("Please enter a valid domain name");
      }
    });

    document.getElementById("cancel-btn").addEventListener("click", ModalManager.close);
  }
}

// Initialize application
document.addEventListener("DOMContentLoaded", () => {
  const styleSheet = document.createElement("style");
  document.head.appendChild(styleSheet);

  const app = new DomainManager();
  app.initialize();
  app.setupEventListeners();

  // Expose for global access
  window.openAddDomainModal = ModalManager.open;
  window.closeAddDomainModal = ModalManager.close;
});
