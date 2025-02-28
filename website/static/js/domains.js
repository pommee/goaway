function prepareRequestData(d) {
  const page = Math.floor(d.start / d.length) + 1;
  return {
    draw: d.draw,
    page: page,
    pageSize: d.length,
    search: d.search.value,
  };
}

function renderStatusAndResponseTime(data) {
  const status = data.blocked
    ? "Blocked"
    : data.cached
      ? "OK (cached)"
      : "OK (forwarded)";
  const responseTime = (data.responseTimeNS / 1_000_000).toFixed(2);
  return `${status}<br>${responseTime} ms`;
}

function renderToggleButton(data) {
  return `<button class="toggle-button blocked-true" data-blocked="true" data-domain="${data.domain}">Whitelist</button>`;
}

async function handleToggleClick(event) {
  const domain = $(event.target).data("domain");
  const currentlyBlocked = $(event.target).data("blocked");
  const newBlockedStatus = !currentlyBlocked;
  $(event.target).data("blocked", newBlockedStatus);

  try {
    const blockReq = await $.get(
      `/api/updateBlockStatus?domain=${domain}&blocked=${newBlockedStatus}`,
    );
    showInfoNotification(blockReq.message);

    $(event.target).text(newBlockedStatus ? "Whitelist" : "Blacklist");
    const buttonClass = newBlockedStatus ? "blocked-true" : "blocked-false";
    $(event.target)
      .removeClass("blocked-true blocked-false")
      .addClass(buttonClass);

    const row = $(`#domains-table tr:contains(${domain})`);
    newBlockedStatus
      ? row.addClass("wasBlocked")
      : row.removeClass("wasBlocked");
  } catch (error) {
    console.error("Error updating block status:", error);
    showErrorNotification("Failed to update block status. Please try again.");
  }
}

async function initializeLogTable() {
  $(document).ready(function () {
    $("#domains-table").DataTable({
      processing: true,
      serverSide: true,
      ajax: {
        url: "/api/domains",
        type: "GET",
        data: function (d) {
          return prepareRequestData(d);
        },
        dataSrc: function (json) {
          const domains = json.domains || [];
          return domains.map((domain) => ({ domain }));
        },
        error: function (xhr) {
          if (xhr.status === 401) {
            window.location.href = "/login.html";
          } else {
            console.error("Failed to load data:", xhr.statusText);
          }
        },
      },
      columns: [{ data: "domain" }, { data: null, render: renderToggleButton }],
      order: [[0, "desc"]],
      drawCallback: function () {
        $("#domains-table tbody tr").each(function () {
          const row = $(this);
          const blockedStatus = row.find("td").eq(3).text().includes("Blocked");
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

function openAddDomainModal() {
  $(".modal-add-domain").show();
}

function closeAddDomainModal() {
  $(".modal-add-domain").hide();
}

async function handleAddDomain() {
  const domain = $("#domain-name").val();
  if (domain) {
    try {
      const blockReq = await $.get(
        `/api/updateBlockStatus?domain=${domain}&blocked=true`,
      );
      showInfoNotification(blockReq.message);
      closeAddDomainModal();
    } catch (error) {
      console.error("Error adding domain:", error);
      showErrorNotification("Failed to add domain. Please try again.");
    }
  } else {
    showErrorNotification("Please enter a domain.");
  }
}

document.addEventListener("DOMContentLoaded", async () => {
  await initializeLogTable();
  $(document).on("click", ".toggle-button", handleToggleClick);
  $("#add-domain-btn").on("click", openAddDomainModal);
  $("#cancel-btn").on("click", closeAddDomainModal);
  $("#confirm-add-domain-btn").on("click", handleAddDomain);
});
