import { GithubLogoIcon, WarningCircleIcon } from "@phosphor-icons/react";
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
      <div className="min-h-96 flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="flex gap-1">
            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce"></div>
            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce delay-100"></div>
            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce delay-200"></div>
          </div>
          <p className="text-slate-400 text-sm font-medium">
            Loading changelog...
          </p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6 text-red-400 bg-red-900/50 bg-opacity-20 rounded-md border border-red-700 flex items-center justify-center">
        <div className="flex flex-col items-center">
          <WarningCircleIcon size={48} />
          <div className="text-lg font-semibold">Failed to load changelog</div>
          <div className="text-sm mt-1">{error}</div>
          <button
            onClick={fetchReleases}
            className="mt-4 px-4 py-2 bg-red-900 hover:bg-red-700 text-white rounded-md transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  const installedVersion = localStorage.getItem("installedVersion");

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Changelog</h1>
        <p className="text-muted-foreground">
          Stay up to date with the latest changes and improvements.
        </p>
      </div>

      {releases.length === 0 ? (
        <div className="bg-accent border rounded-xl p-8 text-center">
          <p>No release information available.</p>
        </div>
      ) : (
        <div className="space-y-6">
          {releases.map((release, idx) => {
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
                className={`group relative border rounded-xl p-6 shadow-sm hover:shadow-md transition-all duration-200 ${
                  isLatest ? "border-primary bg-accent" : "bg-accent"
                }`}
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <h2 className="text-xl font-semibold">{release.name}</h2>
                    <div className="flex gap-2">
                      {isLatest && (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-600 text-white">
                          Latest
                        </span>
                      )}
                      {isInstalled && (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-600 text-white">
                          Installed
                        </span>
                      )}
                    </div>
                  </div>
                  <time className="text-sm text-muted-foreground font-medium">
                    {date.toLocaleDateString(undefined, {
                      year: "numeric",
                      month: "short",
                      day: "numeric"
                    })}
                  </time>
                </div>

                <div className="space-y-6">
                  {sections.length > 0 ? (
                    sections.map((section, sectionIdx) => (
                      <div key={sectionIdx}>
                        <h3 className="text-sm font-semibold tracking-wide mb-3 flex items-center">
                          <span className="w-2 h-2 bg-orange-400 rounded-full mr-2"></span>
                          {section.header}
                        </h3>
                        <div className="space-y-2">
                          {section.commits.map((commit, commitIdx) => (
                            <div
                              key={commitIdx}
                              className="flex items-start gap-3 group/commit"
                            >
                              <div className="ml-1 mt-2"></div>
                              <div className="flex-1 min-w-0">
                                <div className="flex items-start gap-2">
                                  {commit.hash && (
                                    <a
                                      href={commit.url}
                                      target="_blank"
                                      rel="noopener noreferrer"
                                      className="py-0.5 text-muted-foreground hover:text-primary transition-colors"
                                    >
                                      {`[${commit.hash.substring(0, 7)}]`}
                                    </a>
                                  )}
                                  <p>{commit.message}</p>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="flex items-center justify-center py-8 text-muted-foreground text-sm">
                      <span className="italic">
                        No detailed release notes available
                      </span>
                    </div>
                  )}
                </div>

                <div className="mt-4 pt-2">
                  <div className="flex justify-end">
                    <a
                      href={release.html_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-2 px-3 py-1.5 text-sm border-1 border-stone-600 hover:text-white hover:border-stone-400 rounded-sm transition-colors"
                    >
                      <GithubLogoIcon size={16} />
                      View on GitHub
                    </a>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default Changelog;
