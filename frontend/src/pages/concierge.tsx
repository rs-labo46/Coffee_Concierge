import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../auth/use-auth";
import { ApiError, toErrorMessage } from "../lib/api";
import {
  deleteSavedSuggestion,
  hasAccessToken,
  patchSearchPref,
  saveSuggestion,
  setSearchPref,
  startSearchSession,
} from "../lib/concierge-api";
import { ConciergeWsClient, type ConciergeWsStatus } from "../lib/concierge-ws";
import {
  defaultPrefInput,
  itemKindLabel,
  methodLabel,
  moodLabel,
  roastLabel,
  sceneLabel,
  tempPrefLabel,
  type Bean,
  type ConciergeItem,
  type PatchPrefInput,
  type Pref,
  type Recipe,
  type SearchResult,
  type SearchSession,
  type SetPrefInput,
  type Suggestion,
} from "../lib/concierge";

type LocalTurn = {
  id: number;
  role: "user" | "system";
  body: string;
};
type StoredConciergeState = {
  session: SearchSession;
  sessionKey: string;
  pref: Pref | null;
  result: SearchResult | null;
  turns: LocalTurn[];
  savedIDs: number[];
  chatText: string;
  loggedIn: boolean;
  savedAt: number;
};

const conciergeStateKey = "coffee_concierge_state_v1";
const conciergeStateTTLMS = 24 * 60 * 60 * 1000;

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function loadConciergeState(): StoredConciergeState | null {
  const raw = sessionStorage.getItem(conciergeStateKey);

  if (!raw) {
    return null;
  }

  try {
    const parsed: unknown = JSON.parse(raw);

    if (!isRecord(parsed)) {
      return null;
    }

    const savedAt = parsed.savedAt;
    const loggedIn = parsed.loggedIn;

    if (typeof savedAt !== "number") {
      return null;
    }

    if (typeof loggedIn !== "boolean") {
      return null;
    }

    if (Date.now() - savedAt > conciergeStateTTLMS) {
      sessionStorage.removeItem(conciergeStateKey);
      return null;
    }

    if (!isRecord(parsed.session)) {
      return null;
    }

    const sessionID = parsed.session.id;
    const sessionKey = parsed.sessionKey;

    if (typeof sessionID !== "number" || sessionID <= 0) {
      return null;
    }

    if (typeof sessionKey !== "string") {
      return null;
    }

    return parsed as StoredConciergeState;
  } catch {
    sessionStorage.removeItem(conciergeStateKey);
    return null;
  }
}

function saveConciergeState(input: StoredConciergeState): void {
  sessionStorage.setItem(conciergeStateKey, JSON.stringify(input));
}

function clearConciergeState(): void {
  sessionStorage.removeItem(conciergeStateKey);
}
const exampleTexts = [
  "朝に軽めで飲みたい",
  "酸味は弱めでミルクに合うもの",
  "食後に苦めでどっしりした一杯",
  "アイスで飲みやすいもの",
];

function textToPatchInput(text: string): PatchPrefInput {
  const note = text.trim();

  // 自然文の味覚判定はfrontendでは行わない。
  return { note };
}

function mergePatchToSetPref(
  current: SetPrefInput,
  patch: PatchPrefInput,
): SetPrefInput {
  return {
    ...current,
    ...patch,
    flavor: patch.flavor ?? current.flavor,
    acidity: patch.acidity ?? current.acidity,
    bitterness: patch.bitterness ?? current.bitterness,
    body: patch.body ?? current.body,
    aroma: patch.aroma ?? current.aroma,
    mood: patch.mood ?? current.mood,
    method: patch.method ?? current.method,
    scene: patch.scene ?? current.scene,
    temp_pref: patch.temp_pref ?? current.temp_pref,
    excludes: patch.excludes ?? current.excludes,
    note: patch.note ?? current.note,
  };
}

function beanByID(result: SearchResult, suggestion: Suggestion): Bean | null {
  const bean =
    result.beans.find((candidate) => candidate.id === suggestion.bean_id) ||
    null;

  if (bean) {
    return bean;
  }

  if (suggestion.bean && suggestion.bean.id > 0) {
    return suggestion.bean;
  }

  return null;
}

function recipeByID(
  result: SearchResult,
  suggestion: Suggestion,
): Recipe | null {
  if (!suggestion.recipe_id) {
    return null;
  }

  const recipe =
    result.recipes.find((candidate) => candidate.id === suggestion.recipe_id) ||
    null;

  if (recipe) {
    return recipe;
  }

  if (suggestion.recipe && suggestion.recipe.id > 0) {
    return suggestion.recipe;
  }

  return null;
}

