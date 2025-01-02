window.onload = async function () {
  const notificationData = localStorage.getItem("persistentNotification");

  authenticationRequired = await GetRequest("/authentication");
  if (authenticationRequired.disabled) {
    window.location.href = "/index.html";
  }

  if (notificationData) {
    const { headerMessage, type, message } = JSON.parse(notificationData);
    showNotification(headerMessage, type, message);

    localStorage.removeItem("persistentNotification");
  }
};

document.getElementById("login-form").addEventListener("submit", function (e) {
  e.preventDefault();
  const username = document.getElementById("username").value;
  const password = document.getElementById("password").value;

  fetch(GetServerIP() + "/login", {
    method: "POST",
    body: JSON.stringify({ username: username, password: password }),
  })
    .then((response) => {
      if (response.ok) {
        window.location.href = "/index.html";
      } else {
        document.getElementById("error-message").style.display = "block";
      }
    })
    .catch((error) => console.error("Error:", error));
});
