import { toast } from "sonner";

let lastToastMessage: string | null = null;

const getApiBaseUrl = () => {
  return document.location.origin;
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

export async function PostRequest(url: string, bodyData: unknown) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
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
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function GetRequest(url: string) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      credentials: "include",
    });

    if (!res.ok) {
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
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function DeleteRequest(url: string) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
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
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}