function itemByID(
  result: SearchResult,
  suggestion: Suggestion,
): ConciergeItem | null {
  if (!suggestion.item_id) {
    return null;
  }

  const item =
    result.items.find((candidate) => candidate.id === suggestion.item_id) ||
    null;

  if (item) {
    return item;
  }

  if (suggestion.item && suggestion.item.id > 0) {
    return suggestion.item;
  }

  return null;
}

function relatedItemScore(
  item: ConciergeItem,
  bean: Bean | null,
  suggestion: Suggestion,
): number {
  let score = 0;
  const text = `${item.title} ${item.summary}`.toLowerCase();

  if (suggestion.item_id === item.id) {
    score += 100;
  }

  if (bean) {
    const beanName = bean.name.toLowerCase();
    const origin = bean.origin.toLowerCase();
    const roast = bean.roast.toLowerCase();
    const roastText = roastLabel(bean.roast).toLowerCase();

    if (text.includes(beanName)) {
      score += 40;
    }
    if (origin !== "" && text.includes(origin)) {
      score += 16;
    }
    if (text.includes(roast) || text.includes(roastText)) {
      score += 8;
    }
  }

  switch (item.kind) {
    case "recipe":
      score += 8;
      break;
    case "shop":
      score += 7;
      break;
    case "deal":
      score += 6;
      break;
    case "news":
      score += 5;
      break;
    default:
      score += 1;
      break;
  }

  return score;
}

function relatedItemsBySuggestion(
  result: SearchResult,
  suggestion: Suggestion,
  bean: Bean | null,
): ConciergeItem[] {
  const byID = new Map<number, ConciergeItem>();

  const primary = itemByID(result, suggestion);
  if (primary) {
    byID.set(primary.id, primary);
  }

  for (const item of result.items) {
    if (item.id > 0) {
      byID.set(item.id, item);
    }
  }

  return [...byID.values()]
    .sort(
      (a, b) =>
        relatedItemScore(b, bean, suggestion) -
        relatedItemScore(a, bean, suggestion),
    )
    .slice(0, 3);
}

function RelatedItemCard({ item }: { item: ConciergeItem }) {
  return (
    <article className="rounded-2xl border border-[#eadfd4] bg-white px-4 py-4">
      <p className="mb-1 text-[11px] font-black tracking-[0.18em] text-[#a1775b] uppercase">
        {itemKindLabel(item.kind)}
      </p>
      <h5 className="line-clamp-2 text-base font-black leading-6 text-[#4e342e]">
        {item.title}
      </h5>
      <p className="mt-2 line-clamp-3 text-sm font-semibold leading-6 text-[#7a6b62]">
        {item.summary}
      </p>
      <Link
        to={`/items/${item.id}`}
        className="mt-3 inline-flex min-h-10 items-center justify-center rounded-full border border-[#d8c5b8] px-4 py-2 text-sm font-black text-[#7b523a] transition hover:bg-[#f8efe7]"
      >
        関連情報を見る
      </Link>
    </article>
  );
}

function FieldShell({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <label className="block">
      <span className="mb-2 block text-sm font-black text-[#5f4a40]">
        {label}
      </span>
      {children}
    </label>
  );
}

function AxisBadge({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-2xl border border-[#eadfd5] bg-[#fffaf5] px-3 py-2 text-center">
      <p className="text-[11px] font-black tracking-[0.12em] text-[#9a755d] uppercase">
        {label}
      </p>
      <p className="text-lg font-black text-[#4e342e]">{value}</p>
    </div>
  );
}

function ConditionPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-full border border-[#eadfd4] bg-white px-4 py-2 text-sm font-black text-[#6f4e37]">
      <span className="mr-2 text-[#a1775b]">{label}</span>
      {value}
    </div>
  );
}

