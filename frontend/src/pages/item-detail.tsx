import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, toErrorMessage } from "../lib/api";
import { formatDisplayDateTime } from "../lib/date";
import {
  bodyParagraphs,
  cardImage,
  kindBadgeLabel,
  kindDescLabel,
  type Item,
  type ItemDetailRes,
  type Source,
} from "../lib/item";

export function ItemDetailPage() {
  const params = useParams();
  const [item, setItem] = useState<Item | null>(null);
  const [source, setSource] = useState<Source | null>(null);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState("");

  useEffect(() => {
    async function run() {
      setLoading(true);
      setMsg("");

      try {
        const id = Number(params.id || "0");
        if (!id) {
          setMsg("記事IDが不正です。");
          return;
        }

        const res = await api<ItemDetailRes>(`/items/${id}`, {
          method: "GET",
        });

        if (!res) {
          setMsg("記事の取得に失敗しました。");
          return;
        }

        setItem(res.item);
        setSource(res.source);
      } catch (err: unknown) {
        setMsg(toErrorMessage(err, "記事の取得に失敗しました。"));
      } finally {
        setLoading(false);
      }
    }

    void run();
  }, [params.id]);

  const paragraphs = useMemo(() => {
    if (!item) {
      return [];
    }

    return bodyParagraphs(item);
  }, [item]);

  if (loading) {
    return (
      <main className="min-h-screen bg-[#eef2f8] px-4 py-8 md:px-8 md:py-10">
        <div className="mx-auto max-w-[1080px] rounded-[20px] border border-[#b9c8e2] bg-white px-8 py-16 text-center text-lg font-bold text-[#355184] shadow-[0_8px_24px_rgba(27,52,120,0.08)]">
          読み込み中です...
        </div>
      </main>
    );
  }

  if (msg || !item) {
    return (
      <main className="min-h-screen bg-[#eef2f8] px-4 py-8 md:px-8 md:py-10">
        <div className="mx-auto max-w-[1080px] rounded-[20px] border border-[#b9c8e2] bg-white px-8 py-16 text-center shadow-[0_8px_24px_rgba(27,52,120,0.08)]">
          <p className="text-lg font-bold text-[#8a4b3a]">
            {msg || "記事が見つかりませんでした。"}
          </p>
          <div className="mt-6">
            <Link
              to="/"
              className="inline-flex min-h-11 items-center justify-center rounded-full border border-[#2a4fa3] px-4 py-2 text-sm font-bold text-[#2a4fa3] transition hover:bg-[#eef3fb]"
            >
              トップへ戻る
            </Link>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#eef2f8] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto flex max-w-[1080px] flex-col gap-6">
        <div className="flex flex-wrap gap-3">
          <Link
            to="/"
            className="inline-flex min-h-11 items-center justify-center rounded-full border border-[#d0dbed] bg-white px-4 py-2 text-sm font-bold text-[#2a4fa3] transition hover:bg-[#f4f7fc]"
          >
            トップへ戻る
          </Link>
          <Link
            to={`/items?kind=${item.kind}`}
            className="inline-flex min-h-11 items-center justify-center rounded-full border border-[#d0dbed] bg-white px-4 py-2 text-sm font-bold text-[#2a4fa3] transition hover:bg-[#f4f7fc]"
          >
            同カテゴリ一覧
          </Link>
        </div>

        <article className="overflow-hidden rounded-[20px] border border-[#9eb3d6] bg-white shadow-[0_10px_28px_rgba(27,52,120,0.08)]">
          <div className="border-b border-[#c6d4ea] bg-[#f4f7fc] px-6 py-6 md:px-8">
            <div className="flex flex-wrap items-center gap-3 text-sm font-semibold text-[#5b6f98]">
              <span className="rounded-full border border-[#a9bbdc] bg-[#eef3fb] px-3 py-1 text-[11px] font-black tracking-[0.22em] text-[#2a4fa3]">
                {kindBadgeLabel(item.kind)}
              </span>
              <span>{kindDescLabel(item.kind)}</span>
            </div>

            <h1 className="mt-4 text-3xl font-black leading-tight text-[#16326e] md:text-4xl">
              {item.title}
            </h1>

            <div className="mt-4 flex flex-wrap gap-x-6 gap-y-2 text-sm font-semibold text-[#62749a]">
              <span>公開日時: {formatDisplayDateTime(item.published_at)}</span>
              <span>作成日時: {formatDisplayDateTime(item.created_at)}</span>
              <span>出典: {source?.name || "未設定"}</span>
            </div>
          </div>

          <div className="px-6 py-6 md:px-8">
            <div className="overflow-hidden rounded-[16px] border border-[#d7e0ef] bg-[#eef3fb]">
              <img
                src={cardImage(item.image_url)}
                alt={item.title}
                className="h-[260px] w-full object-cover md:h-[360px]"
              />
            </div>

            {item.summary ? (
              <div className="mt-6 rounded-[16px] border border-[#d7e0ef] bg-[#f7faff] px-5 py-4">
                <p className="text-sm font-black tracking-[0.2em] text-[#4668ad] uppercase">
                  概要
                </p>
                <p className="mt-3 text-base font-medium leading-8 text-[#30466d]">
                  {item.summary}
                </p>
              </div>
            ) : null}

            <section className="mt-8 rounded-[16px] border border-[#d7e0ef] bg-white px-5 py-6 md:px-7">
              <p className="mb-5 text-sm font-black tracking-[0.2em] text-[#4668ad] uppercase">
                内容
              </p>

              <div className="space-y-5">
                {paragraphs.map((paragraph, index) => (
                  <p
                    key={`${item.id}-${index}`}
                    className="text-base font-medium leading-9 text-[#30466d]"
                  >
                    {paragraph}
                  </p>
                ))}
              </div>
            </section>
          </div>
        </article>
      </div>
    </main>
  );
}
