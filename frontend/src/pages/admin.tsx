import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { api, toErrorMessage } from "../lib/api";
import { formatDisplayDateTime, formatTokyoDateTimeInput } from "../lib/date";

type ItemKind = "news" | "recipe" | "deal" | "shop";

type Source = {
  id: number;
  name: string;
  site_url: string | null;
  created_at: string;
};

type Item = {
  id: number;
  title: string;
  summary: string | null;
  body: string | null;
  url: string | null;
  image_url: string | null;
  kind: ItemKind;
  source_id: number;
  published_at: string;
  created_at: string;
};

type SourceListRes = {
  sources: Source[];
};

type SourceRes = {
  source: Source;
};

type ItemRes = {
  item: Item;
};

type SourceForm = {
  name: string;
  site_url: string;
};

type ItemForm = {
  title: string;
  summary: string;
  body: string;
  url: string;
  image_url: string;
  kind: ItemKind;
  source_id: string;
  published_at: string;
};

function newSourceForm(): SourceForm {
  return {
    name: "",
    site_url: "",
  };
}

function newItemForm(): ItemForm {
  return {
    title: "",
    summary: "",
    body: "",
    url: "",
    image_url: "",
    kind: "news",
    source_id: "",
    published_at: "",
  };
}

function toNullableText(value: string): string | null {
  const trimmed = value.trim();

  if (!trimmed) {
    return null;
  }

  return trimmed;
}

function toApiDateTime(value: string): string {
  if (!value) {
    return "";
  }

  return `${value}:00+09:00`;
}

function SectionTitle({ title, sub }: { title: string; sub: string }) {
  return (
    <div className="mb-5 border-b border-[#eadfd5] pb-4">
      <p className="mb-2 text-xs font-black uppercase tracking-[0.28em] text-[#a1775b]">
        admin form
      </p>
      <h2 className="text-2xl font-black text-[#4e342e]">{title}</h2>
      <p className="mt-2 text-sm font-semibold leading-7 text-[#766b63]">
        {sub}
      </p>
    </div>
  );
}

function FieldLabel({
  htmlFor,
  label,
  required,
}: {
  htmlFor: string;
  label: string;
  required?: boolean;
}) {
  return (
    <label
      htmlFor={htmlFor}
      className="mb-2 block text-sm font-black tracking-[0.06em] text-[#5f4a40]"
    >
      {label}
      {required ? <span className="ml-1 text-[#a23c2e]">*</span> : null}
    </label>
  );
}

function NavPill({ to, label }: { to: string; label: string }) {
  return (
    <Link
      to={to}
      className="inline-flex min-h-11 items-center justify-center rounded-full border border-[#d9c7b8] bg-white px-5 py-2.5 text-sm font-bold text-[#5a3825] transition hover:bg-[#f5ece5]"
    >
      {label}
    </Link>
  );
}

function SubmitBtn({
  label,
  loadingLabel,
  loading,
  tone = "dark",
}: {
  label: string;
  loadingLabel: string;
  loading: boolean;
  tone?: "dark" | "mid";
}) {
  const cls =
    tone === "dark" ? "bg-[#5a3825] text-white" : "bg-[#8b5e3c] text-white";

  return (
    <button
      type="submit"
      disabled={loading}
      className={`inline-flex min-h-12 min-w-[220px] items-center justify-center rounded-2xl px-6 py-3 text-sm font-bold transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60 ${cls}`}
    >
      {loading ? loadingLabel : label}
    </button>
  );
}

