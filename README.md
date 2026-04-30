# Coffee Concierge — コーヒーAIコンシェルジュ

Coffee Conciergeは、ユーザーの味覚条件・気分・飲み方・シーンから、コーヒー豆をランキング形式で提案するAIコンシェルジュアプリです。

FrontendはVercel、BackendはRender、DatabaseはRender PostgreSQL、Cache/Rate LimitはRender Redis、AI補助にはGemini APIを使用します。

このREADMEは、「何を設定し、どの順番で確認すれば本番環境まで再現できるか」を迷わず進められることを目的にしています。

---

## 1. 全体構成

| 領域               | 技術                              | 配置先                               |
| ------------------ | --------------------------------- | ------------------------------------ |
| Frontend           | React / Vite / TypeScript         | Vercel                               |
| Backend            | Go / Echo v4 / Clean Architecture | Render Web Service                   |
| Database           | PostgreSQL                        | Render PostgreSQL                    |
| Cache / Rate Limit | Redis                             | Render Redis / Key Value             |
| AI                 | Gemini API                        | Render Backendから呼び出し           |
| CI/CD              | GitHub Actions                    | `.github/workflows/ci-cd.yml`        |
| Production Deploy  | Deploy Hook                       | GitHub Actions → Render / Vercel     |
| Seed               | 手動GitHub Actions                | `workflow_dispatch`の`run_seed=true` |

---

## 2. 前提

このプロジェクトでは、FrontendとBackendを別サービスに分けています。

本番環境

```text
Frontend: https://coffee-concierge.vercel.app
Backend : https://coffee-concierge-api.onrender.com
```

ただし、認証系APIはCookie / CSRFの都合により、Frontend側のVercelドメインを経由します。

```text
/auth/* → Vercel Rewrite → Render Backend /auth/*
```

このため、Frontendには以下の2種類のAPI URLがあります。

| 環境変数             | 役割                                      |
| -------------------- | ----------------------------------------- |
| `VITE_API_BASE_URL`  | 通常API用。Render Backendを直接参照する   |
| `VITE_AUTH_BASE_URL` | 認証API用。Vercel FrontendのURLを指定する |

---

## 3. ブランチ運用

本番環境は`main`ブランチを基準に動きます。

| ブランチ                        | 役割                 |
| ------------------------------- | -------------------- |
| `develop`                       | 開発・統合ブランチ   |
| `main`                          | 本番デプロイブランチ |
| `feature/*`, `fix/*`, `chore/*` | 作業ブランチ         |

本番反映は必ず以下の流れで行います。

```text
作業ブランチ → develop → main
```

`fix/*`や`chore/*`を直接`main`にmergeしないでください。

本番反映コマンド例:

```bash
git checkout develop
git pull origin develop

# 必要に応じて作業ブランチをdevelopへmerge済みにする

git checkout main
git pull origin main
git merge develop
git push origin main
```

`main`にpushされると、GitHub ActionsのCI/CDが走り、Render / VercelのDeploy Hookが実行されます。

---

## 4. 絶対に先に確認すること

### 4-1. `.env`をGitHubに上げない

以下を実行し、`.env`がGit管理されていないことを確認します。

```bash
git status --short
git ls-files .env frontend/.env backend/.env
```

何も表示されなければ大丈夫です。

表示された場合は、以下を実行します。

```bash
git rm --cached .env frontend/.env backend/.env 2>/dev/null || true
git add .gitignore
git commit -m "chore: stop tracking env files"
```

すでにGitHubに上げたAPI Key、DB URL、JWT Secretは、削除ではなく再発行してください。

---

## 5. ローカル確認

### 5-1. Backend

```bash
cd backend
go test ./...
go build ./cmd/seed
cd ..
rm backend/seed 2>/dev/null || true
```

`go build ./cmd/seed`を実行すると、`backend/seed`という実行ファイルが生成されることがあります。これはGitに含めないでください。

### 5-2. Frontend

```bash
cd frontend
npm ci
npm run lint
npm run build
cd ..
```

---

## 6. Vercel設定

### 6-1. Project設定

Vercelで`Coffee_Concierge`をImportし、以下のように設定します。

| 項目             | 値              |
| ---------------- | --------------- |
| Framework Preset | Vite            |
| Root Directory   | `frontend`      |
| Install Command  | `npm ci`        |
| Build Command    | `npm run build` |
| Output Directory | `dist`          |

`Root Directory`は必ず`frontend`にしてください。ルートのままだと`package.json`を見つけられず失敗します。

### 6-2. Vercel Environment Variables

Vercel Dashboard → Project → Settings → Environment Variablesに以下を設定します。

