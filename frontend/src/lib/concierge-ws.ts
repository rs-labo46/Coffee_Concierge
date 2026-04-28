import type { SearchResult, Turn } from "./concierge";

type WsServerEvent =
  | {
      type: "turn.added";
      payload?: {
        turn?: Turn;
      };
    }
  | {
      type: "candidate.update";
      payload?: {
        result?: SearchResult;
      };
    }
  | {
      type: "related.update";
      payload?: unknown;
    }
  | {
      type: "ask.followup";
      payload?: {
        questions?: string[];
      };
    }
  | {
      type: "explain.delta";
      payload?: {
        rank?: number;
        text?: string;
        done?: boolean;
      };
    }
  | {
      type: "done";
      payload?: {
        session_id?: number;
        message?: string;
      };
    }
  | {
      type: "error";
      code?: string;
      message?: string;
    };

type WsClientTurnAddEvent = {
  type: "turn.add";
  session_id: number;
  body: string;
};

type WsClientPatchPrefEvent = {
  type: "pref.patch";
  session_id: number;
  diff: Record<string, unknown>;
};

type WsClientPingEvent = {
  type: "ping";
  session_id: number;
};

export type ConciergeWsStatus =
  | "idle"
  | "connecting"
  | "open"
  | "closed"
  | "error";

export type ConciergeWsHandlers = {
  onStatus?: (status: ConciergeWsStatus) => void;
  onTurnAdded?: (turn: Turn) => void;
  onResult?: (result: SearchResult) => void;
  onFollowups?: (questions: string[]) => void;
  onExplainDelta?: (payload: {
    rank?: number;
    text?: string;
    done?: boolean;
  }) => void;
  onError?: (message: string) => void;
  onDone?: () => void;
};

function getWsBaseURL(): string {
  const base = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

  if (base.startsWith("https://")) {
    return base.replace("https://", "wss://");
  }

  if (base.startsWith("http://")) {
    return base.replace("http://", "ws://");
  }

  return `ws://${base}`;
}

function buildGuestWsURL(sessionID: number, sessionKey: string): string {
  const base = getWsBaseURL().replace(/\/$/, "");
  const key = encodeURIComponent(sessionKey);

  return `${base}/ws/guest/search/sessions/${sessionID}?session_key=${key}`;
}

function safeParseEvent(value: string): WsServerEvent | null {
  try {
    return JSON.parse(value) as WsServerEvent;
  } catch {
    return null;
  }
}

export class ConciergeWsClient {
  private socket: WebSocket | null = null;
  private readonly handlers: ConciergeWsHandlers;

  constructor(handlers: ConciergeWsHandlers) {
    this.handlers = handlers;
  }

  connectGuest(sessionID: number, sessionKey: string): void {
    this.close();

    if (sessionID <= 0 || sessionKey.trim() === "") {
      this.handlers.onError?.("WebSocket接続に必要なsession情報がありません。");
      this.handlers.onStatus?.("error");
      return;
    }

    this.handlers.onStatus?.("connecting");

    const socket = new WebSocket(buildGuestWsURL(sessionID, sessionKey));
    this.socket = socket;

    socket.addEventListener("open", () => {
      this.handlers.onStatus?.("open");
    });

    socket.addEventListener("close", () => {
      if (this.socket === socket) {
        this.handlers.onStatus?.("closed");
      }
    });

    socket.addEventListener("error", () => {
      this.handlers.onStatus?.("error");
      this.handlers.onError?.("WebSocket接続でエラーが発生しました。");
    });

    socket.addEventListener("message", (event) => {
      if (typeof event.data !== "string") {
        return;
      }

      const data = safeParseEvent(event.data);
      if (!data) {
        return;
      }

      this.handleEvent(data);
    });
  }

  sendTurn(sessionID: number, body: string): boolean {
    const text = body.trim();

    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.handlers.onError?.("WebSocketが接続されていません。");
      return false;
    }

    if (text === "") {
      this.handlers.onError?.("希望を入力してください。");
      return false;
    }

    const event: WsClientTurnAddEvent = {
      type: "turn.add",
      session_id: sessionID,
      body: text,
    };

    this.socket.send(JSON.stringify(event));
    return true;
  }

  sendPatch(sessionID: number, diff: Record<string, unknown>): boolean {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.handlers.onError?.("WebSocketが接続されていません。");
      return false;
    }

    const event: WsClientPatchPrefEvent = {
      type: "pref.patch",
      session_id: sessionID,
      diff,
    };

    this.socket.send(JSON.stringify(event));
    return true;
  }

  ping(sessionID: number): boolean {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return false;
    }

    const event: WsClientPingEvent = {
      type: "ping",
      session_id: sessionID,
    };

    this.socket.send(JSON.stringify(event));
    return true;
  }

  close(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }

    this.handlers.onStatus?.("closed");
  }

  private handleEvent(event: WsServerEvent): void {
    switch (event.type) {
      case "turn.added":
        if (event.payload?.turn) {
          this.handlers.onTurnAdded?.(event.payload.turn);
        }
        return;

      case "candidate.update":
        if (event.payload?.result) {
          this.handlers.onResult?.(event.payload.result);
        }
        return;

      case "ask.followup":
        this.handlers.onFollowups?.(event.payload?.questions || []);
        return;

      case "explain.delta":
        this.handlers.onExplainDelta?.(event.payload || {});
        return;

      case "done":
        this.handlers.onDone?.();
        return;

      case "error":
        this.handlers.onError?.(event.message || event.code || "WebSocket error");
        return;

      case "related.update":
        return;

      default:
        return;
    }
  }
}
