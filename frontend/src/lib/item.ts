export type ItemKind = "news" | "recipe" | "deal" | "shop";

export type Item = {
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

export type Source = {
  id: number;
  name: string;
  site_url: string | null;
  created_at: string;
};

export type TopRes = {
  news: Item[];
  recipe: Item[];
  deal: Item[];
  shop: Item[];
};

export type ItemListRes = {
  items: Item[];
};

export type SourceListRes = {
  sources: Source[];
};

export type ItemDetailRes = {
  item: Item;
  source: Source;
};

export function isItemKind(value: string | null): value is ItemKind {
  return (
    value === "news" ||
    value === "recipe" ||
    value === "deal" ||
    value === "shop"
  );
}

export function kindBadgeLabel(kind: ItemKind): string {
  switch (kind) {
    case "news":
      return "NEWS";
    case "recipe":
      return "RECIPE";
    case "deal":
      return "DEAL";
    case "shop":
      return "SHOP";
    default:
      return "";
  }
}

export function kindTitleLabel(kind: ItemKind): string {
  switch (kind) {
    case "news":
      return "ニュース";
    case "recipe":
      return "レシピ";
    case "deal":
      return "セール";
    case "shop":
      return "店舗";
    default:
      return "";
  }
}

export function kindDescLabel(kind: ItemKind): string {
  switch (kind) {
    case "news":
      return "市場・焙煎・器具・業界トレンド";
    case "recipe":
      return "抽出手順・味づくり・再現性";
    case "deal":
      return "クーポン・割引・キャンペーン";
    case "shop":
      return "新店・注目店・利用シーン";
    default:
      return "";
  }
}

export function cardImage(url: string | null): string {
  if (url && url.trim() !== "") {
    return url;
  }

  return "https://images.unsplash.com/photo-1509042239860-f550ce710b93?auto=format&fit=crop&w=1200&q=80";
}

export function hasRef(url: string | null): boolean {
  return !!url && url.trim() !== "";
}

export function isFreshItem(value: string): boolean {
  const t = new Date(value).getTime();
  if (Number.isNaN(t)) {
    return false;
  }

  return Date.now() - t <= 1000 * 60 * 60 * 48;
}

export function itemBodyText(item: Item): string {
  if (item.body && item.body.trim() !== "") {
    return item.body.trim();
  }

  if (item.summary && item.summary.trim() !== "") {
    return item.summary.trim();
  }

  return "詳細本文はまだ登録されていません。";
}

export function previewText(item: Item, max = 160): string {
  const raw = itemBodyText(item).replace(/\s+/g, " ").trim();

  if (raw.length <= max) {
    return raw;
  }

  return `${raw.slice(0, max)}…`;
}

export function halfPreviewText(item: Item): string {
  const raw = itemBodyText(item).replace(/\s+/g, " ").trim();
  const max = Math.min(420, Math.max(180, Math.floor(raw.length / 2)));

  if (raw.length <= max) {
    return raw;
  }

  return `${raw.slice(0, max)}…`;
}

export function bodyParagraphs(item: Item): string[] {
  const raw = itemBodyText(item);

  const parts = raw
    .split(/\n{2,}/)
    .map((part) => part.trim())
    .filter((part) => part !== "");

  if (parts.length > 0) {
    return parts;
  }

  return [raw];
}
