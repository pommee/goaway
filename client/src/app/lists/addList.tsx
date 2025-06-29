import {
  PlusIcon,
  SpinnerGapIcon,
  ListIcon,
  LinkIcon,
  InfoIcon,
  CaretDownIcon
} from "@phosphor-icons/react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DialogDescription } from "@radix-ui/react-dialog";

import { ListEntry } from "@/pages/blacklist";
import { GetRequest } from "@/util";

const RECOMMENDED_SOURCES = [
  {
    name: "StevenBlack's hosts",
    website: "https://github.com/StevenBlack/hosts",
    description: "Unified hosts file with base extensions",
    lists: [
      {
        name: "Steven Black's ad-hoc list",
        url: "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
        description: "Includes adware and malware protection (default list)"
      },
      {
        name: "Gambling",
        url: "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/gambling-only/hosts",
        description: "Includes gambling domains"
      },
      {
        name: "Unified hosts + fakenews + gambling + porn",
        url: "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-gambling-porn/hosts",
        description: "Blocks social media, gambling, adult content and more"
      }
    ]
  },
  {
    name: "BlockListProject",
    website: "https://github.com/blocklistproject/Lists",
    description: "Security-focused blocklists",
    lists: [
      {
        name: "Abuse",
        url: "https://blocklistproject.github.io/Lists/abuse.txt",
        description: "Lists of sites created to deceive"
      },
      {
        name: "Fraud",
        url: "https://blocklistproject.github.io/Lists/fraud.txt",
        description: "Sites created to defraud"
      },
      {
        name: "Malware",
        url: "https://blocklistproject.github.io/Lists/malware.txt",
        description: "Known sites that host malware"
      },
      {
        name: "Tracking",
        url: "https://blocklistproject.github.io/Lists/tracking.txt",
        description: "Sites dedicated to tracking and gathering visitor info"
      }
    ]
  },
  {
    name: "DNS-Blocklists",
    website: "https://github.com/hagezi/dns-blocklists",
    description: "Lists ranging in blocking size",
    lists: [
      {
        name: "Light",
        url: "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/domains/light.txt",
        description: "Light domain list"
      },
      {
        name: "Normal",
        url: "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/domains/multi.txt",
        description: "Normal domain list"
      },
      {
        name: "Pro",
        url: "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/domains/pro.txt",
        description: "Pro domain list"
      },
      {
        name: "Pro Plus",
        url: "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/domains/pro.plus.txt",
        description: "Pro Plus domain list"
      },
      {
        name: "Ultimate",
        url: "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/domains/ultimate.txt",
        description: "Ultimate domain list"
      }
    ]
  }
];

