import { GithubLogo } from "@phosphor-icons/react";
import { useEffect, useState } from "react";

const Changelog = () => {
  const [releases, setReleases] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const cachedData = sessionStorage.getItem("githubReleases");
    const cachedTime = sessionStorage.getItem("githubReleasesTimestamp");

    const now = Date.now();
    const cacheExpiry = cachedTime ? parseInt(cachedTime, 10) : 0;

    if (cachedData && now < cacheExpiry) {
      setReleases(JSON.parse(cachedData));
      setLoading(false);
    } else {
      fetchReleases();
    }
  }, []);

  const fetchReleases = async () => {
    const repoUrl = "https://api.github.com/repos/pommee/goaway/releases";

    try {
      const response = await fetch(repoUrl);
      if (!response.ok)
        throw new Error(`Failed to fetch releases: ${response.statusText}`);

      const data = await response.json();
      const cacheControl = response.headers.get("Cache-Control");
      const cacheMaxAgeMatch = cacheControl?.match(/max-age=(\d+)/);
      const cacheMaxAge = cacheMaxAgeMatch
        ? parseInt(cacheMaxAgeMatch[1], 10) * 1000
        : 300000;

      sessionStorage.setItem("githubReleases", JSON.stringify(data));
      sessionStorage.setItem(
        "githubReleasesTimestamp",
        (Date.now() + cacheMaxAge).toString()
      );

      setReleases(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const parseChangelogBody = (body) => {
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

            return {
              hash,
              message: message.replace(/^\*\s*/, ""),
              url: `https://github.com/pommee/goaway/commit/${hash}`
            };
          }

          return { message: commit, hash: null, url: null };
        });

      sections.push({ header, commits });
    }

    return sections;
  };

  if (loading) return <div className="p-4 text-gray-400">Loading...</div>;
  if (error) return <div className="p-4 text-red-500">Error: {error}</div>;

  return (
    <div className="text-gray-100 p-4 space-y-8 font-mono">
      {releases.map((release, idx) => {
        const date = new Date(release.published_at);
        const sections = parseChangelogBody(
          release.body || "No release notes available."
        );
        const installedVersion = localStorage.getItem("installedVersion");

        return (
          <div
            key={release.id}
            className="changelog-entry bg-gray-800 rounded-md p-4"
          >
            <div className="flex items-center gap-2 mb-2">
              <h3 className="text-xl font-bold">{release.name}</h3>
              {idx === 0 && (
                <span className="latest-tag px-2 py-1 text-xs bg-green-800 text-green-100 rounded">
                  Latest
                </span>
              )}
              {release.name.replace("v", "") === installedVersion && (
                <span className="installed-tag px-2 py-1 text-xs bg-blue-800 text-blue-100 rounded ml-1">
                  installed
                </span>
              )}
            </div>

            <p className="text-gray-400 mb-2">
              Release Date: {date.toLocaleDateString()}{" "}
              {date.toLocaleTimeString([], {
                hour: "2-digit",
                minute: "2-digit",
                hour12: false
              })}
            </p>

            <hr className="border-gray-700 my-2" />

            {sections.map((section, sectionIdx) => (
              <div key={sectionIdx} className="mt-4">
                <h4 className="font-semibold mb-2">{section.header}</h4>
                <ul className="space-y-2">
                  {section.commits.map((commit, commitIdx) => (
                    <li key={commitIdx} className="flex items-start">
                      <span className="mr-2">•</span>
                      {commit.hash ? (
                        <>
                          <a
                            href={commit.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-400 hover:underline mr-2"
                          >
                            [{commit.hash}]
                          </a>
                          <span>{commit.message}</span>
                        </>
                      ) : (
                        <span>{commit.message}</span>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            ))}

            <div className="mt-4 bg-gray-900 font-bold w-fit p-1 rounded-sm">
              <a
                href={release.html_url}
                target="_blank"
                rel="noopener noreferrer"
                className="view-release-on-github inline-flex items-center text-blue-400 hover:underline gap-2"
              >
                <GithubLogo size={16} />
                View on GitHub
              </a>
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default Changelog;
