import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { api, toErrorMessage } from "../lib/api";
import { formatDisplayDateTime } from "../lib/date";
import {
  cardImage,
  halfPreviewText,
  isFreshItem,
  kindBadgeLabel,
  kindDescLabel,
  previewText,
  type Item,
  type ItemDetailRes,
  type ItemKind,
  type Source,
  type SourceListRes,
  type TopRes,
} from "../lib/item";

type KindTab = {
  kind: ItemKind;
  label: string;
};

const kindTabs: KindTab[] = [
  { kind: "news", label: "ニュース" },
  { kind: "recipe", label: "レシピ" },
  { kind: "deal", label: "セール" },
  { kind: "shop", label: "店舗" },
];

function itemListByKind(data: TopRes, kind: ItemKind): Item[] {
  switch (kind) {
    case "news":
      return data.news;
    case "recipe":
      return data.recipe;
    case "deal":
      return data.deal;
    case "shop":
      return data.shop;
    default:
      return [];
  }
}

function latestPublishedAt(data: TopRes): string {
  const times = [data.news, data.recipe, data.deal, data.shop]
    .flat()
    .map((item) => item.published_at)
    .filter((value) => value !== "")
    .sort((left, right) => right.localeCompare(left));

  return times[0] || "";
}

function sourceNameById(sources: Source[], sourceID: number): string {
  const found = sources.find((source) => source.id === sourceID);
  return found?.name || "未設定";
}

