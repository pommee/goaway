import { toast } from "sonner";

let lastToastMessage: string | null = null;

declare global {
  interface Window {
    SERVER_CONFIG?: {
      port: number;
    };
  }
}

export const getApiBaseUrl = () => {
  const serverIP = document.location.origin;

  if (typeof window !== "undefined" && window.SERVER_CONFIG) {
    return serverIP.replace(/:\d+$/, ":" + window.SERVER_CONFIG.port);
  }

  return "http://localhost:8080";
};

const showToast = (message: string, id: string) => {
  if (lastToastMessage !== message) {
    toast.warning("Warning", { id: id, description: message });
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
  ignoreAuth = false,
  ignoreError?: boolean
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
      if (ignoreError !== true) {
        showToast(data.error || "Unknown error occurred", "api-error");
      }
      return [res.status, data];
    }

    const data = await tryParseJson();
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.", "server-unreachable");
    return [500, null];
  }
}

export async function GetRequest(url: string, ignoreError?: boolean): Promise<[number, unknown]> {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      credentials: "include"
    });

    if (!res.ok) {
      await isAuthenticated(res);
      const data = await res.json();
      if (ignoreError !== true) {
        showToast(data.error || "Unknown error occurred", "api-error");
      }
      return [res.status, data.error || "Unknown error occurred"];
    }

    let data: unknown;
    try {
      data = await res.json();
    } catch {
      return [res.status, null];
    }
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.", "server-unreachable");
    return [500, null];
  }
}

export async function PatchRequest(url: string, ignoreError?: boolean): Promise<[number, unknown]> {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      method: "PATCH",
      credentials: "include"
    });

    if (!res.ok) {
      await isAuthenticated(res);
      const data = await res.json();
      if (ignoreError !== true) {
        showToast(data.error || "Unknown error occurred", "api-error");
      }
      return [res.status, data.error || "Unknown error occurred"];
    }

    let data: unknown;
    try {
      data = await res.json();
    } catch {
      return [res.status, null];
    }
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.", "server-unreachable");
    return [500, null];
  }
}

export async function PutRequest(
  url: string,
  bodyData: unknown,
  ignoreError?: boolean
): Promise<[number, unknown]> {
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
      if (ignoreError !== true) {
        showToast(data.error, "api-error");
      }
      return [res.status, data.error];
    }

    let data: unknown;
    try {
      data = await res.json();
    } catch {
      return [res.status, null];
    }
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.", "server-unreachable");
    return [500, null];
  }
}

export async function DeleteRequest(
  url: string,
  body: unknown,
  ignoreError?: boolean
): Promise<[number, unknown]> {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      method: "DELETE",
      credentials: "include",
      body: JSON.stringify(body)
    });

    if (!res.ok) {
      await isAuthenticated(res);
      const data = await res.json();
      if (ignoreError !== true) {
        showToast(data.error, "api-error");
      }
      return [res.status, null];
    }

    let data: unknown;
    try {
      data = await res.json();
    } catch {
      return [res.status, null];
    }
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.", "server-unreachable");
    return [500, null];
  }
}
