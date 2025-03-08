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
const fontSelection = document.getElementById("fontSelection");

const colorEntries = [
  {
    name: "Primary Background",
    selector: "--bg-primary",
  },
  {
    name: "Secondary Background",
    selector: "--bg-secondary",
  },
  {
    name: "Tertiary background",
    selector: "--bg-tertiary",
  },
  {
    name: "Hover Background",
    selector: "--hover-bg",
  },
  {
    name: "Metric background",
    selector: "--metric-bg",
  },
  {
    name: "Text Primary",
    selector: "--text-primary",
  },
  {
    name: "Text Secondary",
    selector: "--text-secondary",
  },
  {
    name: "Text Muted",
    selector: "--text-muted",
  },
  {
    name: "Primary Accent",
    selector: "--accent-primary",
  },
  {
    name: "Secondary Accent",
    selector: "--accent-secondary",
  },
  {
    name: "Notification Success",
    selector: "--success-color",
  },
  {
    name: "Notification Warning",
    selector: "--warning-color",
  },
  {
    name: "Notification Error",
    selector: "--danger-color",
  },
  {
    name: "Border",
    selector: "--border-color",
  },
];

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
  const savedTheme = localStorage.getItem("theme") || "dark";
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

  const savedFont =
    localStorage.getItem("selectedFont") || "'JetBrains Mono', monospace";
  document.body.style.fontFamily = savedFont;
  fontSelection.value = savedFont;

  fontSelection.addEventListener("change", () => {
    const selectedFont = fontSelection.value;
    document.body.style.fontFamily = selectedFont;
    localStorage.setItem("selectedFont", selectedFont);
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

function openModal() {
  const rootStyles = getComputedStyle(document.documentElement);
  const colorList = document.getElementById("color-list");
  const previewWindow = document.getElementById("color-scheme-preview");
  const currentTheme = localStorage.getItem("theme") || "dark";

  document.getElementById("colorSchemeModal").style.display = "block";
  colorList.innerHTML = "";
  previewWindow.src = "./index.html";

  colorEntries.forEach((entry) => {
    const colorValue = rootStyles.getPropertyValue(entry.selector).trim();
    const label = document.createElement("label");
    const input = document.createElement("input");

    label.innerText = entry.name;
    input.type = "color";

    if (colorValue.includes("light-dark")) {
      input.value =
        currentTheme === "dark"
          ? getColor(colorValue, 1)
          : getColor(colorValue, 0);
    } else {
      input.value = colorValue;
    }

    input.oninput = (e) => {
      document.documentElement.style.setProperty(
        entry.selector,
        e.target.value
      );
    };

    colorList.appendChild(label);
    colorList.appendChild(input);
    colorList.appendChild(document.createElement("br"));
  });
}

function getColor(entry, index) {
  const regex =
    /\((#[0-9a-fA-F]{6})(?:[0-9a-fA-F]{2})?,\s*(#[0-9a-fA-F]{6})(?:[0-9a-fA-F]{2})?\)/;
  const matches = entry.match(regex);
  return matches ? matches[index + 1] : null;
}

function closeColorScheme() {
  document.getElementById("colorSchemeModal").style.display = "none";
}

function saveColorScheme() {
  const root = document.documentElement;
  let colorScheme = {};

  colorEntries.forEach((entry, index) => {
    const colorPicker = document.querySelectorAll("#color-list input")[index];
    if (colorPicker) {
      root.style.setProperty(entry.selector, colorPicker.value);
      colorScheme[entry.selector] = colorPicker.value;
      localStorage.setItem("color-scheme", JSON.stringify(colorScheme));
    }
  });

  document.getElementById("colorSchemeModal").style.display = "none";
  showInfoNotification("Updated color scheme");
}

function resetColorScheme() {
  const root = document.documentElement;

  colorEntries.forEach((entry) => {
    const defaultValue = getComputedStyle(root)
      .getPropertyValue(entry.selector)
      .trim();

    root.style.setProperty(entry.selector, defaultValue);
  });

  localStorage.removeItem("color-scheme");
  closeColorScheme();
  showInfoNotification("Color scheme reset to default.");
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
