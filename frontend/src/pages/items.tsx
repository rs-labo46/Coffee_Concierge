import { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { api, toErrorMessage } from "../lib/api";
import { formatDisplayDate, formatDisplayDateTime } from "../lib/date";
import {
  cardImage,
  hasRef,
  isItemKind,
  kindBadgeLabel,
  kindDescLabel,
  kindTitleLabel,
  previewText,
  type Item,
  type ItemKind,
  type ItemListRes,
} from "../lib/item";

function ItemCard({ item }: { item: Item }) {
  return (
    <article className="group flex h-full flex-col overflow-hidden rounded-[30px] border border-[#eadfd4] bg-white shadow-[0_10px_26px_rgba(110,78,56,0.08)]">
      <div className="aspect-[16/10] overflow-hidden border-b border-[#eadfd4] bg-[#f7efe8]">
        <img
          src={cardImage(item.image_url)}
          alt={item.title}
          className="h-full w-full object-cover transition duration-300 group-hover:scale-[1.02]"
        />
      </div>

      <div className="flex flex-1 flex-col px-5 py-5">
        <div className="mb-3 flex flex-wrap items-center gap-2 text-sm font-bold text-[#8a756a]">
          <span className="rounded-full border border-[#d8c5b8] bg-[#fffaf5] px-3 py-1 text-[11px] font-black tracking-[0.22em] text-[#7b523a]">
            {kindBadgeLabel(item.kind)}
          </span>
          <span>{formatDisplayDate(item.published_at)}</span>
        </div>

        <h2 className="mb-3 line-clamp-2 text-xl font-black leading-tight text-[#4e342e]">
          {item.title}
        </h2>

        <p className="mb-5 line-clamp-4 text-sm font-semibold leading-7 text-[#6f625b]">
          {previewText(item, 180)}
        </p>

        <div className="mt-auto flex items-center justify-between gap-3">
          <span className="text-xs font-bold text-[#9a8a80]">
            {hasRef(item.url) ? "外部参考あり" : "アプリ内詳細あり"}
          </span>

          <Link
            to={`/items/${item.id}`}
            className="inline-flex min-h-11 items-center justify-center rounded-full border border-[#7b523a] px-4 py-2 text-sm font-black text-[#7b523a] transition hover:bg-[#f8efe7]"
          >
            詳細を見る
          </Link>
        </div>
      </div>
    </article>
  );
}

export function ItemsPage() {
  const [params] = useSearchParams();
  const kindParam = params.get("kind");
  const kind: ItemKind = isItemKind(kindParam) ? kindParam : "news";

  const [items, setItems] = useState<Item[]>([]);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState("");

  useEffect(() => {
    async function run() {
      setLoading(true);
      setMsg("");

      try {
        const res = await api<ItemListRes>(
          `/items?kind=${kind}&limit=20&offset=0`,
          {
            method: "GET",
          },
        );

        setItems(res?.items ?? []);
      } catch (err: unknown) {
        setMsg(toErrorMessage(err, "一覧の取得に失敗しました。"));
      } finally {
        setLoading(false);
      }
    }

    void run();
  }, [kind]);

  const tabs = useMemo(
    () => [
      { kind: "news" as const, label: "主要" },
      { kind: "recipe" as const, label: "レシピ" },
      { kind: "deal" as const, label: "セール" },
      { kind: "shop" as const, label: "店舗" },
    ],
    [],
  );

  if (loading) {
    return (
      <main className="min-h-screen bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
        <div className="mx-auto max-w-[1280px] rounded-[30px] border border-[#eadfd4] bg-white px-8 py-16 text-center text-lg font-black text-[#4e342e] shadow-[0_10px_26px_rgba(110,78,56,0.08)]">
          読み込み中です...
        </div>
      </main>
    );
  }

  if (msg) {
    return (
      <main className="min-h-screen bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
        <div className="mx-auto max-w-[1280px] rounded-[30px] border border-[#e3b8a6] bg-white px-8 py-16 text-center text-lg font-black text-[#8a3d25] shadow-[0_10px_26px_rgba(110,78,56,0.08)]">
          {msg}
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto flex max-w-[1280px] flex-col gap-8">
        <section className="overflow-hidden rounded-[34px] border border-[#eadfd4] bg-white shadow-[0_14px_34px_rgba(110,78,56,0.08)]">
          <div className="border-b border-[#eadfd4] bg-[#fffaf5] px-6 py-7 md:px-8">
            <p className="mb-2 text-sm font-black tracking-[0.24em] text-[#a1775b] uppercase">
              topics archive
            </p>
            <h1 className="text-3xl font-black text-[#4e342e] md:text-4xl">
              {kindTitleLabel(kind)}一覧
            </h1>
            <p className="mt-3 text-sm font-bold text-[#7a6b62]">
              {kindDescLabel(kind)} をまとめて確認できます。
            </p>
            <p className="mt-2 text-xs font-bold text-[#9a8a80]">
              更新目安: {" "}
              {items[0]
                ? formatDisplayDateTime(items[0].published_at)
                : "データなし"}
            </p>
          </div>

          <div className="flex flex-wrap gap-3 px-6 py-4 md:px-8">
            {tabs.map((tab) => (
              <Link
                key={tab.kind}
                to={`/items?kind=${tab.kind}`}
                className={[
                  "rounded-full border px-4 py-2 text-sm font-black transition",
                  tab.kind === kind
                    ? "border-[#7b523a] bg-[#7b523a] text-white shadow-[0_6px_14px_rgba(110,78,56,0.16)]"
                    : "border-[#eadfd4] bg-white text-[#7b523a] hover:bg-[#f8efe7]",
                ].join(" ")}
              >
                {tab.label}
              </Link>
            ))}

            <Link
              to="/"
              className="rounded-full border border-[#eadfd4] bg-white px-4 py-2 text-sm font-black text-[#7b523a] transition hover:bg-[#f8efe7]"
            >
              トップへ戻る
            </Link>
          </div>
        </section>

        {items.length === 0 ? (
          <section className="rounded-[30px] border border-dashed border-[#d8c5b8] bg-white px-8 py-16 text-center shadow-[0_10px_26px_rgba(110,78,56,0.08)]">
            <p className="text-base font-black text-[#6f625b]">
              表示できるデータがありません。
            </p>
          </section>
        ) : (
          <section className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
            {items.map((item) => (
              <ItemCard key={item.id} item={item} />
            ))}
          </section>
        )}
      </div>
    </main>
  );
}
