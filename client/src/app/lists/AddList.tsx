import {
  PlusIcon,
  SpinnerGapIcon,
  ListIcon,
  LinkIcon,
  InfoIcon,
  CaretDownIcon,
  PowerIcon,
  ClipboardTextIcon,
  CodeIcon,
  TrashIcon
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
  DialogFooter,
  DialogDescription
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";

import { ListEntry } from "@/pages/blacklist";
import { PostRequest } from "@/util";

interface MultiListEntry {
  id: string;
  name: string;
  url: string;
}

interface RecommendedList {
  name: string;
  url: string;
  description: string;
}

interface RecommendedSource {
  name: string;
  website: string;
  description: string;
  lists: RecommendedList[];
}

const RECOMMENDED_SOURCES: RecommendedSource[] = [
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
  const [isMultipleTab, setIsMultipleTab] = useState(false);
  const [listActive, setListActive] = useState(true);
  const [multiEntries, setMultiEntries] = useState<MultiListEntry[]>([]);

  const isValidUrl = (urlString: string): boolean => {
    try {
      new URL(urlString);
      return true;
    } catch {
      return false;
    }
  };

  const toggleSourceExpansion = (sourceName: string) => {
    setExpandedSources((prev) => {
      const next = new Set(prev);
      if (next.has(sourceName)) {
        next.delete(sourceName);
      } else {
        next.add(sourceName);
      }
      return next;
    });
  };

  const resetForm = () => {
    setListName("");
    setUrl("");
    setMultiEntries([]);
    setModalOpen(false);
  };

  const handleSaveSingle = async () => {
    if (!listName.trim() || !url.trim()) {
      toast.error("Please fill in both list name and URL");
      return;
    }

    if (!isValidUrl(url)) {
      toast.error("Please enter a valid URL");
      return;
    }

    setIsSaving(true);

    try {
      const [code, response] = await PostRequest("addList", {
        name: listName.trim(),
        url: url.trim(),
        active: listActive
      });

      if (code === 200) {
        const newList: ListEntry = {
          name: listName.trim(),
          url: url.trim(),
          active: response.active,
          blockedCount: response.blockedCount,
          lastUpdated: response.lastUpdated
        };

        onListAdded(newList);
        toast.success(`${listName} has been added successfully!`);
        resetForm();
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

  const handleSaveMultiple = async () => {
    const validEntries = multiEntries.filter(
      (e) => e.name.trim() && e.url.trim()
    );

    if (validEntries.length === 0) {
      toast.error("Please fill in at least one list name and URL");
      return;
    }

    for (const entry of validEntries) {
      if (!isValidUrl(entry.url)) {
        toast.error(`Invalid URL: ${entry.url}`);
        return;
      }
    }

    setIsSaving(true);

    try {
      const [code, response] = await PostRequest("addLists", {
        lists: validEntries.map((e) => ({
          name: e.name.trim(),
          url: e.url.trim(),
          active: listActive
        }))
      });

      if (code === 200) {
        const ignoredCount = response.ignored
          ? Object.keys(response.ignored).length
          : 0;

        if (ignoredCount > 0) {
          toast.warning(`${ignoredCount} lists were ignored`, {
            description: "Reason: Lists already exist"
          });
        } else {
          toast.success(`${validEntries.length} lists added successfully!`);
        }

        resetForm();
      } else {
        toast.error("Failed to add lists. Please try again.");
      }
    } catch (error) {
      toast.error("An error occurred while adding lists", {
        description: `${error}`
      });
    } finally {
      setIsSaving(false);
    }
  };

  const addMultiEntry = () => {
    setMultiEntries((prev) => [
      ...prev,
      { id: Date.now().toString(), name: "", url: "" }
    ]);
  };

  const updateMultiEntry = (
    index: number,
    field: "name" | "url",
    value: string
  ) => {
    setMultiEntries((prev) => {
      const updated = [...prev];
      updated[index][field] = value;
      return updated;
    });
  };

  const removeMultiEntry = (index: number) => {
    setMultiEntries((prev) => prev.filter((_, i) => i !== index));
  };

  const parseBulkUrls = () => {
    const textArea = document.getElementById(
      "bulk-urls"
    ) as HTMLTextAreaElement;
    if (!textArea?.value.trim()) return;

    const urls = textArea.value
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter((line) => line);

    const newEntries: MultiListEntry[] = urls.map((url) => ({
      id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      name: "",
      url
    }));

    setMultiEntries((prev) => [...prev, ...newEntries]);
    textArea.value = "";
  };

  const validMultiEntriesCount = multiEntries.filter(
    (e) => e.name.trim() && e.url.trim()
  ).length;

  const isSingleFormValid = listName.trim() && url.trim();
  const isSaveDisabled = isMultipleTab
    ? validMultiEntriesCount === 0
    : !isSingleFormValid;

  return (
    <div className="mb-5">
      <Dialog open={isSaving}>
        <DialogContent className="flex flex-col items-center justify-center gap-4 max-w-sm">
          <div className="flex text-lg font-medium">
            <SpinnerGapIcon
              className="animate-spin text-primary mt-0.5 mr-2"
              size={24}
            />
            Adding list...
          </div>
          <div className="text-sm text-muted-foreground text-center">
            Please wait while your list is being added.
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogTrigger asChild>
          <Button>
            <PlusIcon className="mr-2" size={18} />
            Add List
          </Button>
        </DialogTrigger>

        <DialogContent className="max-w-full lg:max-w-2/3 overflow-y-auto">
          <DialogHeader className="space-y-3 pb-2">
            <div className="flex items-center gap-3">
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

          <div className="flex max-w-[90vw] flex-col gap-6">
            <Tabs
              defaultValue="single"
              onValueChange={(val) => setIsMultipleTab(val === "multiple")}
            >
              <TabsList className="w-full mb-2 bg-transparent gap-5">
                <TabsTrigger
                  value="single"
                  className="cursor-pointer hover:bg-muted-foreground/10"
                >
                  <ClipboardTextIcon />
                  Single
                </TabsTrigger>
                <TabsTrigger
                  value="multiple"
                  className="cursor-pointer hover:bg-muted-foreground/10"
                >
                  <CodeIcon />
                  Multiple
                </TabsTrigger>
              </TabsList>

              <TabsContent value="single">
                <div className="space-y-4">
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

                  <div className="flex gap-2">
                    <Switch
                      id="active"
                      checked={listActive}
                      onCheckedChange={setListActive}
                      disabled={isSaving}
                    />
                    <Label
                      htmlFor="active"
                      className="text-muted-foreground flex items-center"
                    >
                      List active
                    </Label>
                  </div>
                </div>

                <div className="space-y-3 mt-5">
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
                        className={`text-muted-foreground transition-transform duration-300 ${
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
                                        <p className="text-xs text-muted-foreground">
                                          {list.description}
                                        </p>
                                        <a
                                          href={list.url}
                                          target="_blank"
                                          rel="noopener noreferrer"
                                          className="text-xs text-blue-400 font-mono truncate hover:underline transition-colors"
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
                    blocklist sources, or manually enter your own list details
                    in the form.
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="multiple">
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="space-y-1">
                      <Label className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                        <ListIcon size={16} />
                        Blocklist Entries
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        Add multiple blocklists with custom names and URLs
                      </p>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={addMultiEntry}
                      disabled={isSaving}
                      className="shrink-0"
                    >
                      <PlusIcon size={16} className="mr-1" />
                      Add Entry
                    </Button>
                  </div>

                  <ScrollArea className="max-h-94 overflow-auto">
                    <div className="overflow-y-auto">
                      {multiEntries.map((entry, index) => (
                        <Card
                          key={entry.id}
                          className="bg-transparent border-none py-2 hover:bg-muted-foreground/10 rounded-sm"
                        >
                          <CardContent className="pl-0 pr-2">
                            <div className="flex items-start gap-3">
                              <div className="flex gap-2 w-full">
                                <Input
                                  value={entry.name}
                                  onChange={(e) =>
                                    updateMultiEntry(
                                      index,
                                      "name",
                                      e.target.value
                                    )
                                  }
                                  placeholder="Enter list name"
                                  className="h-9"
                                  disabled={isSaving}
                                />
                                <Input
                                  value={entry.url}
                                  onChange={(e) =>
                                    updateMultiEntry(
                                      index,
                                      "url",
                                      e.target.value
                                    )
                                  }
                                  placeholder="https://example.com/blocklist.txt"
                                  className="h-9 font-mono text-sm w-full"
                                  disabled={isSaving}
                                />
                              </div>

                              {multiEntries.length > 1 && (
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => removeMultiEntry(index)}
                                  disabled={isSaving}
                                  className="h-9 w-9 hover:bg-destructive/50"
                                >
                                  <TrashIcon size={16} />
                                </Button>
                              )}
                            </div>
                          </CardContent>
                        </Card>
                      ))}
                    </div>
                  </ScrollArea>

                  <div className="space-y-2">
                    <Label className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                      <ClipboardTextIcon size={16} />
                      Paste URLs
                    </Label>
                    <p className="text-xs text-muted-foreground">
                      Paste multiple URLs (one per line) and click "Parse URLs"
                      to create entries.
                    </p>
                    <textarea
                      id="bulk-urls"
                      rows={3}
                      placeholder="Paste one URL per line&#10;https://example/blockthis.txt&#10;..."
                      className="w-full border-2 rounded-md p-2 font-mono text-sm resize-y"
                      disabled={isSaving}
                    />
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={parseBulkUrls}
                      disabled={isSaving}
                    >
                      Parse URLs
                    </Button>
                  </div>

                  <div className="space-y-3 pt-4 border-t">
                    <div className="space-y-2">
                      <Label
                        htmlFor="multi-active"
                        className="text-sm font-medium text-muted-foreground flex items-center gap-2"
                      >
                        <PowerIcon size={16} />
                        All lists active
                      </Label>
                      <Switch
                        id="multi-active"
                        checked={listActive}
                        onCheckedChange={setListActive}
                        disabled={isSaving}
                      />
                      <p className="text-xs text-muted-foreground">
                        This setting will be applied to all lists when added
                      </p>
                    </div>
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </div>

          <DialogFooter className="flex flex-col sm:flex-row gap-4">
            {isMultipleTab && multiEntries.length > 0 && (
              <div className="py-1 px-10 border-b-2 border-p-2 border-primary">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Ready to add:</span>
                  <span className="font-medium ml-1">
                    {validMultiEntriesCount} of {multiEntries.length} lists
                  </span>
                </div>
              </div>
            )}
            <Button
              variant="outline"
              onClick={resetForm}
              disabled={isSaving}
              className="order-2 sm:order-1 transition-colors"
            >
              Cancel
            </Button>
            <Button
              variant="default"
              onClick={isMultipleTab ? handleSaveMultiple : handleSaveSingle}
              disabled={isSaving || isSaveDisabled}
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
                  {isMultipleTab ? "Add Lists" : "Add List"}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
