async function fetchReleases() {
  const repoUrl = "https://api.github.com/repos/pommee/goaway/releases";
  const cacheTime = 5 * 60 * 1000;

  const lastFetched = localStorage.getItem("releasesLastFetched");
  const cachedReleases = JSON.parse(
    localStorage.getItem("lastFetchedReleases")
  );
  const now = new Date().getTime();

  if (lastFetched && now - lastFetched < cacheTime) {
    displayReleases(cachedReleases);
    return;
  }

  try {
    const response = await fetch(repoUrl);
    if (!response.ok)
      throw new Error(`Failed to fetch releases: ${response.statusText}`);

    const releases = await response.json();
    displayReleases(releases);

    localStorage.setItem("lastFetchedReleases", JSON.stringify(releases));
    localStorage.setItem("releasesLastFetched", now.toString());
  } catch (error) {
    console.error(error);
    document.getElementById("changelog").innerHTML =
      "<p>Error loading releases.</p>";
  }
}

function parseChangelogBody(body) {
  if (!body) return [];

  const sections = [];
  const sectionRegex = /####\s*(.*?)\s*\n([\s\S]*?)(?=\n####|\n$)/g;
  let match;

  while ((match = sectionRegex.exec(body)) !== null) {
    const header = match[1];
    const commits = match[2]
      .trim()
      .split("\n")
      .map((commit) => commit.trim())
      .filter((commit) => commit.length > 0)
      .map((commit) => {
        const hashMatch = commit.match(/\(([a-f0-9]{7,40})\)$/);

        if (hashMatch) {
          const hash = hashMatch[1];
          const message = commit
            .substring(0, commit.lastIndexOf(`(${hash})`))
            .trim();

          const commitUrl = `https://github.com/pommee/goaway/commit/${hash}`;
          return `<a href="${commitUrl}" target="_blank">[${hash}]</a> ${message.replace(
            /^\*\s*/,
            ""
          )}`;
        }

        return commit;
      });

    sections.push({ header, commits });
  }

  return sections;
}

function displayReleases(releases) {
  const changelogEl = document.getElementById("changelog");
  changelogEl.innerHTML = "";

  releases.forEach((release, idx) => {
    const date = new Date(release.published_at);
    const sections = parseChangelogBody(
      release.body || "No release notes available."
    );

    const releaseEl = document.createElement("div");
    releaseEl.className = "changelog-entry";

    const headerEl = document.createElement("div");
    headerEl.className = "release-header";

    const titleEl = document.createElement("h3");
    titleEl.textContent = release.name;
    headerEl.appendChild(titleEl);

    if (idx === 0) {
      const latestTag = document.createElement("div");
      latestTag.textContent = "latest";
      latestTag.className = "latest-tag";
      headerEl.appendChild(latestTag);
    }

    if (release.name.replace("v", "") === GetInstalledVersion()) {
      const installedTag = document.createElement("div");
      installedTag.textContent = "installed";
      installedTag.className = "installed-tag";
      headerEl.appendChild(installedTag);
    }

    releaseEl.appendChild(headerEl);

    const dateString = `${date.toLocaleDateString()} ${date.toLocaleTimeString(
      [],
      { hour: "2-digit", minute: "2-digit", hour12: false }
    )}`;
    releaseEl.innerHTML += `
      <p><strong>Release Date:</strong> ${dateString}</p>
      <hr class="release-date-separator">
    `;

    sections.forEach((section) => {
      const sectionEl = document.createElement("div");
      sectionEl.innerHTML = `<h4>${section.header}</h4><ul>`;

      section.commits.forEach((commit) => {
        sectionEl.innerHTML += `<li>${commit}</li>`;
      });

      sectionEl.innerHTML += "</ul>";
      releaseEl.appendChild(sectionEl);
    });

    releaseEl.innerHTML += `<a href="${release.html_url}" class="view-release-on-github" target="_blank"><i class="fa-brands fa-github"></i> View on GitHub</a>`;

    changelogEl.appendChild(releaseEl);
  });
}

window.addEventListener("load", fetchReleases);
