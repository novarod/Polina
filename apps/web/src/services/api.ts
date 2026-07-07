export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

interface ApiFetchOptions {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  redirectOn401?: boolean;
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const data: unknown = await response.json();
    if (
      typeof data === "object" &&
      data !== null &&
      "message" in data &&
      typeof data.message === "string"
    ) {
      return data.message;
    }
  } catch {
    return `request failed with status ${response.status}`;
  }
  return `request failed with status ${response.status}`;
}

export async function apiFetch<T>(
  path: string,
  { method = "GET", body, redirectOn401 = true }: ApiFetchOptions = {}
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`/api${path}`, {
      method,
      headers:
        body !== undefined ? { "Content-Type": "application/json" } : undefined,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
    throw new ApiError(0, "Não foi possível conectar ao servidor");
  }

  if (response.status === 401 && redirectOn401) {
    window.location.assign("/login");
    throw new ApiError(401, "Sessão expirada");
  }

  if (!response.ok) {
    throw new ApiError(response.status, await readErrorMessage(response));
  }

  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}
