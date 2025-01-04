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
  const status = data.blocked ? "Blocked" : data.cached ? "OK (cached)" : "OK (forwarded)";
  const responseTime = (data.responseTimeNS / 1_000_000).toFixed(2);
  return `${status}<br>${responseTime} ms`;
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
    const blockReq = await $.get(`/updateBlockStatus?domain=${domain}&blocked=${newBlockedStatus}`);
    showInfoNotification(blockReq.message);

    $(event.target).text(newBlockedStatus ? "Whitelist" : "Blacklist");
    const buttonClass = newBlockedStatus ? "blocked-true" : "blocked-false";
    $(event.target).removeClass("blocked-true blocked-false").addClass(buttonClass);

    const row = $(`#log-table tr:contains(${domain})`);
    newBlockedStatus ? row.addClass("wasBlocked") : row.removeClass("wasBlocked");
  } catch (error) {
    console.error("Error updating block status:", error);
    showErrorNotification("Failed to update block status. Please try again.");
  }
}

async function initializeLogTable() {
  $(document).ready(function () {
    $("#log-table").DataTable({
      processing: true,
      serverSide: true,
      ajax: {
        url: "/queriesData",
        type: "GET",
        data: function (d) {
          return prepareRequestData(d);
        },
        dataSrc: "details",
      },
      columns: [
        { data: "timestamp", render: formatTimestamp },
        { data: "domain" },
        { data: "client", render: (data) => `${data.Name || "Unknown"} | ${data.IP || "N/A"}` },
        { data: null, render: renderStatusAndResponseTime },
        { data: null, render: renderToggleButton },
      ],
      order: [[0, "desc"]],
      drawCallback: function () {
        $("#log-table tbody tr").each(function () {
          const row = $(this);
          const blockedStatus = row.find("td").eq(3).text().includes("Blocked");
          blockedStatus ? row.addClass("wasBlocked") : row.removeClass("wasBlocked");
        });

        $(".toggle-button").each(function () {
          const button = $(this);
          const blocked = button.data("blocked");
          const buttonClass = blocked ? "blocked-true" : "blocked-false";
          button.removeClass("blocked-true blocked-false").addClass(buttonClass);
        });
      },
    });
  });
}

document.addEventListener("DOMContentLoaded", async () => {
  await initializeLogTable();
  $(document).on("click", ".toggle-button", handleToggleClick);
});