| Key                  | Productionの値                              | 説明      |
| -------------------- | ------------------------------------------- | --------- |
| `VITE_API_BASE_URL`  | `https://coffee-concierge-api.onrender.com` | 通常API用 |
| `VITE_AUTH_BASE_URL` | `https://coffee-concierge.vercel.app`       | 認証API用 |

`VITE_AUTH_BASE_URL`は発行されるものではありません。Vercel本番URLを手入力します。

環境変数を追加・変更した場合は、必ずVercelでRedeployしてください。Viteの`VITE_`環境変数はbuild時に埋め込まれるためです。

---

## 7. Vercel Rewrite

`frontend/vercel.json`では、SPA直アクセス対応と認証API proxyを設定しています。

```json
{
  "rewrites": [
    {
      "source": "/auth/:path*",
      "destination": "https://coffee-concierge-api.onrender.com/auth/:path*"
    },
    {
      "source": "/(.*)",
      "destination": "/index.html"
    }
  ]
}
```

これにより、以下が成立します。

| URL             | 実際の意味                       |
| --------------- | -------------------------------- |
| `/login`        | SPAのlogin画面                   |
| `/concierge`    | SPAのconcierge画面               |
| `/auth/login`   | Vercel経由でRender Backendへ転送 |
| `/auth/logout`  | Vercel経由でRender Backendへ転送 |
| `/auth/refresh` | Vercel経由でRender Backendへ転送 |
| `/auth/csrf`    | Vercel経由でRender Backendへ転送 |

---

## 8. Render Backend設定

Render Web Service `coffee-concierge-api`にBackendをデプロイします。

BackendからPostgreSQLへ接続するときは、Render内部ネットワーク上のInternal Hostを使います。一方、GitHub Actionsからseedを流すときはRender外部から接続するため、External Hostを使います。この2つを混同しないでください。

### 8-1. Render Backend環境変数

| Key                 | 例                                    | 説明                                                                |
| ------------------- | ------------------------------------- | ------------------------------------------------------------------- |
| `GO_ENV`            | `prod`                                | 本番環境                                                            |
| `PORT`              | Render側が指定                        | Renderが渡すlisten port。通常はRenderの値を使う                     |
| `FE_URL`            | `https://coffee-concierge.vercel.app` | CORS許可元                                                          |
| `JWT_SECRET`        | 強いランダム文字列                    | JWT署名用。公開禁止                                                 |
| `COOKIE_SECURE`     | `true`                                | 本番Cookie用                                                        |
| `POSTGRES_USER`     | `coffee_user`                         | DBユーザー                                                          |
| `POSTGRES_PASSWORD` | Render Postgres Password              | DBパスワード                                                        |
| `POSTGRES_DB`       | `coffee_concierge`                    | DB名                                                                |
| `POSTGRES_HOST`     | Render Internal Host                  | Backend用。Render Web ServiceからPostgreSQLへ内部接続するためのhost |
| `POSTGRES_PORT`     | `5432`                                | DB port                                                             |
| `REDIS_HOST`        | Render Redis internal host            | Redis host                                                          |
| `REDIS_PORT`        | `6379`                                | Redis port                                                          |
| `GEMINI_API_KEY`    | 用意されたKey                         | Gemini API Key                                                      |
| `GEMINI_MODEL`      | `gemini-2.5-flash`                    | 使用モデル                                                          |
| `GEMINI_USE_MOCK`   | `false`                               | 本番ではfalse                                                       |

---

## 9. Gemini API Key差し替え手順

Gemini API Keyはコードに書かず、Renderの環境変数で設定します。
Gemini API Keyは、用意したKeyへ差し替えてください。

手順:

```text
Render
→ coffee-concierge-api
→ Environment
→ GEMINI_API_KEY を編集
→ Save Changes
→ Redeploy
```

設定値:

| Key               | Value                  |
| ----------------- | ---------------------- |
| `GEMINI_API_KEY`  | 用意したGemini API Key |
| `GEMINI_MODEL`    | `gemini-2.5-flash`     |
| `GEMINI_USE_MOCK` | `false`                |

確認方法:

```text
Render
→ coffee-concierge-api
→ Logs
→ [AI][AUDIT] で検索
```

Gemini成功時は、以下のようなログが出ます。

```text
[AI][AUDIT] kind=ai.success meta={"mode":"search_bundle","provider":"gemini","model":"gemini-2.5-flash","status":"success","final_selection_count":"5"}
```

---

## 10. GitHub Actions CI/CD

現在のworkflowは以下です。

```text
.github/workflows/ci-cd.yml
```

frontend / backend / Docker build / deploy / seedを1つのworkflowに統合しています。

### 10-1. 自動実行されるjob

