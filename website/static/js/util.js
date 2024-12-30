const serverIp = "<?= .ServerIP ?>";

function GetServerIP() {
  return localStorage.getItem("serverIP");
}

function showNotification(headerMessage, type, ...message) {
  const notification = document.createElement("div");
  notification.classList.add("notification", type);

  const header = document.createElement("div");
  header.classList.add("notification-header");
  header.textContent =
    headerMessage == undefined ? "Notification" : headerMessage;

  const messageSection = document.createElement("div");
  messageSection.classList.add("notification-message");
  messageSection.textContent = message;

  notification.appendChild(header);
  notification.appendChild(messageSection);

  document.body.appendChild(notification);

  setTimeout(() => {
    notification.style.opacity = "0";
  }, 5000);

  setTimeout(() => {
    notification.remove();
  }, 6000);
}

function showInfoNotification(...message) {
  showNotification("Info", "info", ...message);
}

function showErrorNotification(...message) {
  showNotification("Error", "error", ...message);
}

function showWarningNotification(...message) {
  showNotification("Warning", "warning", ...message);
}
