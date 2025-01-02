async function getLogs() {
  data = await GetRequest("/queriesData");
  populateLogTable(data);
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

function populateLogTable(logs) {
  $(document).ready(function () {
    logs.details.forEach((detail) => {
      const blockedClass = detail.blocked ? 'class="wasBlocked"' : "";
      const formattedTimestamp = formatTimestamp(detail.timestamp);

      $("#log-table tbody").append(
        `<tr>
            <td>${formattedTimestamp}</td>
            <td>${detail.domain}</td>
            <td ${blockedClass}>${detail.blocked}</td>
            <td>${detail.client.Name}  |  ${detail.client.IP}</td>
        </tr>`,
      );
    });

    $("#log-table").DataTable({
      order: [[0, "desc"]],
      response: true,
    });
  });
}

document.addEventListener("DOMContentLoaded", () => {
  getLogs();
});
