import { ApiError, api, getToken } from "./api";
import type {
  PatchPrefInput,
  PatchPrefResponse,
  SavedSuggestion,
  SaveSuggestionInput,
  SaveSuggestionResponse,
  SetPrefInput,
  SetPrefResponse,
  StartSessionResponse,
} from "./concierge";

type JsonPrimitive = string | number | boolean | null;
type SavedSuggestionListResponse = {
  saved: SavedSuggestion[];
};
type JsonValue = JsonPrimitive | JsonValue[] | { [key: string]: JsonValue };

function need<T>(value: T | undefined): T {
  if (value === undefined) {
    throw new ApiError(500, "invalid_response", "response is empty");
  }

  return value;
}

function authEnabled(sessionKey: string): boolean {
  return sessionKey.trim() === "" && getToken().trim() !== "";
}

function sessionHeaders(
  sessionKey: string,
): Record<string, string> | undefined {
  const key = sessionKey.trim();

  if (key === "") {
    return undefined;
  }

  return { "X-Session-Key": key };
}

function setPrefBody(input: SetPrefInput): { [key: string]: JsonValue } {
  return {
    flavor: input.flavor,
    acidity: input.acidity,
    bitterness: input.bitterness,
    body: input.body,
    aroma: input.aroma,
    mood: input.mood,
    method: input.method,
    scene: input.scene,
    temp_pref: input.temp_pref,
    excludes: input.excludes,
    note: input.note,
  };
}

function patchPrefBody(input: PatchPrefInput): { [key: string]: JsonValue } {
  const body: { [key: string]: JsonValue } = {};

  if (input.flavor !== undefined) {
    body.flavor = input.flavor;
  }
  if (input.acidity !== undefined) {
    body.acidity = input.acidity;
  }
  if (input.bitterness !== undefined) {
    body.bitterness = input.bitterness;
  }
  if (input.body !== undefined) {
    body.body = input.body;
  }
  if (input.aroma !== undefined) {
    body.aroma = input.aroma;
  }
  if (input.mood !== undefined) {
    body.mood = input.mood;
  }
  if (input.method !== undefined) {
    body.method = input.method;
  }
  if (input.scene !== undefined) {
    body.scene = input.scene;
  }
  if (input.temp_pref !== undefined) {
    body.temp_pref = input.temp_pref;
  }
  if (input.excludes !== undefined) {
    body.excludes = input.excludes;
  }
  if (input.note !== undefined) {
    body.note = input.note;
  }

  return body;
}

export function hasAccessToken(): boolean {
  return getToken().trim() !== "";
}

export async function startSearchSession(
  input: { title: string },
  loggedIn: boolean,
): Promise<StartSessionResponse> {
  const res = await api<StartSessionResponse>("/search/sessions", {
    method: "POST",
    auth: loggedIn,
    body: { title: input.title },
  });

  return need(res);
}

export async function setSearchPref(
  sessionID: number,
  input: SetPrefInput,
  sessionKey: string,
): Promise<SetPrefResponse> {
  const res = await api<SetPrefResponse>(`/search/sessions/${sessionID}/pref`, {
    method: "POST",
    auth: authEnabled(sessionKey),
    headers: sessionHeaders(sessionKey),
    body: setPrefBody(input),
  });

  return need(res);
}

export async function patchSearchPref(
  sessionID: number,
  input: PatchPrefInput,
  sessionKey: string,
): Promise<PatchPrefResponse> {
  const res = await api<PatchPrefResponse>(
    `/search/sessions/${sessionID}/pref`,
    {
      method: "PATCH",
      auth: authEnabled(sessionKey),
      headers: sessionHeaders(sessionKey),
      body: patchPrefBody(input),
    },
  );

  return need(res);
}

export async function saveSuggestion(
  input: SaveSuggestionInput,
): Promise<SaveSuggestionResponse> {
  const res = await api<SaveSuggestionResponse>("/saved-suggestions", {
    method: "POST",
    auth: true,
    body: {
      session_id: input.session_id,
      suggestion_id: input.suggestion_id,
    },
  });

  return need(res);
}

export async function listSavedSuggestions(input?: {
  limit?: number;
  offset?: number;
}): Promise<SavedSuggestion[]> {
  const limit = input?.limit ?? 20;
  const offset = input?.offset ?? 0;

  const params = new URLSearchParams();
  params.set("limit", String(limit));
  params.set("offset", String(offset));

  const res = await api<SavedSuggestionListResponse>(
    `/saved-suggestions?${params.toString()}`,
    {
      method: "GET",
      auth: true,
    },
  );

  return need(res).saved;
}
