var resolutionTable;

function renderDomain(domain) {
  return `
  <div class="domain">
    <span class="domain-ip" data-tooltip="${domain}">${domain}</span>
  </div>
  `;
}

function renderIP(data) {
  return `
  <div class="ip-container">
    <span class="ip" data-tooltip="${data}">${data}</span>
  </div>
  `;
}

async function initializeResolutionTable() {
  $(document).ready(function () {
    resolutionTable = $("#resolution-table").DataTable({
      processing: true,
      serverSide: true,
      ajax: {
        url: "/api/resolutions",
        type: "GET",
        dataSrc: "resolutions",
        error: function (xhr) {
          if (xhr.status === 401) {
            Logout();
          } else {
            console.error("Failed to load data:", xhr.statusText);
          }
        },
      },
      columns: [
        { data: "Domain", render: renderDomain },
        { data: "IP", render: renderIP },
      ],
      order: [[0, "desc"]],
    });
  });
}

document.addEventListener("DOMContentLoaded", async () => {
  await initializeResolutionTable();

  const form = document.getElementById("resolutionEntryForm");
  const ipInput = document.getElementById("ipAddress");
  const domainInput = document.getElementById("domainName");
  const ipError = document.getElementById("ipError");
  const domainError = document.getElementById("domainError");
  const submitButton = document.getElementById("submitResolution");

  function isValidIP(ip) {
    const ipPattern =
      /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    return ipPattern.test(ip);
  }

  function isValidDomain(domain) {
    return domain.trim() !== "";
  }

  ipInput.addEventListener("input", function () {
    if (this.value && !isValidIP(this.value)) {
      ipError.textContent = "Please enter a valid IP address";
      ipError.style.display = "block";
      this.classList.add("error");
    } else {
      ipError.style.display = "none";
      this.classList.remove("error");
    }
  });

  domainInput.addEventListener("input", function () {
    if (this.value && !isValidDomain(this.value)) {
      domainError.textContent = "Please enter a valid domain";
      domainError.style.display = "block";
      this.classList.add("error");
    } else {
      domainError.style.display = "none";
      this.classList.remove("error");
    }
  });

  form.addEventListener("submit", function (e) {
    e.preventDefault();

    const ip = ipInput.value.trim();
    const domain = domainInput.value.trim();

    let isValid = true;

    if (!isValidIP(ip)) {
      ipError.textContent = "Please enter a valid IP address";
      ipError.style.display = "block";
      ipInput.classList.add("error");
      isValid = false;
    }

    if (!isValidDomain(domain)) {
      domainError.textContent = "Please enter a valid domain";
      domainError.style.display = "block";
      domainInput.classList.add("error");
      isValid = false;
    }

    if (!isValid) return;

    submitButton.disabled = true;
    submitButton.textContent = "Adding...";

    const requestData = {
      ip: ip,
      domain: domain,
    };

    fetch(GetServerIP() + "/api/resolution", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestData),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`Failed to add ${domain} (${ip})`);
        }
        return true;
      })
      .then(() => {
        showInfoNotification("Added new entry!");
        form.reset();

        if (resolutionTable) {
          resolutionTable.ajax.reload();
        }
      })
      .catch((error) => {
        showErrorNotification(error.message || "Failed to add entry");
      })
      .finally(() => {
        submitButton.disabled = false;
        submitButton.textContent = "Add Entry";
      });
  });
});
