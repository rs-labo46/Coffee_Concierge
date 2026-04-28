export type Score = 1 | 2 | 3 | 4 | 5;

export type Roast = "light" | "medium" | "dark";

export type Method = "drip" | "espresso" | "milk" | "iced";

export type Mood = "morning" | "work" | "relax" | "night";

export type Scene = "work" | "break" | "after_meal" | "relax";

export type TempPref = "hot" | "ice";

export type SessionStatus = "active" | "closed";

export type TurnRole = "user" | "assistant" | "system";

export type TurnKind = "message" | "followup" | "notice";

export type ItemKind = "news" | "recipe" | "deal" | "shop";

export type SearchSession = {
  id: number;
  user_id: number | null;
  title: string;
  status: SessionStatus;
  guest_expires_at: string | null;
  created_at: string;
  updated_at: string;
};

export type Turn = {
  id: number;
  session_id: number;
  role: TurnRole;
  kind: TurnKind;
  body: string;
  created_at: string;
};

export type Pref = {
  id: number;
  session_id: number;
  flavor: number;
  acidity: number;
  bitterness: number;
  body: number;
  aroma: number;
  mood: Mood;
  method: Method;
  scene: Scene;
  temp_pref: TempPref;
  excludes: string[];
  note: string;
  created_at: string;
  updated_at: string;
};

export type Bean = {
  id: number;
  name: string;
  roast: Roast;
  origin: string;
  flavor: number;
  acidity: number;
  bitterness: number;
  body: number;
  aroma: number;
  desc: string;
  buy_url: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type Recipe = {
  id: number;
  bean_id: number;
  name: string;
  method: Method;
  temp_pref: TempPref;
  grind: string;
  ratio: string;
  temp: number;
  time_sec: number;
  steps: string[];
  desc: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type Source = {
  id: number;
  name: string;
  site_url: string;
  created_at: string;
};

export type ConciergeItem = {
  id: number;
  title: string;
  summary: string;
  url: string;
  image_url: string;
  kind: ItemKind;
  source_id: number;
  source?: Source;
  published_at: string;
  created_at: string;
};

export type Suggestion = {
  id: number;
  session_id: number;
  bean_id: number;
  bean?: Bean;
  recipe_id: number | null;
  recipe?: Recipe | null;
  item_id: number | null;
  item?: ConciergeItem | null;
  score: number;
  reason: string;
  rank: number;
  created_at: string;
};

export type SearchResult = {
  suggestions: Suggestion[];
  beans: Bean[];
  recipes: Recipe[];
  items: ConciergeItem[];
  followups: string[];
};

export type SetPrefInput = {
  flavor: Score;
  acidity: Score;
  bitterness: Score;
  body: Score;
  aroma: Score;
  mood: Mood;
  method: Method;
  scene: Scene;
  temp_pref: TempPref;
  excludes: string[];
  note: string;
};

export type PatchPrefInput = {
  flavor?: Score;
  acidity?: Score;
  bitterness?: Score;
  body?: Score;
  aroma?: Score;
  mood?: Mood;
  method?: Method;
  scene?: Scene;
  temp_pref?: TempPref;
  excludes?: string[];
  note?: string;
};

export type StartSessionResponse = {
  session: SearchSession;
  session_key?: string;
};

export type SetPrefResponse = {
  pref: Pref;
  result: SearchResult;
};

export type PatchPrefResponse = {
  pref: Pref;
  result: SearchResult;
};

export type SaveSuggestionInput = {
  session_id: number;
  suggestion_id: number;
};

export type SavedSuggestion = {
  id: number;
  user_id: number;
  session_id: number;
  suggestion_id: number;
  suggestion: Suggestion;
  created_at: string;
};

export type SaveSuggestionResponse = {
  saved: SavedSuggestion;
};

export const defaultPrefInput: SetPrefInput = {
  flavor: 3,
  acidity: 3,
  bitterness: 3,
  body: 3,
  aroma: 3,
  mood: "morning",
  method: "drip",
  scene: "break",
  temp_pref: "hot",
  excludes: [],
  note: "",
};

export function roastLabel(value: Roast): string {
  const labels: Record<Roast, string> = {
    light: "浅煎り",
    medium: "中煎り",
    dark: "深煎り",
  };

  return labels[value];
}

export function methodLabel(value: Method): string {
  const labels: Record<Method, string> = {
    drip: "ドリップ",
    espresso: "エスプレッソ",
    milk: "ミルク向け",
    iced: "アイス",
  };

  return labels[value];
}

export function moodLabel(value: Mood): string {
  const labels: Record<Mood, string> = {
    morning: "朝",
    work: "仕事",
    relax: "リラックス",
    night: "夜",
  };

  return labels[value];
}

export function sceneLabel(value: Scene): string {
  const labels: Record<Scene, string> = {
    work: "作業中",
    break: "休憩",
    after_meal: "食後",
    relax: "くつろぎ",
  };

  return labels[value];
}

export function tempPrefLabel(value: TempPref): string {
  const labels: Record<TempPref, string> = {
    hot: "ホット",
    ice: "アイス",
  };

  return labels[value];
}

export function itemKindLabel(value: ItemKind): string {
  const labels: Record<ItemKind, string> = {
    news: "記事",
    recipe: "レシピ記事",
    deal: "セール",
    shop: "販売導線",
  };

  return labels[value];
}
