type Json = string | number | boolean | null | Json[] | { [key: string]: Json };

type Method = "GET" | "POST" | "PATCH" | "PUT" | "DELETE";

type ApiOption = {
  method?: Method;
  body?: Json;
  auth?: boolean;
  csrf?: boolean;
  headers?: Record<string, string>;
};

type ApiErrBody = {
  error?: string;
  message?: string;
};

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

const tokenKey = "access_token";
const csrfTokenKey = "csrf_token";

function getBaseUrl(): string {
  return import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
}

function joinUrl(base: string, path: string): string {
  const left = base.endsWith("/") ? base.slice(0, -1) : base;
  const right = path.startsWith("/") ? path : `/${path}`;
  return `${left}${right}`;
}

export function getToken(): string {
  return localStorage.getItem(tokenKey) || "";
}
type CsrfRes = {
  token: string;
};

export function getCSRFToken(): string {
  return localStorage.getItem(csrfTokenKey) || "";
}

export function setCSRFToken(token: string): void {
  localStorage.setItem(csrfTokenKey, token);
}

export function clearCSRFToken(): void {
  localStorage.removeItem(csrfTokenKey);
}

export async function ensureCSRFToken(): Promise<string> {
  const current = getCSRFToken();
  if (current) {
    return current;
  }

  const res = await fetch(joinUrl(getBaseUrl(), "/auth/csrf"), {
    method: "GET",
    credentials: "include",
  });

  if (!res.ok) {
    throw new ApiError(
      res.status,
      "csrf_failed",
      "csrf token could not be issued",
    );
  }

  const data = await readJsonSafe<CsrfRes>(res);
  if (!data?.token) {
    throw new ApiError(500, "invalid_response", "csrf response is invalid");
  }

  setCSRFToken(data.token);
  return data.token;
}
export function setToken(token: string): void {
  localStorage.setItem(tokenKey, token);
}

export function clearToken(): void {
  localStorage.removeItem(tokenKey);
}

export function getCookie(name: string): string {
  const items = document.cookie.split(";");

  for (const item of items) {
    const value = item.trim();

    if (value.startsWith(`${name}=`)) {
      return decodeURIComponent(value.slice(name.length + 1));
    }
  }

  return "";
}

function buildHeaders(opt?: ApiOption): Headers {
  const headers = new Headers();

  if (opt?.auth) {
    const token = getToken();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  if (opt?.csrf) {
    const csrf = getCSRFToken() || getCookie("csrf_token");
    if (csrf) {
      headers.set("X-CSRF-Token", csrf);
    }
  }
  if (opt?.headers) {
    for (const [key, value] of Object.entries(opt.headers)) {
      if (value.trim() !== "") {
        headers.set(key, value);
      }
    }
  }

  if (opt?.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  return headers;
}

async function readJsonSafe<T>(res: Response): Promise<T | null> {
  const contentType = res.headers.get("Content-Type") || "";

  if (!contentType.includes("application/json")) {
    return null;
  }

  return (await res.json()) as T;
}

async function throwApiError(res: Response): Promise<never> {
  const body = await readJsonSafe<ApiErrBody>(res);

  const code = body?.error || "request_failed";
  const message = body?.message || body?.error || `HTTP ${res.status}`;

  throw new ApiError(res.status, code, message);
}

type RefreshRes = {
  access_token: string;
};

async function tryRefresh(): Promise<boolean> {
  let csrf = getCSRFToken() || getCookie("csrf_token");

  if (!csrf) {
    try {
      csrf = await ensureCSRFToken();
    } catch {
      clearToken();
      clearCSRFToken();
      return false;
    }
  }

  const res = await fetch(joinUrl(getBaseUrl(), "/auth/refresh"), {
    method: "POST",
    headers: (() => {
      const headers = new Headers();
      headers.set("X-CSRF-Token", csrf);
      return headers;
    })(),
    credentials: "include",
  });

  if (!res.ok) {
    clearToken();
    return false;
  }

  if (res.status === 204) {
    clearToken();
    return false;
  }

  const data = await readJsonSafe<RefreshRes>(res);
  if (!data?.access_token) {
    clearToken();
    return false;
  }

  setToken(data.access_token);
  return true;
}

async function requestOnce<T>(
  path: string,
  opt?: ApiOption,
): Promise<T | undefined> {
  const init: RequestInit = {
    method: opt?.method || "GET",
    headers: buildHeaders(opt),
    credentials: "include",
  };

  if (opt?.body !== undefined) {
    init.body = JSON.stringify(opt.body);
  }

  const res = await fetch(joinUrl(getBaseUrl(), path), init);

  if (!res.ok) {
    await throwApiError(res);
  }

  if (res.status === 204) {
    return undefined;
  }

  const data = await readJsonSafe<T>(res);

  if (data === null) {
    throw new ApiError(500, "invalid_response", "response is not json");
  }

  return data;
}

export async function api<T>(
  path: string,
  opt?: ApiOption,
): Promise<T | undefined> {
  try {
    return await requestOnce<T>(path, opt);
  } catch (err: unknown) {
    if (
      opt?.auth &&
      err instanceof ApiError &&
      err.status === 401 &&
      path !== "/auth/refresh"
    ) {
      const ok = await tryRefresh();

      if (ok) {
        return await requestOnce<T>(path, opt);
      }
    }

    throw err;
  }
}

export function toErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof ApiError) {
    if (err.status === 401) {
      return "認証が切れました。再読み込みして、必要なら再ログインしてください。";
    }

    if (err.status === 403) {
      return "この操作を行う権限がありません。ログイン状態や対象データを確認してください。";
    }

    const msg = err.message.trim();
    if (msg !== "") {
      return msg;
    }
  }

  if (err instanceof Error) {
    const msg = err.message.trim();
    if (msg !== "") {
      return msg;
    }
  }

  return fallback;
}
