import { GithubLogo, WarningCircleIcon } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

const Changelog = () => {
  const [releases, setReleases] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error] = useState(null);

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
    } catch {
      toast.warning("Could not fetch changelog");
    } finally {
      setLoading(false);
    }
  };

  const parseChangelogBody = (body: string) => {
    if (!body) return [];

    const sections = [];
    const sectionRegex = /###\s*(.*?)\s*\n([\s\S]*?)(?=\n###|\n##|$)/g;
    let match;

    while ((match = sectionRegex.exec(body)) !== null) {
      const header = match[1];
      const content = match[2].trim();

      if (!content) continue;

      const commits = content
        .split("\n")
        .map((line) => line.trim())
        .filter((line) => line.length > 0 && line.startsWith("*"))
        .map((commit) => {
          const linkMatch = commit.match(
            /\*\s*(.*?)\s*\(\[([a-f0-9]{7,40})\]\((.*?)\)\)/
          );

          if (linkMatch) {
            const message = linkMatch[1].trim();
            const hash = linkMatch[2];
            const url = linkMatch[3];

            return {
              hash,
              message,
              url
            };
          }

          const hashMatch = commit.match(/\*\s*(.*?)\s*\(([a-f0-9]{7,40})\)$/);
          if (hashMatch) {
            const message = hashMatch[1].trim();
            const hash = hashMatch[2];

            return {
              hash,
              message,
              url: `https://github.com/pommee/goaway/commit/${hash}`
            };
          }

          return {
            message: commit.replace(/^\*\s*/, "").trim(),
            hash: null,
            url: null
          };
        });

      if (commits.length > 0) {
        sections.push({ header, commits });
      }
    }

    return sections;
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center py-16">
        <div className="animate-pulse text-gray-400 flex items-center gap-2">
          <div className="w-4 h-4 rounded-full bg-gray-500 animate-bounce"></div>
          <div className="w-4 h-4 rounded-full bg-gray-500 animate-bounce delay-100"></div>
          <div className="w-4 h-4 rounded-full bg-gray-500 animate-bounce delay-200"></div>
          <span className="ml-2">Loading changelog...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6 text-red-500 bg-red-900 bg-opacity-20 rounded-md border border-red-700 flex items-center justify-center">
        <div className="flex flex-col items-center">
          <WarningCircleIcon size={48} />
          <div className="text-lg font-semibold">Failed to load changelog</div>
          <div className="text-sm mt-1">{error}</div>
          <button
            onClick={fetchReleases}
            className="mt-4 px-4 py-2 bg-red-800 hover:bg-red-700 text-white rounded-md transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  const installedVersion = localStorage.getItem("installedVersion");

  return (
    <div className="text-gray-100 p-4 space-y-6 font-mono max-w-4xl mx-auto">
      {releases.length === 0 ? (
        <div className="p-6 bg-gray-800 rounded-md text-center">
          No release information available.
        </div>
      ) : (
        releases.map((release, idx) => {
          const date = new Date(release.published_at);
          const sections = parseChangelogBody(
            release.body || "No release notes available."
          );
          const isLatest = idx === 0;
          const isInstalled =
            release.name.replace("v", "") === installedVersion;

          return (
            <div
              key={release.id}
              className={`changelog-entry border border-gray-700 rounded-lg p-5 bg-gray-800 transition-all ${
                isLatest ? "shadow-md shadow-green-900/20" : ""
              } ${isInstalled ? "border-l-4 border-l-green-600" : ""}`}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <h3 className="text-xl font-bold">{release.name}</h3>
                  <div className="flex gap-1">
                    {isLatest && (
                      <span className="latest-tag px-2 py-0.5 text-xs bg-green-800 text-green-100 rounded font-semibold">
                        Latest
                      </span>
                    )}
                    {isInstalled && (
                      <span className="installed-tag px-2 py-0.5 text-xs bg-blue-800 text-blue-100 rounded font-semibold">
                        Installed
                      </span>
                    )}
                  </div>
                </div>
                <p className="text-gray-400 text-sm">
                  {date.toLocaleDateString(undefined, {
                    year: "numeric",
                    month: "short",
                    day: "numeric"
                  })}
                </p>
              </div>

              <hr className="border-gray-700 my-3" />

              {sections.length > 0 ? (
                sections.map((section, sectionIdx) => (
                  <div key={sectionIdx} className="mt-4">
                    <h4 className="font-semibold mb-2 text-gray-200 bg-gray-700 px-2 py-1 rounded inline-block">
                      {section.header}
                    </h4>
                    <ul className="space-y-2 mt-2">
                      {section.commits.map((commit, commitIdx) => (
                        <li key={commitIdx} className="flex items-start group">
                          <span className="mr-2 text-green-400">•</span>
                          <div>
                            {commit.hash ? (
                              <div className="flex items-baseline">
                                <a
                                  href={commit.url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="text-blue-400 hover:text-blue-300 hover:underline text-sm font-mono mr-2 transition-colors"
                                >
                                  [{commit.hash.substring(0, 7)}]
                                </a>
                                <span className="text-gray-200">
                                  {commit.message}
                                </span>
                              </div>
                            ) : (
                              <span className="text-gray-200">
                                {commit.message}
                              </span>
                            )}
                          </div>
                        </li>
                      ))}
                    </ul>
                  </div>
                ))
              ) : (
                <div className="py-2 text-gray-400 italic">
                  No detailed release notes available.
                </div>
              )}

              <div className="mt-4 flex justify-end">
                <a
                  href={release.html_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-1 px-3 py-1 bg-gray-900 hover:bg-gray-700 text-blue-400 rounded-md transition-colors text-sm"
                >
                  <GithubLogo size={14} />
                  View on GitHub
                </a>
              </div>
            </div>
          );
        })
      )}
    </div>
  );
};

export default Changelog;
