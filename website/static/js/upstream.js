const upstreamDNSElement = document.getElementById("upstream-dns");

async function getUpstreamDNS() {
  data = await GetRequest("/settings");
  populateUpstreamDNSArea(data.settings.UpstreamDNS);
}

function populateUpstreamDNSArea(dnsList) {
  console.log(dnsList);
}

document.addEventListener("DOMContentLoaded", () => {
  getUpstreamDNS();
});

function saveDns() {
  const dnsInput = document.getElementById("upstream-dns").value;
  console.log("Saved DNS IPs:", dnsInput);
  alert("DNS IPs saved successfully!");
}