export function AdminPage() {
  const [sources, setSources] = useState<Source[]>([]);
  const [loadingSources, setLoadingSources] = useState<boolean>(true);
  const [sourcesMsg, setSourcesMsg] = useState<string>("");
  const [sourceForm, setSourceForm] = useState<SourceForm>(newSourceForm());
  const [savingSource, setSavingSource] = useState<boolean>(false);
  const [sourceMsg, setSourceMsg] = useState<string>("");
  const [itemForm, setItemForm] = useState<ItemForm>(newItemForm());
  const [savingItem, setSavingItem] = useState<boolean>(false);
  const [itemMsg, setItemMsg] = useState<string>("");
  const [createdItem, setCreatedItem] = useState<Item | null>(null);

  const firstSourceId = useMemo<string>(() => {
    if (sources.length === 0) {
      return "";
    }

    return String(sources[0].id);
  }, [sources]);

  async function loadSources(): Promise<void> {
    setLoadingSources(true);
    setSourcesMsg("");

    try {
      const res = await api<SourceListRes>("/sources", {
        method: "GET",
      });

      if (!res) {
        setSources([]);
        setSourcesMsg("source一覧が空です。");
        return;
      }

      setSources(res.sources);

      setItemForm((prev) => {
        if (prev.source_id) {
          return prev;
        }

        if (res.sources.length === 0) {
          return prev;
        }

        return {
          ...prev,
          source_id: String(res.sources[0].id),
        };
      });
    } catch (err: unknown) {
      setSources([]);
      setSourcesMsg(toErrorMessage(err, "source一覧の取得に失敗しました。"));
    } finally {
      setLoadingSources(false);
    }
  }

  useEffect(() => {
    void loadSources();
  }, []);

  useEffect(() => {
    if (!itemForm.source_id && firstSourceId) {
      setItemForm((prev) => ({
        ...prev,
        source_id: firstSourceId,
      }));
    }
  }, [firstSourceId, itemForm.source_id]);

  function onChangeSourceForm(e: React.ChangeEvent<HTMLInputElement>): void {
    const { name, value } = e.target;

    setSourceForm((prev) => ({
      ...prev,
      [name]: value,
    }));
  }

  function onChangeItemForm(
    e:
      | React.ChangeEvent<HTMLInputElement>
      | React.ChangeEvent<HTMLTextAreaElement>
      | React.ChangeEvent<HTMLSelectElement>,
  ): void {
    const { name, value } = e.target;

    setItemForm((prev) => ({
      ...prev,
      [name]: value,
    }));
  }

  async function onSubmitSource(
    e: React.FormEvent<HTMLFormElement>,
  ): Promise<void> {
    e.preventDefault();

    setSourceMsg("");
    setSavingSource(true);

    try {
      const body = {
        name: sourceForm.name.trim(),
        site_url: toNullableText(sourceForm.site_url),
      };

      const res = await api<SourceRes>("/sources", {
        method: "POST",
        auth: true,
        body,
      });

      if (res?.source) {
        setSourceMsg(`source「${res.source.name}」を作成しました。`);
      } else {
        setSourceMsg("sourceを作成しました。");
      }

      setSourceForm(newSourceForm());
      await loadSources();
    } catch (err: unknown) {
      setSourceMsg(toErrorMessage(err, "sourceの作成に失敗しました。"));
    } finally {
      setSavingSource(false);
    }
  }

  async function onSubmitItem(
    e: React.FormEvent<HTMLFormElement>,
  ): Promise<void> {
    e.preventDefault();

    setItemMsg("");
    setCreatedItem(null);

    if (sources.length === 0) {
      setItemMsg("先にsourceを1件以上登録してください。");
      return;
    }

    const sourceId = Number(itemForm.source_id);

    if (!sourceId) {
      setItemMsg("source を選択してください。");
      return;
    }

    if (!itemForm.published_at) {
      setItemMsg("公開日時を入力してください。");
      return;
    }

    setSavingItem(true);

    try {
      const body = {
        title: itemForm.title.trim(),
        summary: toNullableText(itemForm.summary),
        body: toNullableText(itemForm.body),
        url: toNullableText(itemForm.url),
        image_url: toNullableText(itemForm.image_url),
        kind: itemForm.kind,
        source_id: sourceId,
        published_at: toApiDateTime(itemForm.published_at),
      };

      const res = await api<ItemRes>("/items", {
        method: "POST",
        auth: true,
        body,
      });

      if (res?.item) {
        setCreatedItem(res.item);
        setItemMsg(`item「${res.item.title}」を作成しました。`);
      } else {
        setItemMsg("itemを作成しました。");
      }

      setItemForm((prev) => ({
        ...newItemForm(),
        source_id: prev.source_id || firstSourceId,
        kind: "news",
      }));
    } catch (err: unknown) {
      setItemMsg(toErrorMessage(err, "itemの作成に失敗しました。"));
    } finally {
      setSavingItem(false);
    }
  }

  return (
    <main className="min-h-[calc(100vh-120px)] bg-[#f6f1eb] px-4 py-8 text-stone-800 md:px-8 md:py-10">
      <div className="mx-auto max-w-[1280px]">
        <header className="mb-8 overflow-hidden rounded-[36px] border border-[#e6d9ce] bg-white shadow-[0_10px_28px_rgba(110,78,56,0.08)]">
          <div className="border-b border-[#e9ddd3] bg-gradient-to-r from-[#5a3825] via-[#7a5239] to-[#a67c52] px-6 py-8 text-white md:px-8 md:py-9">
            <p className="text-sm font-black uppercase tracking-[0.26em] text-white/80">
              Coffee SPA Admin
            </p>
            <h1 className="mt-2 text-3xl font-black md:text-4xl">管理画面</h1>
          </div>

          <div className="px-6 py-5 md:px-8">
            <div className="rounded-[28px] border border-[#eadfd5] bg-[#fcf8f4] px-4 py-4 md:px-6">
              <p className="mb-3 text-center text-xs font-black uppercase tracking-[0.28em] text-[#a1775b]">
                quick action
              </p>

              <div className="flex flex-wrap items-center justify-center gap-3">
                <NavPill to="/" label="公開トップへ" />
              </div>
            </div>
          </div>
        </header>

        <div className="grid gap-6 lg:grid-cols-[0.92fr_1.08fr]">
          <section className="rounded-[32px] border border-[#e6d9ce] bg-white p-6 shadow-[0_10px_28px_rgba(110,78,56,0.06)] md:p-7">
            <SectionTitle title="Source 登録" sub="出典元を登録します。" />

            <form
              onSubmit={(e) => void onSubmitSource(e)}
              className="space-y-5"
            >
              <div>
                <FieldLabel htmlFor="source-name" label="Source名" required />
                <input
                  id="source-name"
                  name="name"
                  type="text"
                  value={sourceForm.name}
                  onChange={onChangeSourceForm}
                  placeholder="例: Coffee Daily"
                  className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                />
              </div>

              <div>
                <FieldLabel htmlFor="source-site-url" label="サイトURL" />
                <input
                  id="source-site-url"
                  name="site_url"
                  type="url"
                  value={sourceForm.site_url}
                  onChange={onChangeSourceForm}
                  placeholder="https://example.com"
                  className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                />
              </div>

              <div className="flex justify-center pt-1">
                <SubmitBtn
                  label="Sourceを作成"
                  loadingLabel="作成中..."
                  loading={savingSource}
                  tone="dark"
                />
              </div>
            </form>

            {sourceMsg ? (
              <p className="mt-5 rounded-[22px] border border-[#ead9cd] bg-[#f8efe7] px-4 py-3 text-sm font-bold text-[#5a3825]">
                {sourceMsg}
              </p>
            ) : null}
          </section>

          <section className="rounded-[32px] border border-[#e6d9ce] bg-white p-6 shadow-[0_10px_28px_rgba(110,78,56,0.06)] md:p-7">
            <SectionTitle
              title="Item 登録"
              sub="ニュース、レシピ、セール、店舗情報を登録します。"
            />

            <form onSubmit={(e) => void onSubmitItem(e)} className="space-y-5">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="md:col-span-2">
                  <FieldLabel htmlFor="item-title" label="タイトル" required />
                  <input
                    id="item-title"
                    name="title"
                    type="text"
                    value={itemForm.title}
                    onChange={onChangeItemForm}
                    placeholder="例: 今週の新作コーヒー豆"
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                  />
                </div>

                <div className="md:col-span-2">
                  <FieldLabel htmlFor="item-summary" label="要約" />
                  <textarea
                    id="item-summary"
                    name="summary"
                    value={itemForm.summary}
                    onChange={onChangeItemForm}
                    placeholder="短い説明"
                    rows={4}
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                  />
                </div>

                <div className="md:col-span-2">
                  <FieldLabel htmlFor="item-body" label="本文" />
                  <textarea
                    id="item-body"
                    name="body"
                    value={itemForm.body}
                    onChange={onChangeItemForm}
                    placeholder="詳細ページで表示する本文。段落を分けたい場合は空行を入れてください。"
                    rows={8}
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold leading-7 text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                  />
                </div>

                <div>
                  <FieldLabel htmlFor="item-url" label="遷移先URL" />
                  <input
                    id="item-url"
                    name="url"
                    type="url"
                    value={itemForm.url}
                    onChange={onChangeItemForm}
                    placeholder="https://example.com/article"
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                  />
                </div>

                <div>
                  <FieldLabel htmlFor="item-image-url" label="画像URL" />
                  <input
                    id="item-image-url"
                    name="image_url"
                    type="url"
                    value={itemForm.image_url}
                    onChange={onChangeItemForm}
                    placeholder="https://example.com/image.jpg"
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                  />
                </div>

                <div>
                  <FieldLabel htmlFor="item-kind" label="種別" required />
                  <select
                    id="item-kind"
                    name="kind"
                    value={itemForm.kind}
                    onChange={onChangeItemForm}
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                  >
                    <option value="news">news</option>
                    <option value="recipe">recipe</option>
                    <option value="deal">deal</option>
                    <option value="shop">shop</option>
                  </select>
                </div>

                <div>
                  <FieldLabel
                    htmlFor="item-source-id"
                    label="Source"
                    required
                  />
                  <select
                    id="item-source-id"
                    name="source_id"
                    value={itemForm.source_id}
                    onChange={onChangeItemForm}
                    disabled={sources.length === 0}
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca] disabled:cursor-not-allowed disabled:bg-stone-100"
                  >
                    {sources.length === 0 ? (
                      <option value="">sourceを先に登録</option>
                    ) : (
                      sources.map((source) => (
                        <option key={source.id} value={String(source.id)}>
                          {source.name} / id:{source.id}
                        </option>
                      ))
                    )}
                  </select>
                </div>

                <div className="md:col-span-2">
                  <FieldLabel
                    htmlFor="item-published-at"
                    label="公開日時"
                    required
                  />
                  <input
                    id="item-published-at"
                    name="published_at"
                    type="datetime-local"
                    value={formatTokyoDateTimeInput(itemForm.published_at)}
                    onChange={onChangeItemForm}
                    className="w-full rounded-2xl border border-[#d8c8bc] bg-[#fffdfb] px-4 py-3.5 text-sm font-semibold text-[#4e342e] outline-none transition focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#ead8ca]"
                  />
                </div>
              </div>

              <div className="flex justify-center pt-1">
                <SubmitBtn
                  label="Itemを作成"
                  loadingLabel="作成中..."
                  loading={savingItem || sources.length === 0}
                  tone="mid"
                />
              </div>
            </form>

            {itemMsg ? (
              <p className="mt-5 rounded-[22px] border border-[#ead9cd] bg-[#f8efe7] px-4 py-3 text-sm font-bold text-[#5a3825]">
                {itemMsg}
              </p>
            ) : null}

            {createdItem ? (
              <div className="mt-5 rounded-[28px] border border-[#e3d6cc] bg-[#fffaf6] p-5">
                <p className="text-sm font-black tracking-[0.18em] text-[#a1775b] uppercase">
                  latest item
                </p>

                <dl className="mt-4 grid gap-3 text-sm text-stone-700">
                  <div className="grid grid-cols-[110px_1fr] gap-3">
                    <dt className="font-bold text-[#5f4a40]">ID</dt>
                    <dd>{createdItem.id}</dd>
                  </div>

                  <div className="grid grid-cols-[110px_1fr] gap-3">
                    <dt className="font-bold text-[#5f4a40]">タイトル</dt>
                    <dd>{createdItem.title}</dd>
                  </div>

                  <div className="grid grid-cols-[110px_1fr] gap-3">
                    <dt className="font-bold text-[#5f4a40]">種別</dt>
                    <dd>{createdItem.kind}</dd>
                  </div>

                  <div className="grid grid-cols-[110px_1fr] gap-3">
                    <dt className="font-bold text-[#5f4a40]">Source ID</dt>
                    <dd>{createdItem.source_id}</dd>
                  </div>

                  <div className="grid grid-cols-[110px_1fr] gap-3">
                    <dt className="font-bold text-[#5f4a40]">公開日時</dt>
                    <dd>{createdItem.published_at}</dd>
                  </div>
                </dl>
              </div>
            ) : null}
          </section>
        </div>

        <section className="mt-6 rounded-[32px] border border-[#e6d9ce] bg-white p-6 shadow-[0_10px_28px_rgba(110,78,56,0.06)] md:p-7">
          <SectionTitle
            title="Source 一覧"
            sub="登録済みの出典元を確認します。"
          />

          {loadingSources ? (
            <p className="text-sm font-semibold text-stone-600">loading...</p>
          ) : null}

          {!loadingSources && sourcesMsg ? (
            <p className="rounded-[22px] border border-[#ead9cd] bg-[#f8efe7] px-4 py-3 text-sm font-bold text-[#5a3825]">
              {sourcesMsg}
            </p>
          ) : null}

          {!loadingSources && !sourcesMsg && sources.length === 0 ? (
            <p className="text-sm font-semibold text-stone-600">
              まだありません。
            </p>
          ) : null}

          {!loadingSources && sources.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full border-separate border-spacing-y-3 text-sm">
                <thead>
                  <tr className="text-left text-[#7b6d63]">
                    <th className="px-3 py-2 font-black">id</th>
                    <th className="px-3 py-2 font-black">name</th>
                    <th className="px-3 py-2 font-black">site_url</th>
                    <th className="px-3 py-2 font-black">created_at</th>
                  </tr>
                </thead>

                <tbody>
                  {sources.map((source) => (
                    <tr
                      key={source.id}
                      className="rounded-2xl bg-[#fffaf6] text-stone-800"
                    >
                      <td className="px-3 py-3 align-top font-semibold">
                        {source.id}
                      </td>
                      <td className="px-3 py-3 align-top font-bold">
                        {source.name}
                      </td>
                      <td className="px-3 py-3 align-top">
                        {source.site_url ? (
                          <a
                            href={source.site_url}
                            target="_blank"
                            rel="noreferrer"
                            className="break-all font-semibold text-[#7a5239] underline"
                          >
                            {source.site_url}
                          </a>
                        ) : (
                          <span className="text-stone-400">-</span>
                        )}
                      </td>
                      <td className="px-3 py-3 align-top font-semibold">
                        {formatDisplayDateTime(source.created_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}
        </section>
      </div>
    </main>
  );
}
