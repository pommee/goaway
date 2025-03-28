import { toast } from "sonner";

let lastToastMessage: string | null = null;

export const getApiBaseUrl = () => {
  if (window.SERVER_CONFIG && window.SERVER_CONFIG.apiBaseURL) {
    return window.SERVER_CONFIG.apiBaseURL;
  }

  return import.meta.env.VITE_API_URL || "/api";
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

export async function PostRequest(url: string, bodyData: any) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/${url}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(bodyData),
      credentials: "include",
    });

    if (!res.ok) {
      const data = await res.json();
      showToast(data.error);
      return [res.status, null];
    }

    const data = await res.json();
    return [res.status, data];
  } catch (error) {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function GetRequest(url: string) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/${url}`, {
      credentials: "include",
    });

    if (!res.ok) {
      const data = await res.json();
      showToast(data.error);
      return [res.status, null];
    }

    const data = await res.json();
    return [res.status, data];
  } catch (error) {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function DeleteRequest(url: string) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/${url}`, {
      method: "DELETE",
      credentials: "include",
    });

    if (!res.ok) {
      const data = await res.json();
      showToast(data.error);
      return [res.status, null];
    }

    const data = await res.json();
    return [res.status, data];
  } catch (error) {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}
