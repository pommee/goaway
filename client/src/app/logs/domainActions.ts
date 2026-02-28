import { DeleteRequest, PostRequest } from "@/util";
import { toast } from "sonner";

export async function BlacklistDomain(domain: string) {
  try {
    await DeleteRequest(`whitelist?domain=${domain}`, null, true);

    const [status] = await PostRequest(
      "custom",
      { domains: [domain] },
      false,
      false
    );
    if (status === 200) {
      toast.success(`Blacklisted ${domain}`);
    } else {
      toast.error(`Failed to block ${domain}`);
    }
  } catch {
    toast.error("An error occurred while sending the request.");
  }
}

export async function WhitelistDomain(domain: string) {
  await DeleteRequest(`blacklist?domain=${domain}`, null, true);

  const [code, response] = await PostRequest(
    "whitelist",
    {
      domain: domain
    },
    false,
    true
  );
  if (code === 200) {
    toast.success(`Whitelisted ${domain}`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}
