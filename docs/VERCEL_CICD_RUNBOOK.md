# Coffee Concierge Vercel CI/CD Runbook

## 目的

Coffee Concierge の frontend を Vercel に自動デプロイする。

対象は frontend の Vite/React アプリ。
backend の Go/Echo API、PostgreSQL、Redis は Vercel に直接載せない。
API は別サービスにデプロイし、その公開 URL を VITE_API_BASE_URL に設定する。

## 0. 先に確認すること

.env を GitHub に上げない。

確認コマンド:

    git status --short
    git ls-files .env frontend/.env backend/.env

何も表示されなければ OK。

## 1. ローカルで frontend を確認

    cd frontend
    npm ci
    npm run lint
    npm run build
    cd ..

成功条件:

- npm run lint が成功する
- npm run build が成功する
- frontend/dist が生成される

## 2. Vercel プロジェクト設定

| 項目 | 値 |
|---|---|
| Framework Preset | Vite |
| Root Directory | frontend |
| Install Command | npm ci |
| Build Command | npm run build |
| Output Directory | dist |

## 3. Vercel 環境変数

Vercel に以下を設定する。

| Key | 値 |
|---|---|
| VITE_API_BASE_URL | 本番API URL |

本番では localhost を指定しない。

## 4. backend 側 CORS

Vercel の本番URLを backend の FE_URL に設定する。

例:

    FE_URL=https://coffee-concierge.vercel.app

## 5. ブランチ運用

| ブランチ | 役割 |
|---|---|
| chore/vercel-cicd | 作業ブランチ |
| develop | 検証・統合 |
| main | 本番 |

流れ:

    chore/vercel-cicd
    ↓
    develop
    ↓
    main
    ↓
    Vercel Production Deployment

## 6. 完了条件

- GitHub Actions の frontend-ci が成功する
- Vercel Preview URL が発行される
- main merge で Production Deployment が成功する
- Vercel本番URLで画面が表示される
- /login 直アクセスで404にならない
- /concierge 直アクセスで404にならない
- API URL が localhost ではなく本番APIを向いている
- backend 側 CORS の FE_URL が Vercel本番URLと一致している
- .env が GitHub に追跡されていない
