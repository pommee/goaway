const TABLE_CONFIG = {
  paging: true,
  pageLength: 10,
  serverSide: true,
  processing: true,
  columns: [
    { data: null, render: (data) => data },
    {
      data: null,
      render: (data) => `<button class="update-block-status" data-domain="${data}">unblock</button>`,
    },
  ],
  ajax: function (data, callback, settings) {
    const page = Math.floor(settings.start / settings.length) + 1;
    const pageSize = settings.length;
    const search = $("#domains-table_filter input").val();

    $.get(`${API.base}${API.domains}?page=${page}&pageSize=${pageSize}&search=${search}`, function (response) {
      callback({
        draw: settings.draw,
        recordsTotal: response.total,
        recordsFiltered: response.total,
        data: response.domains,
      });
    });
  },
};

// DOM Elements
const elements = {
  tableContainer: () => document.getElementById("domains-table-container"),
  modal: () => document.getElementById("add-domain-modal"),
  domainInput: () => document.getElementById("domain-name"),
  searchInput: () => document.getElementById("domain-search"),
  addDomainBtn: () => document.getElementById("add-domain-btn"),
  confirmAddDomainBtn: () => document.getElementById("confirm-add-domain-btn"),
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

    $("#domains-table_filter input").on("input", () => {
      this.table.ajax.reload();
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
      const response = await fetch(`${API.base}${API.updateStatus(domain, true)}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

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
  constructor(domainManager) {
    this.domainManager = domainManager;
  }

  open() {
    elements.modal().style.display = "block";
  }

  close() {
    elements.modal().style.display = "none";
    elements.domainInput().value = "";
  }

  onAddDomainClick() {
    const domain = elements.domainInput().value;
    if (domain) {
      this.domainManager.domainTable.addDomain(domain);
      this.close();
    } else {
      showWarningNotification("Please enter a valid domain.");
    }
  }
}

class DomainManager {
  constructor() {
    this.domainTable = new DomainTable();
    this.modalManager = new ModalManager(this);
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
    elements.addDomainBtn().addEventListener("click", () => {
      this.modalManager.open();
    });

    elements.confirmAddDomainBtn().addEventListener("click", () => {
      this.modalManager.onAddDomainClick();
    });
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const styleSheet = document.createElement("style");
  document.head.appendChild(styleSheet);

  const app = new DomainManager();
  app.initialize();
  app.setupEventListeners();

  // Expose for global access
  window.openAddDomainModal = app.modalManager.open.bind(app.modalManager);
  window.closeAddDomainModal = app.modalManager.close.bind(app.modalManager);
});
