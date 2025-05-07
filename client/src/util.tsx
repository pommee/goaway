import { toast } from "sonner";

let lastToastMessage: string | null = null;

const getApiBaseUrl = () => {
  const fullURL = document.location.origin;
  const newBaseUrl = fullURL.replace(/:\d+$/, ":8080");
  return newBaseUrl;
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
  console.log(res.status);
  if (res.status === 401) {
    document.location.href = "/login";
  }
}

export async function PostRequest(url: string, bodyData: unknown) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      method: "POST",
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
      return [res.status, null];
    }

    const data = await res.json();
    return [res.status, data];
  } catch {
    showToast("Could not reach server, try again later.");
    return [500, null];
  }
}

export async function GetRequest(url: string): Promise<[number, any]> {
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

export async function DeleteRequest(url: string) {
  try {
    const res = await fetch(`${getApiBaseUrl()}/api/${url}`, {
      method: "DELETE",
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
