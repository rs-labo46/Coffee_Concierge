import { Link } from "react-router-dom";
import { useAuth } from "../auth/use-auth";
import { useEffect, useState } from "react";
import {
  itemKindLabel,
  methodLabel,
  roastLabel,
  tempPrefLabel,
  type SavedSuggestion,
} from "../lib/concierge";
import { listSavedSuggestions } from "../lib/concierge-api";
import { toErrorMessage } from "../lib/api";
function formatDate(value: string): string {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function SavedSuggestionCard({ saved }: { saved: SavedSuggestion }) {
  const suggestion = saved.suggestion;
  const bean =
    suggestion?.bean && suggestion.bean.id > 0 ? suggestion.bean : null;
  const recipe =
    suggestion?.recipe && suggestion.recipe.id > 0 ? suggestion.recipe : null;
  const item =
    suggestion?.item && suggestion.item.id > 0 ? suggestion.item : null;

  return (
    <article className="rounded-[28px] border border-[#eadfd5] bg-white px-5 py-5 shadow-[0_8px_20px_rgba(110,78,56,0.06)]">
      <div className="mb-4 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
        <div>
          <p className="mb-2 text-xs font-black tracking-[0.22em] text-[#a1775b] uppercase">
            saved suggestion #{saved.suggestion_id}
          </p>

          <h3 className="text-xl font-black text-[#4e342e]">
            {bean?.name || `Suggestion #${saved.suggestion_id}`}
          </h3>

          {bean ? (
            <p className="mt-1 text-sm font-bold text-[#7a6b62]">
              {roastLabel(bean.roast)} / {bean.origin}
            </p>
          ) : (
            <p className="mt-1 text-sm font-bold text-[#7a6b62]">
              豆情報を取得できませんでした。
            </p>
          )}
        </div>

        <span className="rounded-full bg-[#f4ebe3] px-4 py-2 text-xs font-black text-[#7b523a]">
          {formatDate(saved.created_at)}
        </span>
      </div>

      {suggestion?.reason ? (
        <p className="rounded-2xl bg-[#fffaf5] px-4 py-3 text-sm font-bold leading-7 text-[#5f4a40]">
          {suggestion.reason}
        </p>
      ) : null}

      <div className="mt-4 grid gap-3 md:grid-cols-2">
        {recipe ? (
          <section className="rounded-2xl bg-[#fcf6f0] px-4 py-4">
            <p className="mb-1 text-xs font-black tracking-[0.18em] text-[#a1775b] uppercase">
              recipe
            </p>
            <p className="text-base font-black text-[#4e342e]">{recipe.name}</p>
            <p className="mt-1 text-sm font-semibold text-[#766b63]">
              {methodLabel(recipe.method)} / {tempPrefLabel(recipe.temp_pref)} /{" "}
              {recipe.temp}℃
            </p>
          </section>
        ) : null}

        {item ? (
          <section className="rounded-2xl bg-[#fcf6f0] px-4 py-4">
            <p className="mb-1 text-xs font-black tracking-[0.18em] text-[#a1775b] uppercase">
              {itemKindLabel(item.kind)}
            </p>
            <p className="line-clamp-2 text-base font-black text-[#4e342e]">
              {item.title}
            </p>
            <Link
              to={`/items/${item.id}`}
              className="mt-3 inline-flex rounded-full border border-[#d8c5b8] bg-white px-4 py-2 text-sm font-black text-[#7b523a] transition hover:bg-[#f8efe7]"
            >
              関連情報を見る
            </Link>
          </section>
        ) : null}
      </div>
    </article>
  );
}

function SavedSuggestionsSection() {
  const [savedList, setSavedList] = useState<SavedSuggestion[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    let active = true;

    async function load() {
      setLoading(true);
      setError("");

      try {
        const list = await listSavedSuggestions({ limit: 20, offset: 0 });

        if (active) {
          setSavedList(list);
        }
      } catch (err: unknown) {
        if (active) {
          setError(toErrorMessage(err, "保存済み提案を取得できませんでした。"));
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    void load();

    return () => {
      active = false;
    };
  }, []);

  return (
    <section className="rounded-[30px] border border-[#eadfd5] bg-white px-6 py-6">
      <div className="mb-5 flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="mb-2 text-sm font-black tracking-[0.24em] text-[#a1775b] uppercase">
            saved suggestions
          </p>
          <h2 className="text-2xl font-black text-[#4e342e]">保存済みの提案</h2>
        </div>

        <Link
          to="/concierge"
          className="inline-flex rounded-full bg-[#4e342e] px-5 py-3 text-sm font-black text-white transition hover:opacity-90"
        >
          コンシェルジュで探す
        </Link>
      </div>

      {loading ? (
        <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-6 text-sm font-bold text-[#766b63]">
          保存済み提案を読み込み中です。
        </div>
      ) : null}

      {error ? (
        <div className="rounded-[24px] border border-[#e3b8a6] bg-[#fff4ef] px-5 py-4 text-sm font-bold text-[#8a3d25]">
          {error}
        </div>
      ) : null}

      {!loading && !error && savedList.length === 0 ? (
        <div className="rounded-[24px] border border-dashed border-[#d8c5b8] bg-[#fffaf5] px-5 py-8 text-center">
          <p className="text-lg font-black text-[#4e342e]">
            まだ保存済みの提案はありません。
          </p>
          <p className="mt-2 text-sm font-semibold text-[#766b63]">
            コンシェルジュで検索し、気に入った候補を保存してください。
          </p>
        </div>
      ) : null}

      {!loading && !error && savedList.length > 0 ? (
        <div className="grid gap-4">
          {savedList.map((saved) => (
            <SavedSuggestionCard key={saved.id} saved={saved} />
          ))}
        </div>
      ) : null}
    </section>
  );
}
export function MePage() {
  const { user } = useAuth();

  if (!user) {
    return <div>no user</div>;
  }

  const roleTone =
    user.role === "admin"
      ? "bg-[#f1e3d6] text-[#7b523a]"
      : "bg-[#ece6ff] text-[#6a55aa]";

  return (
    <main className="min-h-[calc(100vh-120px)] bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto flex max-w-[1280px] flex-col gap-6">
        <section className="overflow-hidden rounded-[36px] border border-[#e6d9ce] bg-[#fffdfa] shadow-[0_10px_28px_rgba(110,78,56,0.08)]">
          <div className="border-b border-[#eadfd5] bg-gradient-to-r from-[#6f4e37] via-[#8b6448] to-[#ceb39e] px-8 py-10 text-white md:px-10">
            <p className="mb-3 text-sm font-black tracking-[0.32em] text-white/80 uppercase">
              my page
            </p>

            <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
              <div>
                <h1 className="mb-3 text-4xl font-black leading-tight md:text-5xl">
                  マイページ
                </h1>
              </div>

              <div className="flex flex-wrap gap-3">
                {user.role === "admin" ? (
                  <Link
                    to="/admin"
                    className="inline-flex rounded-full border border-white/30 bg-white/10 px-5 py-3 text-sm font-bold text-white transition hover:bg-white/20"
                  >
                    管理画面へ
                  </Link>
                ) : null}
              </div>
            </div>
          </div>

          <div className="grid gap-6 px-8 py-8 md:px-10 md:py-10 lg:grid-cols-[1.05fr_0.95fr]">
            <div className="grid gap-6">
              <section className="rounded-[30px] border border-[#eadfd5] bg-white px-6 py-6">
                <div className="mb-5 flex flex-wrap items-center gap-3">
                  <span
                    className={`rounded-full px-4 py-2 text-xs font-black tracking-[0.24em] uppercase ${roleTone}`}
                  >
                    {user.role}
                  </span>

                  <span className="rounded-full bg-[#f4ebe3] px-4 py-2 text-xs font-black tracking-[0.24em] text-[#7b523a] uppercase">
                    {user.email_verified ? "verified" : "not verified"}
                  </span>
                </div>

                <h2 className="mb-4 text-2xl font-black text-[#4e342e]">
                  アカウント概要
                </h2>

                <div className="grid gap-4 md:grid-cols-2">
                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
                      token version
                    </p>
                    <p className="text-xl font-black text-[#4e342e]">
                      {user.token_ver}
                    </p>
                  </div>

                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
                      id
                    </p>
                    <p className="text-xl font-black text-[#4e342e]">
                      {user.id}
                    </p>
                  </div>

                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5 md:col-span-2">
                    <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
                      email
                    </p>
                    <p className="break-all text-lg font-black text-[#4e342e]">
                      {user.email}
                    </p>
                  </div>
                </div>
              </section>
            </div>

            <div className="grid gap-6">
              <section className="rounded-[30px] border border-[#eadfd5] bg-white px-6 py-6">
                <p className="mb-3 text-sm font-black tracking-[0.24em] text-[#a1775b] uppercase">
                  アカウントのステータス
                </p>

                <h2 className="mb-4 text-2xl font-black text-[#4e342e]">
                  現在の状態
                </h2>

                <div className="grid gap-4">
                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-sm font-black text-[#5f4a40]">
                      ロール
                    </p>
                    <p className="text-base font-semibold leading-7 text-[#766b63]">
                      現在のロールは{" "}
                      <span className="font-black text-[#4e342e]">
                        {user.role}
                      </span>{" "}
                      です。
                    </p>
                  </div>

                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-sm font-black text-[#5f4a40]">
                      メール確認
                    </p>
                    <p className="text-base font-semibold leading-7 text-[#766b63]">
                      状態は{" "}
                      <span className="font-black text-[#4e342e]">
                        {user.email_verified ? "確認済み" : "未確認"}
                      </span>{" "}
                      です。
                    </p>
                  </div>
                </div>
              </section>
            </div>
          </div>
          <div className="px-8 pb-8 md:px-10 md:pb-10">
            <SavedSuggestionsSection />
          </div>
        </section>
      </div>
    </main>
  );
}
