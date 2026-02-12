# FleetTrack - 開発進捗サマリー

最終更新: 2026-02-12

---

## 全体概要

FleetTrackはタクシー/ハイヤー事業者向けの車両配車・運行管理システム。
バックエンド(Go)、フロントエンド(React)、モバイルアプリ(React Native)の3層構成。

---

## 1. バックエンド (Go) — 完成度: 98%

### 完了済み

| カテゴリ | 内容 |
|---------|------|
| API | 49エンドポイント実装済 (認証, ユーザー, 車両, 配車, 位置情報, 予約, ダッシュボード, 設定) |
| 認証 | JWT (access + refresh token), RBAC (admin/dispatcher/driver) |
| データベース | PostgreSQL + PostGIS, マイグレーション, シードデータ |
| ミドルウェア | CORS, 認証, ロール別アクセス制御, リクエストログ, レート制限 |
| リアルタイム更新 | ポーリング (10-15秒間隔) による位置情報・配車通知 |
| ファイル管理 | アバター画像アップロード (uploads/) |
| テスト | 33ユニットテスト (JWT, 認証ミドルウェア, Config, Model, AppError, レート制限) |
| APIドキュメント | OpenAPI 3.0 仕様 + Swagger UI (`/api/docs`) |
| レート制限 | トークンバケット方式 (20 req/s, burst 40) per IP |

### 未完了

| 項目 | 優先度 |
|------|--------|
| ハンドラー統合テスト | 中 |

---

## 2. フロントエンド (React + TypeScript + Vite) — 完成度: 95%

### 完了済み

| カテゴリ | 内容 |
|---------|------|
| 認証 | ログイン, ログアウト, トークンリフレッシュ, ロールベースルーティング |
| ダッシュボード | 統一マップUI, リアルタイム車両表示, Google Maps連携 |
| 配車フロー | BottomSheet UI, 出発地/目的地選択, 車両選択, 予約送信 |
| 車両管理 | 一覧, 詳細, 登録, 編集, 削除 |
| ドライバー管理 | 一覧, 詳細, 登録, 編集, 削除, 勤怠表示 |
| 予約管理 | 一覧表示, ステータス管理 |
| ユーザー管理 | 一覧, 登録, 編集, 削除 |
| 設定 | 会社情報, システム設定 |
| UI/UX | レスポンシブ, BottomSheet fitContent, 車両凡例, サイドバー |
| エラーハンドリング | ErrorBoundary (グローバル) |
| コード品質 | ESLint エラー 0件, TypeScript strict モード |

### 補足: i18n (国際化) — 対応済み
Web フロントエンドは 4言語対応済み (en / ja / ko / zh)。動的ロケール読み込み・キャッシュ機構も実装済み。

### 未完了

| 項目 | 優先度 |
|------|--------|
| E2Eテスト (Playwright等) | 低 |

---

## 3. モバイルアプリ (React Native 0.78) — 完成度: 60%

### 完了済み

| カテゴリ | 内容 |
|---------|------|
| プロジェクト構成 | React Native 0.78, TypeScript, Zustand, React Navigation |
| 認証 | ログイン/ログアウト, EncryptedStorageによるトークン永続化, 自動復元 |
| API通信 | Axios, Platform別ベースURL (iOS/Android), トークンインターセプター |
| 画面 | ログイン, ホーム (出勤/退勤), トリップ一覧, トリップ詳細, プロフィール |
| 位置情報 | @react-native-community/geolocation, 15秒間隔レポート |
| 通知 | Firebase未設定時のグレースフルフォールバック |
| ネイティブ | iOS/Androidプロジェクト生成済み, npm install完了 |

### 未完了

| 項目 | 優先度 |
|------|--------|
| pod install (Xcode必要) | **高** |
| iOSビルド・実機テスト | **高** |
| バックグラウンド位置情報 | 中 |
| プッシュ通知 (Firebase設定) | 中 |
| アプリアイコン・スプラッシュ | 低 |

---

## 4. インフラ — 完成度: 60%

### 完了済み

| カテゴリ | 内容 |
|---------|------|
| Git | リポジトリ初期化, .gitignore設定済 |
| 環境変数 | .env.example (バックエンド) |
| Docker | docker-compose.yml (PostgreSQL + PostGIS) |
| CI/CD | GitHub Actions (Go テスト+ビルド, React lint+ビルド) |

### 未完了

| 項目 | 優先度 |
|------|--------|
| 本番デプロイ設定 | 中 |
| HTTPS/TLS設定 | 中 |
| 監視・ログ集約 | 低 |

---

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| バックエンド | Go, Chi, sqlx, PostgreSQL, PostGIS, JWT |
| フロントエンド | React 19, TypeScript, Vite, Zustand, Google Maps API, plain CSS (custom properties) |
| モバイル | React Native 0.78, TypeScript, Zustand, React Navigation 7, EncryptedStorage |
| インフラ | Docker Compose, PostgreSQL 16 + PostGIS, GitHub Actions CI |
