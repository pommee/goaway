document.querySelectorAll(".toggle input").forEach(function (toggle) {
  toggle.addEventListener("change", function () {
    const statusText = this.checked ? "On" : "Off";
    this.nextElementSibling.textContent = statusText;
  });
});
