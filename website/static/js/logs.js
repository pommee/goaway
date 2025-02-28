const enableLiveQueries = document.getElementById("enableLiveQueries");

var logTable;
let socket;
let liveQueriesEnabled = false;

enableLiveQueries.addEventListener("change", () => {
  liveQueriesEnabled = enableLiveQueries.checked;
});

socket = new WebSocket(
  `ws://${GetServerIP().replace("http://", "")}/api/liveQueries`
);

socket.onmessage = function (event) {
  if (liveQueriesEnabled) {
    logTable.row.add(JSON.parse(event.data)).draw();
  }
};

socket.onclose = function (event) {
  showInfoNotification("Websocket connection was closed.");
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

function prepareRequestData(d) {
  const page = Math.floor(d.start / d.length) + 1;
  return {
    draw: d.draw,
    page: page,
    pageSize: d.length,
    search: d.search.value,
    sortColumn: d.columns[d.order[0].column].data,
    sortDirection: d.order[0].dir,
  };
}

function renderStatusAndResponseTime(data) {
  var status = data.blocked
    ? "Blocked"
    : data.cached
    ? "OK (cached)"
    : "OK (forwarded)";
  status = `${status} ${data.status}`;
  const responseTime = (data.responseTimeNS / 1_000_000).toFixed(2);
  return `${status}<br>${responseTime} ms`;
}

function renderDomain(domain) {
  return `
  <div class="domain">
    <span class="domain-ip" data-tooltip="${domain}">${domain}</span>
  </div>
  `;
}

function renderIP(data) {
  let ipList = "";
  let firstIP = "";
  if (data !== null) {
    ipList = data.join("\n");
    firstIP = data[0];
  }
  return `
  <div class="ip-container">
    <span class="ip" data-tooltip="${ipList}">${firstIP}</span>
  </div>
  `;
}

function renderToggleButton(data) {
  const toggleBtnTxt = data.blocked ? "Whitelist" : "Blacklist";
  const buttonClass = data.blocked ? "blocked-true" : "blocked-false";
  return `<button class="toggle-button ${buttonClass}" data-blocked="${data.blocked}" data-domain="${data.domain}">${toggleBtnTxt}</button>`;
}

async function handleToggleClick(event) {
  const domain = $(event.target).data("domain");
  const currentlyBlocked = $(event.target).data("blocked");
  const newBlockedStatus = !currentlyBlocked;
  $(event.target).data("blocked", newBlockedStatus);

  try {
    const blockReq = await $.get(
      `/api/updateBlockStatus?domain=${domain}&blocked=${newBlockedStatus}`
    );
    showInfoNotification(blockReq.message);

    $(event.target).text(newBlockedStatus ? "Whitelist" : "Blacklist");
    const buttonClass = newBlockedStatus ? "blocked-true" : "blocked-false";
    $(event.target)
      .removeClass("blocked-true blocked-false")
      .addClass(buttonClass);

    const row = $(`#log-table tr:contains(${domain})`);
    newBlockedStatus
      ? row.addClass("wasBlocked")
      : row.removeClass("wasBlocked");
  } catch (error) {
    console.error("Error updating block status:", error);
    showErrorNotification("Failed to update block status. Please try again.");
  }
}

async function handleClearLogsClick() {
  const modal = document.getElementById("clear-logs-modal");
  const closeModal = document.getElementsByClassName("close")[0];
  const confirmButton = document.getElementById("confirm-clear-logs-button");
  const cancelButton = document.getElementById("cancel-clear-logs-button");

  modal.style.display = "block";

  closeModal.onclick = function () {
    modal.style.display = "none";
  };

  cancelButton.onclick = function () {
    modal.style.display = "none";
  };

  confirmButton.onclick = async function () {
    try {
      await $.ajax({
        url: "/api/queries",
        type: "DELETE",
      });

      logTable.destroy();
      await initializeLogTable();
      showInfoNotification("Logs cleared successfully.");
    } catch (error) {
      console.error("Error clearing logs:", error);
      showErrorNotification("Failed to clear logs. Please try again.");
    } finally {
      modal.style.display = "none";
    }
  };

  window.onclick = function (event) {
    if (event.target == modal) {
      modal.style.display = "none";
    }
  };
}

async function initializeLogTable() {
  $(document).ready(function () {
    logTable = $("#log-table").DataTable({
      processing: true,
      serverSide: true,
      ajax: {
        url: "/api/queries",
        type: "GET",
        data: function (d) {
          return prepareRequestData(d);
        },
        dataSrc: "details",
        error: function (xhr) {
          if (xhr.status === 401) {
            window.location.href = "/login.html";
          } else {
            console.error("Failed to load data:", xhr.statusText);
          }
        },
      },
      columns: [
        { data: "timestamp", render: formatTimestamp },
        { data: "domain", render: renderDomain },
        { data: "ip", render: renderIP },
        {
          data: "client",
          render: (data) => `${data.Name || "Unknown"} | ${data.IP || "N/A"}`,
        },
        { data: null, render: renderStatusAndResponseTime },
        { data: "queryType" },
        { data: null, render: renderToggleButton },
      ],
      order: [[0, "desc"]],
      drawCallback: function () {
        $("#log-table tbody tr").each(function () {
          const row = $(this);
          const blockedStatus = row.find("td").eq(4).text().includes("Blocked");
          blockedStatus
            ? row.addClass("wasBlocked")
            : row.removeClass("wasBlocked");
        });

        $(".toggle-button").each(function () {
          const button = $(this);
          const blocked = button.data("blocked");
          const buttonClass = blocked ? "blocked-true" : "blocked-false";
          button
            .removeClass("blocked-true blocked-false")
            .addClass(buttonClass);
        });
      },
    });
  });
}

document.addEventListener("DOMContentLoaded", async () => {
  await initializeLogTable();
  $(document).on("click", ".toggle-button", handleToggleClick);
  $(document).on("click", "#clear-logs-button", handleClearLogsClick);
});