export function AddList({
  onListAdded
}: {
  onListAdded: (list: ListEntry) => void;
}) {
  const [listName, setListName] = useState("");
  const [url, setUrl] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [showSources, setShowSources] = useState(false);
  const [expandedSources, setExpandedSources] = useState<Set<string>>(
    new Set()
  );

  const toggleSourceExpansion = (sourceName: string) => {
    const newExpanded = new Set(expandedSources);
    if (newExpanded.has(sourceName)) {
      newExpanded.delete(sourceName);
    } else {
      newExpanded.add(sourceName);
    }
    setExpandedSources(newExpanded);
  };

  const handleSave = async () => {
    if (!listName.trim() || !url.trim()) {
      toast.error("Please fill in both list name and URL");
      return;
    }

    try {
      new URL(url);
    } catch {
      toast.error("Please enter a valid URL");
      return;
    }

    setIsSaving(true);

    try {
      const [code, response] = await GetRequest(
        `addList?name=${encodeURIComponent(
          listName.trim()
        )}&url=${encodeURIComponent(url.trim())}`
      );

      if (code === 200) {
        const newList: ListEntry = {
          name: listName.trim(),
          url: url.trim(),
          active: response.list.active,
          blockedCount: response.list.blockedCount,
          lastUpdated: response.list.lastUpdated
        };

        onListAdded(newList);
        toast.success(`${listName} has been added successfully!`);
        setModalOpen(false);
        setListName("");
        setUrl("");
      } else {
        toast.error("Failed to add list. Please try again.");
      }
    } catch (error) {
      toast.error("An error occurred while adding the list", {
        description: `${error}`
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    setModalOpen(false);
    setListName("");
    setUrl("");
  };

  const isFormValid = listName.trim() && url.trim();

  return (
    <div className="mb-5">
      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogTrigger asChild>
          <Button className="shadow-sm hover:shadow-md transition-all duration-200">
            <PlusIcon className="mr-2" size={18} />
            Add List
          </Button>
        </DialogTrigger>

        <DialogContent className="sm:max-w-250 max-h-[90vh] overflow-y-auto">
          <DialogHeader className="space-y-3 pb-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-full">
                <ListIcon size={24} className="text-primary" />
              </div>
              <div>
                <DialogTitle className="text-xl font-semibold">
                  Add New Blocklist
                </DialogTitle>
                <DialogDescription className="text-sm text-muted-foreground mt-1">
                  Import a predefined blocklist from a URL
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>

          <div className="space-y-6">
            <div className="space-y-4 p-4 bg-accent rounded-lg border">
              <div className="space-y-2">
                <Label
                  htmlFor="name"
                  className="text-sm font-medium text-muted-foreground flex items-center gap-2"
                >
                  <ListIcon size={16} />
                  List Name
                </Label>
                <Input
                  id="name"
                  value={listName}
                  placeholder="Enter a descriptive name for this list"
                  onChange={(e) => setListName(e.target.value)}
                  className="border-2"
                  disabled={isSaving}
                />
              </div>

              <div className="space-y-2">
                <Label
                  htmlFor="url"
                  className="text-sm font-medium text-muted-foreground flex items-center gap-2"
                >
                  <LinkIcon size={16} />
                  List URL
                </Label>
                <Input
                  id="url"
                  value={url}
                  placeholder="https://example.com/blocklist.txt"
                  onChange={(e) => setUrl(e.target.value)}
                  className="border-2 font-mono text-sm"
                  disabled={isSaving}
                />
                <p className="text-xs text-muted-foreground">
                  Enter the direct URL to a hosts file or domain list
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <InfoIcon size={16} className="text-primary" />
                <span className="text-sm font-medium">
                  Popular Blocklist Sources
                </span>
              </div>

              <div className="border rounded-lg overflow-hidden">
                <button
                  onClick={() => setShowSources(!showSources)}
                  className="w-full p-3 hover:bg-accent transition-colors flex items-center justify-between text-left border-b cursor-pointer"
                  disabled={isSaving}
                >
                  <span className="text-sm font-medium">
                    Browse Recommended Lists ({RECOMMENDED_SOURCES.length})
                  </span>
                  <CaretDownIcon
                    size={16}
                    className={`text-gray-500 transition-transform duration-300 ${
                      showSources ? "rotate-0" : "-rotate-90"
                    }`}
                  />
                </button>

                <div
                  className={`transition-all duration-300 ease-in-out overflow-hidden ${
                    showSources
                      ? "max-h-[1000px] opacity-100"
                      : "max-h-0 opacity-0"
                  }`}
                >
                  <div className="divide-y">
                    {RECOMMENDED_SOURCES.map((source) => (
                      <div
                        key={source.name}
                        className="border-b last:border-b-0"
                      >
                        <button
                          onClick={() => toggleSourceExpansion(source.name)}
                          className="w-full p-3 hover:bg-accent transition-colors flex items-center justify-between text-left cursor-pointer"
                          disabled={isSaving}
                        >
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <h4 className="font-medium">{source.name}</h4>
                              <a
                                href={source.website}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-muted-foreground hover:text-blue-400 transition-colors"
                                onClick={(e) => e.stopPropagation()}
                              >
                                <LinkIcon size={14} />
                              </a>
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                              {source.description}
                            </p>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-xs text-muted-foreground bg-accent px-2 py-1 rounded-full">
                              {source.lists.length} lists
                            </span>
                            <CaretDownIcon
                              size={16}
                              className={`text-muted-foreground transition-transform duration-300 ${
                                expandedSources.has(source.name)
                                  ? "rotate-0"
                                  : "-rotate-90"
                              }`}
                            />
                          </div>
                        </button>

                        <div
                          className={`transition-all duration-300 ease-in-out overflow-hidden ${
                            expandedSources.has(source.name)
                              ? "max-h-[800px] opacity-100"
                              : "max-h-0 opacity-0"
                          }`}
                        >
                          <div className="p-1 transform transition-transform duration-300">
                            {source.lists.map((list, index) => (
                              <div
                                key={list.url}
                                className={`p-3 bg-accent mx-3 mb-3 rounded border-l-4 border-primary transform transition-all duration-300 ${
                                  expandedSources.has(source.name)
                                    ? "translate-y-0 opacity-100"
                                    : "translate-y-[-10px] opacity-0"
                                }`}
                                style={{
                                  transitionDelay: expandedSources.has(
                                    source.name
                                  )
                                    ? `${index * 50}ms`
                                    : "0ms"
                                }}
                              >
                                <div className="flex items-start justify-between gap-3">
                                  <div className="flex-1 min-w-0">
                                    <h5 className="font-medium text-sm">
                                      {list.name}
                                    </h5>
                                    <p className="text-xs text-muted-foreground mt-1">
                                      {list.description}
                                    </p>
                                    <a
                                      href={list.url}
                                      target="_blank"
                                      className="text-xs text-blue-500 mt-1 font-mono truncate hover:underline transition-colors"
                                    >
                                      {list.url}
                                    </a>
                                  </div>
                                  <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => {
                                      setListName(list.name);
                                      setUrl(list.url);
                                    }}
                                    disabled={isSaving}
                                    className="text-xs hover:bg-primary hover:text-primary-foreground transition-colors"
                                  >
                                    Use This
                                  </Button>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
              <div className="text-xs text-muted-foreground ml-1">
                <strong>Tip:</strong> Expand the list above to see popular
                blocklist sources, or manually enter your own list details in
                the form.
              </div>
            </div>
          </div>

          <DialogFooter className="flex flex-col sm:flex-row gap-4">
            <Button
              variant="outline"
              onClick={handleCancel}
              disabled={isSaving}
              className="order-2 sm:order-1 transition-colors"
            >
              Cancel
            </Button>
            <Button
              variant="default"
              onClick={handleSave}
              disabled={isSaving || !isFormValid}
              className="order-1 sm:order-2 shadow-sm hover:shadow-md transition-all duration-200"
            >
              {isSaving ? (
                <>
                  <SpinnerGapIcon className="animate-spin mr-2 h-4 w-4" />
                  Adding List...
                </>
              ) : (
                <>
                  <PlusIcon className="mr-2 h-4 w-4" />
                  Add List
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
