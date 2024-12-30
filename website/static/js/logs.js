function getLogs() {
  fetch("http://localhost:8080/queriesData")
    .then(function (response) {
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then(function (data) {
      populateLogTable(data);
    })
    .catch(function (error) {
      console.error("Failed to fetch logs:", error);
    });
}

function populateLogTable(logs) {
  $(document).ready(function () {
    logs.details.forEach((detail) => {
      const blockedClass = detail.blocked ? 'class="wasBlocked"' : "";

      $("#log-table tbody").append(
        `<tr>
            <td>${detail.timestamp}</td>
            <td>${detail.domain}</td>
            <td ${blockedClass}>${detail.blocked}</td>
        </tr>`,
      );
    });

    $("#log-table").DataTable();
  });
}

document.addEventListener("DOMContentLoaded", () => {
  getLogs();
});
