const inputs = document.querySelectorAll("#main input, #main select");
const savePopup = document.getElementById("save-popup");
const saveButton = document.getElementById("save-btn");
const dismissButton = document.getElementById("dismiss-btn");

document.addEventListener("DOMContentLoaded", () => {
  initializeSettings();
  let isModified = false;

  inputs.forEach((input) => {
    input.addEventListener("input", () => {
      isModified = true;
      showPopup();
    });
  });

  function showPopup() {
    if (isModified) {
      savePopup.style.display = "block";
    }
  }

  function hidePopup() {
    savePopup.style.display = "none";
  }

  saveButton.addEventListener("click", () => {
    saveSettings();
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

  const cacheTTLInSeconds = settings.CacheTTL / 1000000000;
  document.getElementById("cacheTTL").value = cacheTTLInSeconds;
  document.getElementById("logLevel").selectedIndex = settings.LogLevel;
  document.getElementById("disableLogging").checked = settings.LoggingDisabled;
}

async function getSettings() {
  try {
    const response = await fetch(GetServerIP() + "/settings", {
      method: "GET",
    });

    if (!response.ok) {
      console.error("Failed to fetch settings.");
      return null;
    }

    const data = await response.json();
    return data.settings;
  } catch (error) {
    console.error("Error fetching settings:", error);
    throw error;
  }
}

function saveSettings() {
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

  fetch(GetServerIP() + "/settings", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(settings),
  })
    .then((response) => {
      if (response.ok) {
        console.log("Settings updated successfully.");
      } else {
        console.error("Failed to update settings.");
      }
    })
    .catch((error) => console.error("Error:", error));
}
