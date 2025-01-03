const notifications = [];

if (!GetServerIP()) {
  var serverIP = document.location.origin;
  localStorage.setItem("serverIP", serverIP);
}

function GetServerIP() {
  return localStorage.getItem("serverIP");
}

function GetRequest(url) {
  return new Promise((resolve, reject) => {
    fetch(GetServerIP() + url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((response) => {
        if (response.status >= 400) {
          if (response.status === 401) {
            showPersistentNotification(
              "Info",
              "info",
              "You have been logged out. Please log in again.",
            );
            localStorage.clear();
            window.location.href = "/login.html";
          }
          throw new Error("Network response was not ok");
        }
        return response.json();
      })
      .then((data) => {
        resolve(data);
      })
      .catch((error) => {
        reject(error);
      });
  });
}

function PostRequest(url, body) {
  return new Promise((resolve, reject) => {
    fetch(GetServerIP() + url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: body,
    })
      .then((response) => {
        if (response.status >= 400) {
          if (response.status === 401) {
            showPersistentNotification(
              "Info",
              "info",
              "You have been logged out. Please log in again.",
            );
            localStorage.clear();
            window.location.href = "/login.html";
          }
          throw new Error("Network response was not ok");
        }
        return response.json();
      })
      .then((data) => {
        resolve(data);
      })
      .catch((error) => {
        reject(error);
      });
  });
}

function DeleteRequest(url) {
  return new Promise((resolve, reject) => {
    fetch(GetServerIP() + url, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((response) => {
        if (response.status >= 400) {
          if (response.status === 401) {
            showPersistentNotification(
              "Info",
              "info",
              "You have been logged out. Please log in again.",
            );
            localStorage.clear();
            window.location.href = "/login.html";
          }
          throw new Error("Network response was not ok");
        }
        return response.json();
      })
      .then((data) => {
        resolve(data);
      })
      .catch((error) => {
        reject(error);
      });
  });
}

function Logout() {
  localStorage.clear();
  window.location.href = "/login.html";
}

function showNotification(headerMessage, type, ...message) {
  const notification = document.createElement("div");
  notification.classList.add("notification", type);

  const header = document.createElement("div");
  header.classList.add("notification-header");
  header.textContent = headerMessage || "Notification";

  const messageSection = document.createElement("div");
  messageSection.classList.add("notification-message");
  messageSection.textContent = message.join(" ");

  notification.appendChild(header);
  notification.appendChild(messageSection);

  document.body.appendChild(notification);

  const offset = notifications.reduce(
    (acc, el) => acc + el.offsetHeight + 10,
    0,
  );
  notification.style.bottom = `${10 + offset}px`;
  notifications.push(notification);

  setTimeout(() => {
    notification.style.opacity = "0";
  }, 5000);

  setTimeout(() => {
    notification.remove();
    notifications.splice(notifications.indexOf(notification), 1);
    notifications.forEach((el, i) => {
      const newOffset = notifications
        .slice(0, i)
        .reduce((acc, el) => acc + el.offsetHeight + 10, 0);
      el.style.bottom = `${10 + newOffset}px`;
    });
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

function showPersistentNotification(headerMessage, type, ...message) {
  const notificationData = {
    headerMessage: headerMessage || "Notification",
    type: type,
    message: message.join(" "),
  };

  localStorage.setItem(
    "persistentNotification",
    JSON.stringify(notificationData),
  );

  showNotification(headerMessage, type, ...message);
}