| Job                        | 実行条件     | 内容                          |
| -------------------------- | ------------ | ----------------------------- |
| `Frontend lint and build`  | PR / push    | frontend lint/build           |
| `Backend test and build`   | PR / push    | Go test/build                 |
| `Backend Docker build`     | PR / push    | backend Dockerfile build      |
| `Deploy production`        | `main` push  | Render/Vercel Deploy Hook実行 |
| `Seed production manually` | 手動実行のみ | 本番DBへseed投入              |

---

## 11. GitHub Secrets

GitHub → Repository → Settings → Secrets and variables → Actionsに以下を登録します。

### 11-1. Deploy Hook用

| Secret                   | 説明                                |
| ------------------------ | ----------------------------------- |
| `RENDER_DEPLOY_HOOK_URL` | Render Web ServiceのDeploy Hook URL |
| `VERCEL_DEPLOY_HOOK_URL` | Vercel ProjectのDeploy Hook URL     |

### 11-2. 手動Seed用

| Secret                   | 説明                          |
| ------------------------ | ----------------------------- |
| `PROD_POSTGRES_USER`     | Render Postgres user          |
| `PROD_POSTGRES_PASSWORD` | Render Postgres password      |
| `PROD_POSTGRES_DB`       | Render Postgres DB name       |
| `PROD_POSTGRES_HOST`     | Render Postgres External Host |
| `PROD_POSTGRES_PORT`     | `5432`                        |
| `SEED_ADMIN_EMAIL`       | seedで作る管理者メール        |
| `SEED_ADMIN_PASSWORD`    | seed管理者パスワード          |

重要: `PROD_POSTGRES_HOST`はRenderのInternal Hostではなく、External Database URLのhost部分を入れます。

例:

```text
postgresql://coffee_user:xxxxx@xxxxx.oregon-postgres.render.com/coffee_concierge
```

この場合、`PROD_POSTGRES_HOST`は以下です。

```text
xxxxx.oregon-postgres.render.com
```

---

## 12. 本番seed手順

本番DBに検索対象データを入れる場合のみ実行します。通常は初回セットアップ時、または本番DBを作り直したときだけ実行します。

通常の`main` pushではseedは走りません。必ず手動実行です。DBに書き込む処理なので、連打しないでください。

### 12-1. GitHub CLIで実行する場合

```bash
gh workflow run ci-cd.yml --ref main -f run_seed=true
```

実行確認:

```bash
gh run list --workflow=ci-cd.yml --limit 5
```

最新runの詳細:

```bash
gh run view <RUN_ID>
```

失敗ログ:

```bash
gh run view <RUN_ID> --log-failed
```

### 12-2. GitHub UIで実行する場合

```text
GitHub
→ Actions
→ Coffee Concierge CI/CD
→ Run workflow
→ Branch: main
→ run_seed: true
→ Run workflow
```

成功条件:

```text
Seed production manually: success
seed completed
```

GitHub CLIでログを見る場合:

```bash
gh run view <RUN_ID>
gh run view <RUN_ID> --log
```

注意: seedは本番DBへ書き込みます。通常は初回だけ実行します。複数回実行する場合は、重複データが発生していないか確認してください。

---

## 13. 本番確認手順

### 13-1. 基本画面

| URL                                             | 成功条件          |
| ----------------------------------------------- | ----------------- |
| `https://coffee-concierge.vercel.app/`          | Topページ表示     |
| `https://coffee-concierge.vercel.app/login`     | Login画面表示     |
| `https://coffee-concierge.vercel.app/concierge` | Concierge画面表示 |

### 13-2. 認証

Chrome DevTools → Networkで確認します。

| API            | 成功条件 |
| -------------- | -------- |
| `/auth/csrf`   | 200      |
| `/auth/login`  | 200      |
| `/me`          | 200      |
| `/auth/logout` | 200      |

`/auth/*`のRequest URLは以下になるのが正しいです。

```text
https://coffee-concierge.vercel.app/auth/...
```

`https://coffee-concierge-api.onrender.com/auth/...`に直接飛んでいる場合は、`VITE_AUTH_BASE_URL`の設定またはVercel redeployを確認してください。

### 13-3. コンシェルジュ検索

1. login
2. `/concierge`に移動
3. 条件を入力
4. 検索実行
5. RANK 1〜5の候補が出ることを確認

Networkで以下を確認します。

| API                          | 成功条件 |
| ---------------------------- | -------- |
| `/search/sessions`           | 201      |
| `/search/sessions/{id}/pref` | 200      |

---

## 14. Gemini成功ログ確認

RenderのWeb Service側ログを見ます。Postgresのログではありません。

```text
Render
→ coffee-concierge-api
→ Logs
→ [AI][AUDIT] で検索
```

成功時の例:

