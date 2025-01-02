async function getLogs() {
  data = await GetRequest("/queriesData");
  await populateLogTable(data);
}

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

async function populateLogTable(logs) {
  $(document).ready(function () {
    logs.details.forEach((detail) => {
      const blockedClass = detail.blocked ? 'class="wasBlocked"' : "";
      const formattedTimestamp = formatTimestamp(detail.timestamp);
      const toggleBtnTxt = detail.blocked == true ? "Whitelist" : "Blacklist";
      let status;

      status = detail.blocked
        ? "Blocked"
        : detail.cached
          ? "OK (cached)"
          : "OK (forwarded)";
      status += "<br>" + (detail.responseTimeNS / 1000000).toFixed(2) + " ms";

      $("#log-table tbody").append(
        `<tr id="log-${detail.domain} class="${blockedClass}>
            <td>${formattedTimestamp}</td>
            <td>${detail.domain}</td>
            <td>${detail.client.Name}  |  ${detail.client.IP}</td>
            <td ${blockedClass}>${status}</td>
            <td><button class="toggle-button blocked-${detail.blocked}" data-blocked="${detail.blocked}" data-domain="${detail.domain}">${toggleBtnTxt}</button></td>
        </tr>`,
      );
    });

    $("#log-table").DataTable({
      order: [[0, "desc"]],
      response: true,
    });

    $(".toggle-button").on("click", async function () {
      const domain = $(this).data("domain");
      const currentlyBlocked = $(this).data("blocked");

      const newBlockedStatus = !currentlyBlocked;
      $(this).data("blocked", newBlockedStatus);

      blockReq = await GetRequest(
        "/updateBlockStatus?domain=" + domain + "&blocked=" + newBlockedStatus,
      );
      showInfoNotification(blockReq.message);

      const row = $(`#log-${domain}`);
      row
        .find("td")
        .eq(2)
        .text(newBlockedStatus ? "Blacklist" : "Whitelist");
      row.find("td").eq(2).toggleClass("wasBlocked", newBlockedStatus);
    });
  });
}

document.addEventListener("DOMContentLoaded", () => {
  getLogs();
});
