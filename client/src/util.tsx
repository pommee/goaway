import { toast } from "sonner";

let lastToastMessage: string | null = null;

export const getApiBaseUrl = () => {
  if (typeof window !== "undefined" && window.SERVER_CONFIG) {
    return window.SERVER_CONFIG.apiBaseURL;
  }

  return "http://localhost:8080";
};

const showToast = (message: string) => {
  if (lastToastMessage !== message) {
    toast.warning("Warning", { description: message });
    lastToastMessage = message;

    setTimeout(() => {
      lastToastMessage = null;
    }, 5000);
  }
};

async function isAuthenticated(res: Response) {
  if (res.status === 401) {
    document.location.href = "/login";
  }
}

export async function PostRequest(
  url: string,
  bodyData: unknown,
  ignoreAuth = false
) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(bodyData),
      credentials: "include"
    });

    const tryParseJson = async () => {
      const text = await res.text();
      return text ? JSON.parse(text) : null;
    };

    if (!res.ok) {
      if (!ignoreAuth) {
        await isAuthenticated(res);
      }
      const data = await tryParseJson();
      showToast(data?.error || "Unknown error occurred.");
      return [res.status, null];
    }

    const data = await tryParseJson();
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function GetRequest(url: string) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      credentials: "include"
    });

    if (!res.ok) {
      await isAuthenticated(res);
      const data = await res.json();
      showToast(data.error);
      return [res.status, null];
    }

    const data = await res.json();
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function PutRequest(url: string, bodyData: unknown) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(bodyData),
      credentials: "include"
    });

    if (!res.ok) {
      await isAuthenticated(res);
      const data = await res.json();
      showToast(data.error);
      return [res.status, data.error];
    }

    let data;
    try {
      data = await res.json();
    } catch {
      return [res.status, null];
    }
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function DeleteRequest(url: string, body) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      method: "DELETE",
      credentials: "include",
      body: JSON.stringify(body)
    });

    if (!res.ok) {
      await isAuthenticated(res);
      const data = await res.json();
      showToast(data.error);
      return [res.status, null];
    }

    let data;
    try {
      data = await res.json();
    } catch {
      return [res.status, null];
    }
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}