```text
[AI][AUDIT] kind=ai.success meta={"candidate_count":"20","duration_ms":"27828","fetch_count":"1000","final_selection_count":"5","followup_count":"3","gemini_selection_count":"5","mode":"search_bundle","model":"gemini-2.5-flash","provider":"gemini","result_limit":"5","status":"success"}
```

| ログ                      | 意味                     |
| ------------------------- | ------------------------ |
| `kind=ai.request`         | Geminiへのリクエスト開始 |
| `kind=ai.success`         | Gemini成功               |
| `kind=ai.failed`          | Gemini処理失敗           |
| `kind=ai.fallback`        | Go側rankerへfallback     |
| `provider=gemini`         | Gemini使用               |
| `model=gemini-2.5-flash`  | 使用モデル               |
| `final_selection_count=5` | 最終候補5件              |

---

## 15. パスワード再設定

現状のパスワード再設定は、実メール送信ではなく、Render Logsに再設定URLを出力する方式です。これは検証用の実装です。

`/auth/password/forgot`を実行すると、Backendが再設定tokenを生成し、以下のようなログを`coffee-concierge-api → Logs`に出力します。

```text
[MAIL][RESET] to=user@example.com link=https://coffee-concierge.vercel.app/reset-password?token=...
```

手順:

1. `/login`のパスワード再設定からメールを入力
2. Networkで`/auth/password/forgot`が200になることを確認
3. Render `coffee-concierge-api → Logs`を開く
4. `[MAIL][RESET]`のログを探す
5. `https://coffee-concierge.vercel.app/reset-password?token=...`をブラウザで開く
6. 新しいパスワードを設定
7. 新パスワードでloginできることを確認

つまり、現状は以下です。

```text
再設定URLを作る → Render Logsに出す
```

実運用で一般ユーザーに提供する場合は、Render LogsにURLを出すだけでは不十分です。ユーザーのメール受信箱に再設定メールを届ける必要があるため、Resend / SendGrid / Mailgun / AWS SESなどのメール送信サービスへ差し替えます。

本番運用では以下に変更します。

```text
再設定URLを作る → メール送信サービスAPIを使ってユーザーへメール送信する
```

この差し替えでは、認証フロー全体を作り直す必要はありません。`Mailer`の具体実装を、ログ出力用の実装から本番メール送信用の実装へ変更します。

---

## 16. よくある失敗と対応

| 症状                           | 原因                           | 対応                                               |
| ------------------------------ | ------------------------------ | -------------------------------------------------- |
| `/login`直アクセスで404        | SPA rewrite不足                | `frontend/vercel.json`を確認                       |
| auth APIがRender直通になる     | `VITE_AUTH_BASE_URL`未設定     | Vercel envを設定してRedeploy                       |
| logoutが400/403                | Cookie/CSRF不整合              | `/auth/*` proxy, Cookie, CSRF header確認           |
| seedが止まる                   | DB hostがInternal Host         | GitHub SecretにExternal Hostを設定                 |
| seedがduplicate keyで失敗      | 既存メール重複                 | `SEED_ADMIN_EMAIL`を既存ユーザーと被らない値に変更 |
| Geminiログが出ない             | API側ログを見ていない          | `coffee-concierge-api → Logs`を確認                |
| Gemini成功か不明               | `[AI][AUDIT]`未確認            | `kind=ai.success`を探す                            |
| Vercel env変更が反映されない   | 再ビルドしていない             | Redeployする                                       |
| GitHub Actionsでseedが走らない | pushイベントではseedしない設計 | `workflow_dispatch run_seed=true`で手動実行        |

---

## 17. 完了条件

以下が確認できれば、本番デプロイ確認は完了です。

- Vercel本番URLでTopページが表示される
- `/login`と`/concierge`に直アクセスできる
- login / logoutが200で成功する
- 本番DBにseedデータが入っている
- コンシェルジュ検索でRANK 1〜5が表示される
- Render API Logsに`[AI][AUDIT] kind=ai.success`が出る
- `provider=gemini`が確認できる
- `model=gemini-2.5-flash`が確認できる
- `final_selection_count=5`が確認できる
- GitHub Actionsの`Coffee Concierge CI/CD`が成功する

---

## 18. 補足

このアプリは、単なる静的サイトではなく、以下を含む本番構成です。

- VercelによるFrontend Production Deployment
- RenderによるGo/Echo API Production Deployment
- PostgreSQLによる永続データ管理
- Redisによる補助機能
- Gemini APIによるコーヒー候補選定
- CSRF / Cookieを考慮した認証フロー
- GitHub ActionsによるCI/CD
- Deploy Hookによる本番反映
- 手動seed workflowによる本番初期データ投入
- Render LogsによるAI成功証跡

Gemini API Keyは環境のものへ差し替えてください。差し替え後はRenderのRedeployが必要です。