function SuggestionRow({
  suggestion,
  result,
  displayRank,
  selected,
  onSelect,
}: {
  suggestion: Suggestion;
  result: SearchResult;
  displayRank: number;
  selected: boolean;
  onSelect: (id: number) => void;
}) {
  const bean = beanByID(result, suggestion);

  return (
    <button
      type="button"
      onClick={() => onSelect(suggestion.id)}
      className={
        selected
          ? "flex w-full items-start gap-4 border-b border-[#eadfd4] bg-[#fffaf5] px-3 py-4 text-left transition last:border-b-0 cursor-pointer"
          : "flex w-full items-start gap-4 border-b border-[#eadfd4] px-3 py-4 text-left transition last:border-b-0 hover:bg-[#fffaf5]"
      }
    >
      <span
        className={
          selected
            ? "mt-1 inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#7b523a] text-base font-black text-white"
            : "mt-1 inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#f3e7dc] text-base font-black text-[#7b523a]"
        }
      >
        {displayRank}
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <h3 className="line-clamp-1 text-[20px] font-black leading-8 text-[#4e342e]">
            {bean?.name || `Bean #${suggestion.bean_id}`}
          </h3>
          <span className="rounded-full bg-[#b08968] px-3 py-1 text-xs font-black text-white">
            SCORE {suggestion.score}
          </span>
        </div>

        <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm font-bold text-[#8a756a]">
          {bean ? <span>{roastLabel(bean.roast)}</span> : null}
          {bean?.origin ? <span>{bean.origin}</span> : null}
          <span>RANK {displayRank}</span>
        </div>

        <p className="mt-2 line-clamp-2 text-sm font-semibold leading-6 text-[#6f625b]">
          {suggestion.reason || "この候補に合う理由はまだ生成されていません。"}
        </p>
      </div>
    </button>
  );
}

function SuggestionCard({
  suggestion,
  result,
  displayRank,
  canSave,
  savingID,
  savedIDs,
  onToggleSave,
}: {
  suggestion: Suggestion;
  result: SearchResult;
  displayRank: number;
  canSave: boolean;
  savingID: number | null;
  savedIDs: number[];
  onToggleSave: (suggestion: Suggestion) => Promise<void>;
}) {
  const bean = beanByID(result, suggestion);
  const recipe = recipeByID(result, suggestion);
  const relatedItems = relatedItemsBySuggestion(result, suggestion, bean);
  const saved = savedIDs.includes(suggestion.id);

  return (
    <article className="overflow-hidden rounded-[30px] border border-[#eadfd4] bg-white shadow-[0_12px_30px_rgba(110,78,56,0.08)]">
      <div className="border-b border-[#eadfd4] bg-[#fbf4ec] px-5 py-4">
        <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <div className="flex items-start gap-4">
            <div className="flex h-16 w-16 shrink-0 flex-col items-center justify-center rounded-2xl bg-[#4e342e] text-white shadow-[0_8px_18px_rgba(78,52,46,0.18)]">
              <span className="text-[10px] font-black tracking-[0.16em] uppercase text-[#f2dcc8]">
                rank
              </span>
              <span className="text-2xl font-black leading-none">
                {displayRank}
              </span>
            </div>

            <div>
              <p className="mb-2 inline-flex rounded-full bg-white px-3 py-1 text-sm font-black text-[#7b523a]">
                SCORE {suggestion.score}
              </p>
              <h3 className="text-2xl font-black text-[#4e342e]">
                {bean?.name || `Bean #${suggestion.bean_id}`}
              </h3>
              {bean ? (
                <p className="mt-1 text-sm font-bold text-[#7a6b62]">
                  {roastLabel(bean.roast)} / {bean.origin}
                </p>
              ) : null}
            </div>
          </div>

          {canSave ? (
            <button
              type="button"
              disabled={savingID === suggestion.id}
              onClick={() => void onToggleSave(suggestion)}
              className={
                saved
                  ? "inline-flex min-h-11 items-center justify-center rounded-full border border-[#d8c5b8] bg-white px-5 py-2 text-sm font-black text-[#7b523a] transition hover:bg-[#f8efe7] disabled:cursor-not-allowed disabled:opacity-60 cursor-pointer"
                  : "inline-flex min-h-11 items-center justify-center rounded-full bg-[#4e342e] px-5 py-2 text-sm font-black text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:bg-[#bca99c]"
              }
            >
              {savingID === suggestion.id
                ? saved
                  ? "解除中..."
                  : "保存中..."
                : saved
                  ? "保存解除"
                  : "保存する"}
            </button>
          ) : (
            <Link
              to="/login"
              className="inline-flex min-h-11 items-center justify-center rounded-full border border-[#d8c5b8] bg-white px-5 py-2 text-sm font-black text-[#7b523a] transition hover:bg-[#f8efe7]"
            >
              保存はログイン後
            </Link>
          )}
        </div>
      </div>

      <div className="px-5 py-5">
        <div className="grid gap-5 lg:grid-cols-[1.2fr_0.8fr]">
          <div>
            <p className="mb-4 rounded-2xl bg-[#fffaf5] px-4 py-3 text-sm font-bold leading-7 text-[#5f4a40]">
              {suggestion.reason ||
                "この候補に合う理由はまだ生成されていません。"}
            </p>

            {bean ? (
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
                <AxisBadge label="flavor" value={bean.flavor} />
                <AxisBadge label="acidity" value={bean.acidity} />
                <AxisBadge label="bitter" value={bean.bitterness} />
                <AxisBadge label="body" value={bean.body} />
                <AxisBadge label="aroma" value={bean.aroma} />
              </div>
            ) : null}

            {bean?.desc ? (
              <p className="mt-4 text-sm font-semibold leading-7 text-[#77685f]">
                {bean.desc}
              </p>
            ) : null}
          </div>

          <div>
            {recipe ? (
              <section className="rounded-3xl border border-[#eadfd4] bg-[#fcf8f3] p-4">
                <p className="mb-1 text-xs font-black tracking-[0.2em] text-[#a1775b] uppercase">
                  recipe
                </p>
                <h4 className="text-lg font-black text-[#4e342e]">
                  {recipe.name}
                </h4>
                <p className="mt-2 text-sm font-bold text-[#74675f]">
                  {methodLabel(recipe.method)} /{" "}
                  {tempPrefLabel(recipe.temp_pref)} / {recipe.temp}℃
                </p>
                <p className="mt-2 text-sm font-semibold leading-6 text-[#7a6b62]">
                  {recipe.grind} / {recipe.ratio} / {recipe.time_sec}秒
                </p>
                {recipe.desc ? (
                  <p className="mt-2 text-sm font-semibold leading-6 text-[#7a6b62]">
                    {recipe.desc}
                  </p>
                ) : null}
              </section>
            ) : null}
          </div>
        </div>

        {relatedItems.length > 0 ? (
          <section className="mt-5 rounded-3xl border border-[#eadfd4] bg-[#fffaf5] p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <p className="mb-1 text-xs font-black tracking-[0.2em] text-[#a1775b] uppercase">
                  related items
                </p>
                <h4 className="text-lg font-black text-[#4e342e]">
                  関連する記事・ショップ・ニュース
                </h4>
              </div>
              <span className="rounded-full bg-white px-3 py-1 text-xs font-black text-[#7b523a]">
                {relatedItems.length}件
              </span>
            </div>

            <div className="grid gap-3">
              {relatedItems.map((relatedItem) => (
                <RelatedItemCard key={relatedItem.id} item={relatedItem} />
              ))}
            </div>
          </section>
        ) : null}
      </div>
    </article>
  );
}

function SuggestionDetailModal({
  open,
  suggestion,
  result,
  displayRank,
  canSave,
  savingID,
  savedIDs,
  onToggleSave,
  onClose,
}: {
  open: boolean;
  suggestion: Suggestion | null;
  result: SearchResult | null;
  displayRank: number;
  canSave: boolean;
  savingID: number | null;
  savedIDs: number[];
  onToggleSave: (suggestion: Suggestion) => Promise<void>;
  onClose: () => void;
}) {
  useEffect(() => {
    if (!open) {
      return undefined;
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  if (!open || !suggestion || !result) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-[#4e342e]/45 px-4 py-6">
      <div className="absolute inset-0" onClick={onClose} />

      <div className="relative z-10 max-h-[90vh] w-full max-w-[1080px] overflow-hidden rounded-[22px] border border-[#eadfd4] bg-white shadow-[0_20px_48px_rgba(110,78,56,0.25)]">
        <div className="flex items-start justify-between gap-4 border-b border-[#eadfd4] bg-[#fcf6f0] px-5 py-4">
          <div>
            <p className="text-sm font-black tracking-[0.24em] text-[#a1775b] uppercase">
              rank {displayRank} preview
            </p>
            <h2 className="mt-1 text-2xl font-black text-[#4e342e]">
              おすすめ詳細
            </h2>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-11 w-11 items-center justify-center rounded-full border border-[#d8c5b8] text-xl font-black text-[#7b523a] transition hover:bg-white cursor-pointer"
          >
            ×
          </button>
        </div>

        <div className="max-h-[calc(90vh-88px)] overflow-y-auto px-5 py-5 md:px-6 md:py-6">
          <SuggestionCard
            suggestion={suggestion}
            result={result}
            displayRank={displayRank}
            canSave={canSave}
            savingID={savingID}
            savedIDs={savedIDs}
            onToggleSave={onToggleSave}
          />
        </div>
      </div>
    </div>
  );
}

export function ConciergePage() {
  const { user } = useAuth();
  const [form, setForm] = useState<SetPrefInput>(defaultPrefInput);
  const [session, setSession] = useState<SearchSession | null>(null);
  const [sessionKey, setSessionKey] = useState<string>("");
  const [pref, setPref] = useState<Pref | null>(null);
  const [result, setResult] = useState<SearchResult | null>(null);
  const [selectedSuggestionID, setSelectedSuggestionID] = useState<
    number | null
  >(null);
  const [detailOpen, setDetailOpen] = useState<boolean>(false);
  const [chatText, setChatText] = useState<string>("");
  const [turns, setTurns] = useState<LocalTurn[]>([]);
  const [savedIDs, setSavedIDs] = useState<number[]>([]);
  const [savingID, setSavingID] = useState<number | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [message, setMessage] = useState<string>("");
  const [error, setError] = useState<string>("");
  const wsRef = useRef<ConciergeWsClient | null>(null);
  const resultRef = useRef<HTMLElement | null>(null);
  const [wsStatus, setWsStatus] = useState<ConciergeWsStatus>("idle");

  const loggedIn = !!user && hasAccessToken();
  const sessionOwnedByUser =
    session?.user_id !== null && session?.user_id !== undefined;
  const canSaveSuggestion = loggedIn && sessionOwnedByUser;

  const suggestions = useMemo(() => {
    return [...(result?.suggestions || [])].sort((a, b) => {
      if (b.score === a.score) {
        return a.rank - b.rank;
      }
      return b.score - a.score;
    });
  }, [result]);

  const selectedSuggestion = useMemo(() => {
    return (
      suggestions.find(
        (suggestion) => suggestion.id === selectedSuggestionID,
      ) || null
    );
  }, [selectedSuggestionID, suggestions]);

  const selectedDisplayRank = useMemo(() => {
    if (!selectedSuggestion) {
      return 0;
    }
    const index = suggestions.findIndex(
      (suggestion) => suggestion.id === selectedSuggestion.id,
    );
    return index >= 0 ? index + 1 : selectedSuggestion.rank;
  }, [selectedSuggestion, suggestions]);

  useEffect(() => {
    if (suggestions.length === 0) {
      setSelectedSuggestionID(null);
      setDetailOpen(false);
      return;
    }

    setSelectedSuggestionID((current) => {
      if (
        current &&
        suggestions.some((suggestion) => suggestion.id === current)
      ) {
        return current;
      }
      return null;
    });
  }, [suggestions]);

  useEffect(() => {
    return () => {
      wsRef.current?.close();
    };
  }, []);
  const addTurn = useCallback((role: LocalTurn["role"], body: string) => {
    setTurns((current) => [
      ...current,
      { id: Date.now() + current.length, role, body },
    ]);
  }, []);

  const scrollToResults = useCallback(() => {
    window.requestAnimationFrame(() => {
      window.requestAnimationFrame(() => {
        resultRef.current?.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      });
    });
  }, []);

  const connectGuestWs = useCallback(
    (targetSessionID: number, targetSessionKey: string) => {
      const client = new ConciergeWsClient({
        onStatus: setWsStatus,
        onResult: (nextResult) => {
          setResult(nextResult);
          addTurn(
            "system",
            "WebSocketで会話内容を反映して候補を更新しました。",
          );
          setMessage("会話内容をもとに候補を更新しました。");
          scrollToResults();
        },
        onError: (msg) => {
          setError(msg);
          setLoading(false);
        },
        onDone: () => {
          setLoading(false);
        },
      });

      wsRef.current = client;
      client.connectGuest(targetSessionID, targetSessionKey);
    },
    [addTurn, scrollToResults],
  );

  useEffect(() => {
    const stored = loadConciergeState();

    if (!stored) {
      return;
    }

    const storedIsGuest =
      stored.session.user_id === null || stored.session.user_id === undefined;

    if (stored.loggedIn !== loggedIn || (storedIsGuest && loggedIn)) {
      clearConciergeState();
      return;
    }

    setSession(stored.session);
    setSessionKey(stored.sessionKey);
    setPref(stored.pref);
    setResult(stored.result);
    setTurns(stored.turns);
    setSavedIDs(stored.savedIDs);
    setChatText(stored.chatText);
    setMessage("前回の検索結果を復元しました。");

    if (!loggedIn && stored.sessionKey.trim() !== "") {
      connectGuestWs(stored.session.id, stored.sessionKey);
    }
  }, [loggedIn, connectGuestWs]);

  useEffect(() => {
    if (!session || !result) {
      return;
    }

    saveConciergeState({
      session,
      sessionKey,
      pref,
      result,
      turns,
      savedIDs,
      chatText,
      loggedIn,
      savedAt: Date.now(),
    });
  }, [session, sessionKey, pref, result, turns, savedIDs, chatText, loggedIn]);

  async function startWithInput(input: SetPrefInput, userText: string) {
    clearConciergeState();
    const started = await startSearchSession(
      { title: "コーヒーAIコンシェルジュ" },
      loggedIn,
    );
    const nextSessionKey = started.session_key || "";
    const prefRes = await setSearchPref(
      started.session.id,
      input,
      nextSessionKey,
    );

    setSession(started.session);
    setSessionKey(nextSessionKey);
    setPref(prefRes.pref);
    setResult(prefRes.result);
    setSavedIDs([]);
    addTurn("user", userText);
    addTurn("system", "希望を条件に変換して候補を検索しました。");
    setMessage("条件に合う候補を表示しました。");

    if (!loggedIn && nextSessionKey !== "") {
      connectGuestWs(started.session.id, nextSessionKey);
    }
  }

  async function patchWithInput(input: PatchPrefInput, userText: string) {
    if (!session) {
      const nextForm = mergePatchToSetPref(form, input);
      setForm(nextForm);
      await startWithInput(nextForm, userText);
      return;
    }

    const patched = await patchSearchPref(session.id, input, sessionKey);
    setPref(patched.pref);
    setResult(patched.result);
    addTurn("user", userText);
    addTurn("system", "追加の希望を反映して再検索しました。");
    setMessage("条件を更新して再検索しました。");
  }

  async function onSendText(text: string) {
    const trimmed = text.trim();

    if (trimmed === "") {
      setError("希望を入力してください。");
      return;
    }

    if (!loggedIn) {
      setError("AI検索を利用するにはログインしてください。");
      return;
    }

    setLoading(true);
    setError("");
    setMessage("");

    try {
      if (session && !loggedIn && wsRef.current && wsStatus === "open") {
        const ok = wsRef.current.sendTurn(session.id, trimmed);

        if (ok) {
          addTurn("user", trimmed);
          setChatText("");
          return;
        }
      }

      const input = textToPatchInput(trimmed);
      await patchWithInput(input, trimmed);
      setChatText("");
      scrollToResults();
    } catch (err: unknown) {
      setError(toErrorMessage(err, "希望を送信できませんでした。"));
    } finally {
      if (!(session && !loggedIn && wsStatus === "open")) {
        setLoading(false);
      }
    }
  }

  async function onPatch(input: PatchPrefInput) {
    const label = input.note || "条件を再調整";

    setLoading(true);
    setError("");
    setMessage("");

    try {
      await patchWithInput(input, label);
      scrollToResults();
    } catch (err: unknown) {
      setError(toErrorMessage(err, "条件を更新できませんでした。"));
    } finally {
      setLoading(false);
    }
  }

  async function onToggleSave(suggestion: Suggestion) {
    if (!session) {
      setError("保存対象のsessionがありません。");
      return;
    }

    const saved = savedIDs.includes(suggestion.id);

    setSavingID(suggestion.id);
    setError("");
    setMessage("");

    try {
      if (saved) {
        await deleteSavedSuggestion(suggestion.id);

        setSavedIDs((current) => current.filter((id) => id !== suggestion.id));

        setMessage("保存を解除しました。");
        return;
      }

      await saveSuggestion({
        session_id: session.id,
        suggestion_id: suggestion.id,
      });

      setSavedIDs((current) => {
        if (current.includes(suggestion.id)) {
          return current;
        }

        return [...current, suggestion.id];
      });

      setMessage("提案を保存しました。");
    } catch (err: unknown) {
      if (err instanceof ApiError && err.status === 409) {
        setSavedIDs((current) => {
          if (current.includes(suggestion.id)) {
            return current;
          }

          return [...current, suggestion.id];
        });

        setMessage("この提案はすでに保存済みです。");
        return;
      }

      if (err instanceof ApiError && err.status === 404 && saved) {
        setSavedIDs((current) => current.filter((id) => id !== suggestion.id));

        setMessage("保存はすでに解除されています。");
        return;
      }

      setError(
        toErrorMessage(
          err,
          saved ? "保存を解除できませんでした。" : "保存できませんでした。",
        ),
      );
    } finally {
      setSavingID(null);
    }
  }

  const messagePanel = message ? (
    <div className="rounded-3xl border border-[#cfe1c8] bg-[#f5fff2] px-5 py-4 text-sm font-bold text-[#3f6b36]">
      {message}
    </div>
  ) : null;

  const currentConditionPanel = pref ? (
    <section className="rounded-[32px] border border-[#eadfd4] bg-white p-5 shadow-[0_10px_28px_rgba(110,78,56,0.06)]">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
            current condition
          </p>
          <h3 className="text-2xl font-black text-[#4e342e]">
            {moodLabel(pref.mood)} / {methodLabel(pref.method)} /{" "}
            {sceneLabel(pref.scene)} / {tempPrefLabel(pref.temp_pref)}
          </h3>
          <p className="mt-2 text-sm font-semibold text-[#7a6b62]">
            {pref.note || "補足メモなし"}
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          <ConditionPill label="香り" value={String(pref.flavor)} />
          <ConditionPill label="酸味" value={String(pref.acidity)} />
          <ConditionPill label="苦味" value={String(pref.bitterness)} />
          <ConditionPill label="コク" value={String(pref.body)} />
          <ConditionPill label="アロマ" value={String(pref.aroma)} />
        </div>
      </div>

      <div className="mt-5 flex flex-wrap gap-2">
        <button
          type="button"
          onClick={() =>
            void onPatch({ body: 2, note: "もう少し軽めにしたい" })
          }
          className="rounded-full bg-[#f3e7dc] px-4 py-2 text-sm font-black text-[#6f4e37] cursor-pointer"
        >
          軽めにする
        </button>
        <button
          type="button"
          onClick={() => void onPatch({ acidity: 1, note: "酸味を弱めたい" })}
          className="rounded-full bg-[#f3e7dc] px-4 py-2 text-sm font-black text-[#6f4e37] cursor-pointer"
        >
          酸味を弱める
        </button>
        <button
          type="button"
          onClick={() => void onPatch({ bitterness: 5, note: "苦めがよい" })}
          className="rounded-full bg-[#f3e7dc] px-4 py-2 text-sm font-black text-[#6f4e37] cursor-pointer"
        >
          苦め
        </button>
        <button
          type="button"
          onClick={() =>
            void onPatch({ method: "milk", note: "ミルクに合うものがよい" })
          }
          className="rounded-full bg-[#f3e7dc] px-4 py-2 text-sm font-black text-[#6f4e37] cursor-pointer"
        >
          ミルク向け
        </button>
        <button
          type="button"
          onClick={() =>
            void onPatch({
              method: "iced",
              temp_pref: "ice",
              note: "アイスで飲みたい",
            })
          }
          className="rounded-full bg-[#f3e7dc] px-4 py-2 text-sm font-black text-[#6f4e37] cursor-pointer"
        >
          アイス向け
        </button>
      </div>
    </section>
  ) : null;

  return (
    <main className="mx-auto max-w-[1440px] px-4 py-8 md:px-8">
      <section className="mb-8 overflow-hidden rounded-[36px] border border-[#eadfd4] bg-white shadow-[0_16px_40px_rgba(110,78,56,0.08)]">
        <div className="bg-gradient-to-r from-[#4e342e] via-[#7b523a] to-[#b08968] px-6 py-10 text-white md:px-10">
          <p className="mb-3 text-sm font-black tracking-[0.28em] uppercase text-[#f5dfca]">
            coffee ai concierge
          </p>
          <h2 className="max-w-4xl text-3xl font-black leading-tight md:text-5xl">
            今飲みたいコーヒーを、言葉で教えてください。
          </h2>
          <p className="mt-4 max-w-3xl text-base font-semibold leading-8 text-[#f8ecdf]">
            自然文を検索条件に整理し、登録済みの豆・レシピ・関連情報から候補を返します。
          </p>
        </div>

        <div className="grid gap-6 px-6 py-6 lg:grid-cols-[1fr_0.75fr] md:px-10">
          <div className="rounded-[30px] border border-[#eadfd4] bg-[#fffaf5] p-5">
            <FieldShell label="コンシェルジュに希望を伝える">
              <textarea
                value={chatText}
                onChange={(event) => setChatText(event.target.value)}
                rows={1}
                placeholder="例: 朝に軽めで飲みたい。酸味は弱めで、ミルクにも合うもの。"
                className="w-full rounded-2xl border border-[#dfcfc2] bg-white px-4 py-3 text-sm font-bold leading-7 text-[#4e342e] outline-none transition focus:border-[#7b523a] focus:ring-4 focus:ring-[#eadfd5]"
              />
            </FieldShell>

            {error ? (
              <div className="mt-4 rounded-3xl border border-[#e3b8a6] bg-[#fff4ef] px-5 py-4 text-sm font-bold text-[#8a3d25]">
                {error}
              </div>
            ) : null}

            <button
              type="button"
              disabled={loading}
              onClick={() => void onSendText(chatText)}
              className="mt-4 inline-flex min-h-12 w-full items-center justify-center rounded-full bg-[#4e342e] px-6 py-3 text-sm font-black text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:bg-[#bca99c] cursor-pointer"
            >
              {loading
                ? "検索中..."
                : session
                  ? "送信して再検索"
                  : "送信して検索"}
            </button>

            <div className="mt-4 flex flex-wrap gap-2">
              {exampleTexts.map((text) => (
                <button
                  key={text}
                  type="button"
                  onClick={() => void onSendText(text)}
                  className="rounded-full bg-[#f3e7dc] px-4 py-2 text-sm font-black text-[#6f4e37] transition hover:bg-[#ead9c9] cursor-pointer"
                >
                  {text}
                </button>
              ))}
            </div>

            <p className="mt-4 text-sm font-semibold leading-7 text-[#7a6b62]">
              {loggedIn
                ? "ログイン中なので、AI検索・提案保存・履歴再開に対応します。"
                : "AI検索はログイン後に利用できます。保存と履歴一覧もログイン後に使えます。"}
            </p>
            <p className="mt-2 text-xs font-black tracking-[0.12em] text-[#9a755d] uppercase">
              ws: {loggedIn ? "http fallback" : wsStatus}
            </p>
          </div>

          <div className="rounded-[30px] border border-[#eadfd4] bg-white p-5">
            <p className="mb-3 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
              conversation
            </p>
            <div className="max-h-[280px] space-y-3 overflow-auto pr-1">
              {turns.length > 0 ? (
                turns.map((turn) => (
                  <div
                    key={turn.id}
                    className={[
                      "rounded-2xl px-4 py-3 text-sm font-bold leading-7",
                      turn.role === "user"
                        ? "ml-6 bg-[#4e342e] text-white"
                        : "mr-6 bg-[#fbf4ec] text-[#5f4a40]",
                    ].join(" ")}
                  >
                    {turn.body}
                  </div>
                ))
              ) : (
                <div className="rounded-2xl bg-[#fbf4ec] px-4 py-3 text-sm font-bold leading-7 text-[#5f4a40]">
                  「朝向け」「酸味弱め」「ミルクに合う」など、普段の言葉で入力してください。
                </div>
              )}
            </div>
          </div>
        </div>

        {currentConditionPanel || messagePanel ? (
          <div className="space-y-5 px-6 pb-6 md:px-10">
            {currentConditionPanel}
            {messagePanel}
          </div>
        ) : null}
      </section>

      <section ref={resultRef} className="scroll-mt-8 space-y-5">
        {result && suggestions.length > 0 ? (
          <div className="space-y-5">
            <div className="rounded-[32px] border border-[#eadfd4] bg-white px-5 py-5 shadow-[0_12px_30px_rgba(110,78,56,0.06)]">
              <div className="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
                <div>
                  <p className="text-2xl font-black tracking-[0.24em] text-[#a1775b] uppercase">
                    ranking result
                  </p>
                  <h2 className="mt-1 text-2xl font-black text-[#4e342e]"></h2>
                </div>
                <span className="w-fit rounded-full bg-[#f3e7dc] px-4 py-2 text-sm font-black text-[#7b523a]">
                  {suggestions.length}件
                </span>
              </div>

              <div className="overflow-hidden rounded-[22px] border border-[#eadfd4] bg-white ">
                {suggestions.map((suggestion, index) => (
                  <SuggestionRow
                    key={suggestion.id}
                    suggestion={suggestion}
                    result={result}
                    displayRank={index + 1}
                    selected={
                      selectedSuggestion?.id === suggestion.id && detailOpen
                    }
                    onSelect={(id) => {
                      setSelectedSuggestionID(id);
                      setDetailOpen(true);
                    }}
                  />
                ))}
              </div>
            </div>
          </div>
        ) : (
          <div className="rounded-[32px] border border-dashed border-[#d8c5b8] bg-white px-6 py-12 text-center">
            <h3 className="text-2xl font-black text-[#4e342e]">
              まだ候補はありません
            </h3>
            <p className="mt-3 text-sm font-semibold text-[#7a6b62]">
              希望を言葉で入力すると、豆・レシピ・関連情報が表示されます。
            </p>
          </div>
        )}
      </section>

      <SuggestionDetailModal
        open={detailOpen}
        suggestion={selectedSuggestion}
        result={result}
        displayRank={selectedDisplayRank}
        canSave={canSaveSuggestion}
        savingID={savingID}
        savedIDs={savedIDs}
        onToggleSave={onToggleSave}
        onClose={() => setDetailOpen(false)}
      />
    </main>
  );
}
