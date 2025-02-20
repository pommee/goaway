async function fetchReleases() {
  const repoUrl = "https://api.github.com/repos/pommee/goaway/releases";

  const lastFetched = localStorage.getItem("lastFetched");
  const lastFetchedReleases = JSON.parse(
    localStorage.getItem("lastFetchedReleases")
  );
  const now = new Date().getTime();

  if (lastFetched && now - lastFetched < 5 * 60 * 1000) {
    console.log("Using cached releases from last fetch.");
    displayReleases(lastFetchedReleases);
    return;
  }

  try {
    const response = await fetch(repoUrl);
    if (!response.ok) {
      throw new Error(`Failed to fetch releases: ${response.statusText}`);
    }
    const releases = await response.json();
    displayReleases(releases);

    localStorage.setItem("lastFetchedReleases", JSON.stringify(releases));
    localStorage.setItem("lastFetched", now);
  } catch (error) {
    console.error(error);
    document.getElementById("changelog").innerHTML =
      "<p>Error loading releases.</p>";
  }
}

function parseChangelogBody(body) {
  const changelogSections = [];
  const regex = /####\s*(.*?)\s*\n([\s\S]*?)(?=\n####|\n$)/g;
  let match;

  while ((match = regex.exec(body)) !== null) {
    const sectionHeader = match[1];
    const sectionCommits = match[2]
      .trim()
      .split("\n")
      .map((commit) => commit.trim())
      .filter((commit) => commit.length > 0);

    const processedCommits = sectionCommits.map((commit) => {
      return commit.replace(/\(([a-f0-9]{7,40})\)/g, (match, hash) => {
        const commitUrl = `https://github.com/pommee/goaway/commit/${hash}`;
        return `<a href="${commitUrl}" target="_blank">${match}</a>`;
      });
    });

    changelogSections.push({
      header: sectionHeader,
      commits: processedCommits,
    });
  }

  return changelogSections;
}

function displayReleases(releases) {
  const changelogSection = document.getElementById("changelog");
  changelogSection.innerHTML = "";

  releases.forEach((release) => {
    const releaseDate = new Date(release.published_at);
    const changelogBody = release.body
      ? release.body
      : "No release notes available.";
    const parsedChangelog = parseChangelogBody(changelogBody);

    const releaseElement = document.createElement("div");
    releaseElement.classList.add("changelog-entry");

    releaseElement.innerHTML = `
    <h3>${release.name}</h3>
    <p><strong>Release Date:</strong> ${releaseDate.toLocaleDateString()}</p>
    <hr class="release-date-separator">
  `;

    parsedChangelog.forEach((section) => {
      const sectionElement = document.createElement("div");
      sectionElement.innerHTML = `<h4>${section.header}</h4><ul>`;
      section.commits.forEach((commit) => {
        commit = commit.replaceAll("*", "");
        sectionElement.innerHTML += `<li>${commit}</li>`;
      });
      sectionElement.innerHTML += "</ul>";
      releaseElement.appendChild(sectionElement);
    });

    releaseElement.innerHTML += `<a href="${release.html_url}" class="view-release-on-github" target="_blank">View on GitHub</a>`;

    changelogSection.appendChild(releaseElement);
  });
}

window.addEventListener("load", fetchReleases);
