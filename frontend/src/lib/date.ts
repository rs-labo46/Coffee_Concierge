function toTokyoParts(value: string): Map<string, string> | null {
  if (!value) {
    return null;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return null;
  }

  const parts = new Intl.DateTimeFormat("ja-JP", {
    timeZone: "Asia/Tokyo",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).formatToParts(date);

  return new Map(parts.map((part) => [part.type, part.value]));
}

export function formatTokyoDateTimeInput(value: string): string {
  const map = toTokyoParts(value);
  if (!map) {
    return value;
  }

  return `${map.get("year")}-${map.get("month")}-${map.get("day")}T${map.get("hour")}:${map.get("minute")}`;
}

export function formatDisplayDate(value: string): string {
  const map = toTokyoParts(value);
  if (!map) {
    return "";
  }

  return `${map.get("year")}/${map.get("month")}/${map.get("day")}`;
}

export function formatDisplayDateTime(value: string): string {
  const map = toTokyoParts(value);
  if (!map) {
    return value;
  }

  return `${map.get("year")}-${map.get("month")}-${map.get("day")} ${map.get("hour")}:${map.get("minute")}`;
}
