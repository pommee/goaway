const inputs = document.querySelectorAll(
  "#main input:not(#currentPassword):not(#newPassword):not(#confirmPassword), #main select"
);
const toggleTheme = document.getElementById("toggleTheme");
const savePopup = document.getElementById("save-popup");
const saveButton = document.getElementById("save-btn");
const dismissButton = document.getElementById("dismiss-btn");
const confirmPasswordInput = document.getElementById("confirmPassword");
const changePasswordButton = document.getElementById("change-password-btn");
const passwordModal = document.getElementById("password-modal");
const currentPasswordInput = document.getElementById("currentPassword");
const newPasswordInput = document.getElementById("newPassword");
const savePasswordButton = document.getElementById("save-password-btn");
const cancelPasswordButton = document.getElementById("cancel-password-btn");
const passwordError = document.getElementById("password-error");

document.addEventListener("DOMContentLoaded", () => {
  initializeSettings();
  let isModified = false;

  inputs.forEach((input) => {
    input.addEventListener("input", () => {
      isModified = true;
      showPopup();
    });
  });

  const root = document.documentElement;
  const savedTheme = localStorage.getItem("theme") || "light";
  root.style.colorScheme = savedTheme;
  toggleTheme.checked = savedTheme === "light";

  toggleTheme.addEventListener("change", () => {
    const newTheme = root.style.colorScheme === "light" ? "dark" : "light";
    root.style.colorScheme = newTheme;
    toggleTheme.checked = newTheme === "light";
    localStorage.setItem("theme", newTheme);
  });

  confirmPasswordInput.addEventListener("input", validatePasswords);

  changePasswordButton.addEventListener("click", () => {
    passwordModal.style.display = "block";
  });

  savePasswordButton.addEventListener("click", () => {
    if (validatePasswords()) {
      savePassword();
      passwordModal.style.display = "none";
      passwordError.textContent = "";
      Logout();
    } else {
      passwordError.textContent = "Passwords do not match.";
    }
  });

  cancelPasswordButton.addEventListener("click", () => {
    passwordModal.style.display = "none";
  });

  newPasswordInput.addEventListener("input", validatePasswords);
  confirmPasswordInput.addEventListener("input", validatePasswords);

  function showPopup() {
    if (isModified) {
      savePopup.style.display = "block";
    }
  }

  function hidePopup() {
    savePopup.style.display = "none";
  }

  saveButton.addEventListener("click", async () => {
    await saveSettings();
    isModified = false;
    hidePopup();
  });

  dismissButton.addEventListener("click", () => {
    isModified = false;
    hidePopup();
  });
});

async function initializeSettings() {
  let settings;
  try {
    settings = await getSettings();
  } catch (error) {
    showErrorNotification(error);
    return;
  }

  const cacheTTLInSeconds = settings.dns.CacheTTL;
  document.getElementById("cacheTTL").value = cacheTTLInSeconds;
  document.getElementById("logLevel").selectedIndex = settings.dns.LogLevel;
  document.getElementById("disableLogging").checked =
    settings.dns.LoggingDisabled;
  document.getElementById("statisticsRetention").value =
    settings.dns.StatisticsRetention;
}

async function getSettings() {
  const settings = await GetRequest("/settings");
  return settings;
}

async function saveSettings() {
  const settings = {};

  document
    .querySelectorAll(".setting-item input, .setting-item select")
    .forEach((input) => {
      if (input.type === "checkbox") {
        settings[input.id] = input.checked;
      } else {
        settings[input.id] = input.value;
      }
    });

  await PostRequest("/settings", JSON.stringify(settings));
}

function validatePasswords() {
  const password = newPasswordInput.value;
  const confirmPassword = confirmPasswordInput.value;
  if (password !== confirmPassword) {
    confirmPasswordInput.setCustomValidity("Passwords do not match");
    return false;
  } else {
    confirmPasswordInput.setCustomValidity("");
    return true;
  }
}

function savePassword() {
  const settings = {
    currentPassword: currentPasswordInput.value,
    newPassword: newPasswordInput.value,
  };

  fetch(GetServerIP() + "/api/password", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(settings),
  })
    .then((response) => {
      if (response.ok) {
        showInfoNotification("Password changed successfully.");
      } else {
        showInfoNotification("Failed to change password.");
      }
      currentPasswordInput.value = "";
      newPasswordInput.value = "";
      confirmPasswordInput.value = "";
    })
    .catch((error) => console.error("Error:", error));
}