function HeadlineRow({
  item,
  sourceName,
  onOpen,
}: {
  item: Item;
  sourceName: string;
  onOpen: (id: number) => void;
}) {
  return (
    <button
      type="button"
      onClick={() => onOpen(item.id)}
      className="flex w-full items-start gap-3 border-b border-[#d8e1ef] px-1 py-3 text-left transition hover:bg-[#f6f9fe]"
    >
      <span className="mt-[10px] h-2.5 w-2.5 shrink-0 rounded-full bg-[#204b9b]" />

      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <h3 className="line-clamp-1 text-[18px] font-bold leading-8 text-[#16326e]">
            {item.title}
          </h3>
          {isFreshItem(item.published_at) ? (
            <span className="rounded-full bg-[#ffb400] px-2 py-0.5 text-[11px] font-black text-white">
              NEW
            </span>
          ) : null}
        </div>

        <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-semibold text-[#6b7ea6]">
          <span>{formatDisplayDateTime(item.published_at)}</span>
          <span>{sourceName}</span>
          <span>{kindBadgeLabel(item.kind)}</span>
        </div>
      </div>
    </button>
  );
}

function DetailModal({
  open,
  item,
  source,
  loading,
  msg,
  onClose,
}: {
  open: boolean;
  item: Item | null;
  source: Source | null;
  loading: boolean;
  msg: string;
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

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-[#12284a]/45 px-4 py-6">
      <div className="absolute inset-0" onClick={onClose} />

      <div className="relative z-10 max-h-[90vh] w-full max-w-[840px] overflow-hidden rounded-[20px] border border-[#b8c8e2] bg-white shadow-[0_20px_48px_rgba(18,40,74,0.25)]">
        <div className="flex items-start justify-between gap-4 border-b border-[#d7e0ef] bg-[#f4f7fc] px-5 py-4">
          <div>
            <p className="text-sm font-black tracking-[0.24em] text-[#4668ad] uppercase">
              preview
            </p>
            <h2 className="mt-1 text-2xl font-black text-[#16326e]">
              記事プレビュー
            </h2>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-11 w-11 items-center justify-center rounded-full border border-[#c7d5eb] text-xl font-black text-[#2a4fa3] transition hover:bg-white"
          >
            ×
          </button>
        </div>

        <div className="max-h-[calc(90vh-88px)] overflow-y-auto px-5 py-5 md:px-6 md:py-6">
          {loading ? (
            <div className="rounded-[16px] border border-[#d7e0ef] bg-[#f7faff] px-5 py-10 text-center text-base font-bold text-[#355184]">
              読み込み中です...
            </div>
          ) : msg ? (
            <div className="rounded-[16px] border border-[#f0c7c2] bg-[#fff4f2] px-5 py-10 text-center text-base font-bold text-[#8a4b3a]">
              {msg}
            </div>
          ) : item ? (
            <div className="space-y-5">
              <img
                src={cardImage(item.image_url)}
                alt={item.title}
                className="h-[240px] w-full rounded-[16px] border border-[#d7e0ef] object-cover"
              />

              <div className="flex flex-wrap items-center gap-2 text-xs font-semibold text-[#6b7ea6]">
                <span className="rounded-full border border-[#a9bbdc] bg-[#eef3fb] px-3 py-1 font-black tracking-[0.2em] text-[#2a4fa3]">
                  {kindBadgeLabel(item.kind)}
                </span>
                <span>{formatDisplayDateTime(item.published_at)}</span>
                <span>{source?.name || "未設定"}</span>
              </div>

              <h3 className="text-2xl font-black leading-tight text-[#16326e]">
                {item.title}
              </h3>

              {item.summary ? (
                <p className="rounded-[16px] border border-[#d7e0ef] bg-[#f7faff] px-4 py-4 text-sm font-semibold leading-7 text-[#30466d]">
                  {item.summary}
                </p>
              ) : null}

              <p className="text-base font-medium leading-8 text-[#30466d]">
                {halfPreviewText(item)}
              </p>

              <div className="flex flex-wrap justify-center gap-3">
                <Link
                  to={`/items/${item.id}`}
                  className="inline-flex min-h-11 items-center justify-center rounded-full bg-[#2a4fa3] px-5 py-2 text-sm font-bold text-white transition hover:opacity-90"
                >
                  もっと見る
                </Link>
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}

export default function TopPage() {
  const [data, setData] = useState<TopRes>({
    news: [],
    recipe: [],
    deal: [],
    shop: [],
  });
  const [sources, setSources] = useState<Source[]>([]);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState("");
  const [activeKind, setActiveKind] = useState<ItemKind>("news");
  const [modalOpen, setModalOpen] = useState(false);
  const [detailItem, setDetailItem] = useState<Item | null>(null);
  const [detailSource, setDetailSource] = useState<Source | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [detailMsg, setDetailMsg] = useState("");

  useEffect(() => {
    async function run() {
      setLoading(true);
      setMsg("");

      try {
        const [topRes, sourceRes] = await Promise.all([
          api<TopRes>("/items/top?limit=8", { method: "GET" }),
          api<SourceListRes>("/sources", { method: "GET" }),
        ]);

        setData(
          topRes || {
            news: [],
            recipe: [],
            deal: [],
            shop: [],
          },
        );
        setSources(sourceRes?.sources || []);
      } catch (err: unknown) {
        setMsg(toErrorMessage(err, "トップ情報の取得に失敗しました。"));
      } finally {
        setLoading(false);
      }
    }

    void run();
  }, []);

  const items = useMemo(
    () => itemListByKind(data, activeKind),
    [activeKind, data],
  );
  const featured = items[0] || null;
  const subItems = items.slice(0, 8);
  const updatedAt = useMemo(() => latestPublishedAt(data), [data]);

  async function openModal(itemID: number) {
    setModalOpen(true);
    setLoadingDetail(true);
    setDetailMsg("");

    try {
      const res = await api<ItemDetailRes>(`/items/${itemID}`, {
        method: "GET",
      });

      if (!res) {
        setDetailMsg("記事の取得に失敗しました。");
        setDetailItem(null);
        setDetailSource(null);
        return;
      }

      setDetailItem(res.item);
      setDetailSource(res.source);
    } catch (err: unknown) {
      setDetailMsg(toErrorMessage(err, "記事の取得に失敗しました。"));
      setDetailItem(null);
      setDetailSource(null);
    } finally {
      setLoadingDetail(false);
    }
  }

  function closeModal() {
    setModalOpen(false);
  }

  if (loading) {
    return (
      <main className="min-h-screen bg-[#eef2f8] px-4 py-8 md:px-8 md:py-10">
        <div className="mx-auto max-w-[1280px] rounded-[20px] border border-[#9eb3d6] bg-white px-8 py-16 text-center text-lg font-bold text-[#355184] shadow-[0_8px_24px_rgba(27,52,120,0.08)]">
          読み込み中です...
        </div>
      </main>
    );
  }

  if (msg) {
    return (
      <main className="min-h-screen bg-[#eef2f8] px-4 py-8 md:px-8 md:py-10">
        <div className="mx-auto max-w-[1280px] rounded-[20px] border border-[#9eb3d6] bg-white px-8 py-16 text-center text-lg font-bold text-[#8a4b3a] shadow-[0_8px_24px_rgba(27,52,120,0.08)]">
          {msg}
        </div>
      </main>
    );
  }

  return (
    <>
      <main className="min-h-screen bg-[#eef2f8] px-3 py-6 md:px-8 md:py-10">
        <div className="mx-auto max-w-[1280px]">
          <section className="overflow-hidden rounded-[18px] border border-[#91a7d0] bg-white shadow-[0_10px_28px_rgba(27,52,120,0.08)]">
            <div className="border-b border-[#b8c8e2] bg-[#f0f4fb] px-3 py-3 md:px-5">
              <div className="flex flex-wrap items-center gap-1 md:gap-2">
                {kindTabs.map((tab) => (
                  <button
                    key={tab.kind}
                    type="button"
                    onClick={() => setActiveKind(tab.kind)}
                    className={[
                      "rounded-t-md border border-b-0 px-4 py-3 text-lg font-bold transition md:px-6",
                      activeKind === tab.kind
                        ? "border-[#91a7d0] bg-white text-[#16326e]"
                        : "border-transparent bg-transparent text-[#2a4fa3] hover:bg-[#e7eef9]",
                    ].join(" ")}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>
            </div>

            <div className="px-4 py-5 md:px-6 md:py-6">
              <div className="mb-4 flex flex-wrap items-center justify-between gap-3 text-[#5b6f98]">
                <div>
                  <p className="text-sm font-black tracking-[0.22em] text-[#4668ad] uppercase">
                    topics update
                  </p>
                  <p className="mt-1 text-[15px] font-semibold">
                    {updatedAt
                      ? `${formatDisplayDateTime(updatedAt)} 更新`
                      : "更新データなし"}
                  </p>
                </div>

                <div className="text-sm font-semibold text-[#7084ad]">
                  {kindDescLabel(activeKind)}
                </div>
              </div>

              <div className="grid gap-6 lg:grid-cols-[1.55fr_0.75fr]">
                <section>
                  <div className="rounded-[12px] border border-[#d7e0ef] bg-white px-4 py-2">
                    {subItems.length === 0 ? (
                      <div className="px-1 py-10 text-center text-base font-bold text-[#526482]">
                        このカテゴリのデータはまだありません。
                      </div>
                    ) : (
                      subItems.map((item) => (
                        <HeadlineRow
                          key={item.id}
                          item={item}
                          sourceName={sourceNameById(sources, item.source_id)}
                          onOpen={openModal}
                        />
                      ))
                    )}
                  </div>

                  <div className="flex flex-wrap gap-6 px-1 pt-4 text-[18px] font-bold text-[#2a4fa3]">
                    <Link
                      to={`/items?kind=${activeKind}`}
                      className="hover:underline"
                    >
                      トピックス一覧
                    </Link>
                  </div>
                </section>

                <aside>
                  {featured ? (
                    <button
                      type="button"
                      onClick={() => openModal(featured.id)}
                      className="flex w-full flex-col rounded-[12px] border border-[#d7e0ef] bg-white text-left shadow-[0_8px_20px_rgba(27,52,120,0.06)] transition hover:-translate-y-0.5"
                    >
                      <div className="overflow-hidden border-b border-[#d7e0ef] bg-[#eef3fb]">
                        <img
                          src={cardImage(featured.image_url)}
                          alt={featured.title}
                          className="h-[190px] w-full object-cover"
                        />
                      </div>

                      <div className="px-4 py-4">
                        <span className="rounded-full border border-[#a9bbdc] bg-[#eef3fb] px-3 py-1 text-[11px] font-black tracking-[0.2em] text-[#2a4fa3]">
                          {kindBadgeLabel(featured.kind)}
                        </span>

                        <h2 className="mt-3 text-[34px] font-black leading-tight text-[#16326e] lg:text-[26px]">
                          {featured.title}
                        </h2>

                        <p className="mt-3 text-sm font-medium leading-7 text-[#4d5c7c]">
                          {previewText(featured, 120)}
                        </p>

                        <div className="mt-4 space-y-1 text-sm font-semibold text-[#6b7ea6]">
                          <p>{formatDisplayDateTime(featured.published_at)}</p>
                          <p>{sourceNameById(sources, featured.source_id)}</p>
                        </div>
                      </div>
                    </button>
                  ) : null}
                </aside>
              </div>
            </div>
          </section>
        </div>
      </main>

      <DetailModal
        open={modalOpen}
        item={detailItem}
        source={detailSource}
        loading={loadingDetail}
        msg={detailMsg}
        onClose={closeModal}
      />
    </>
  );
}
