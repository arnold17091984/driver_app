# FleetTrack - 次にすべきこと

最終更新: 2026-02-12

---

## 即時対応 (ブロッカー)

### 1. Xcodeインストール完了 → iOS ビルド

**状態**: Xcodeダウンロード中

```bash
# Xcodeインストール確認
xcode-select -p
xcodebuild -version

# Command Line Tools設定
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer

# CocoaPodsインストール
cd mobile/ios
pod install
cd ..

# iOSシミュレータビルド
npx react-native run-ios
```

### 2. iOS実機テスト

**前提条件**:
- Apple Developer Account (無料版: 7日間有効)
- iPhone実機 (USB接続)
- Mac と iPhone が同一Wi-Fi

**手順**:
1. `ios/DriverApp.xcworkspace` をXcodeで開く
2. Signing & Capabilities → Personal Team選択
3. Bundle Identifier → ユニーク名に変更 (例: `com.yourname.driverapp`)
4. 接続デバイスを選択 → Run

**API接続設定**:
```bash
# MacのIPアドレス確認
ifconfig en0 | grep inet
```

`mobile/src/services/apiClient.ts` の `setApiBase()` でMacのIPを指定:
```typescript
import { setApiBase } from './services/apiClient';
setApiBase('192.168.x.x'); // MacのローカルIP
```

### 3. E2Eテストシナリオ

| # | ステップ | 確認内容 |
|---|---------|---------|
| 1 | ログイン | driver001 / password123 |
| 2 | ホーム画面 | ユーザー名、出勤ボタン表示 |
| 3 | 出勤 | タップ → ステータス変更 |
| 4 | Web側で配車 | PCブラウザから配車を作成 |
| 5 | 配車確認 | モバイルでトリップ表示 |
| 6 | トリップ進行 | 受領→移動中→到着→完了 |
| 7 | 退勤 | タップ → ステータス変更 |

---

## 短期 (1-2週間)

### 4. バックグラウンド位置情報

現在はフォアグラウンドのみ。バックグラウンド対応には:
- `react-native-background-geolocation` (Transistor Software) の導入
- iOS: Background Modes → Location updates 有効化
- Android: Foreground Service設定

### 5. プッシュ通知 (Firebase)

1. Firebase Console でプロジェクト作成
2. `GoogleService-Info.plist` (iOS) / `google-services.json` (Android) 設置
3. `@react-native-firebase/app` + `@react-native-firebase/messaging` インストール
4. `notificationService.ts` のFirebase連携を有効化

### 6. バックエンドテスト

```bash
cd backend
go test ./... -v
```

主要テスト対象:
- 認証フロー (ログイン, トークンリフレッシュ)
- 配車CRUD
- 位置情報レポート
- RBAC (権限チェック)

---

## 中期 (2-4週間)

### 7. CI/CD パイプライン

GitHub Actionsで:
- Go テスト + ビルド
- React ビルド + Lint
- React Native ビルド (iOS/Android)

### 8. 本番デプロイ

- バックエンド: AWS ECS / GCP Cloud Run
- フロントエンド: Vercel / Cloudflare Pages
- データベース: AWS RDS / GCP Cloud SQL (PostGIS対応)
- HTTPS: Let's Encrypt / Cloudflare

### 9. モバイルアプリ配布

- iOS: TestFlight (Apple Developer Program $99/年 必要)
- Android: Google Play Internal Testing

---

## 長期 (1-3ヶ月)

### 10. 追加機能

- 乗客向けアプリ (予約・追跡)
- 運賃自動計算
- ドライバー評価システム
- 運行レポート・分析ダッシュボード
- 多言語対応 (i18n)

---

## ファイル構成

```
driver/
├── backend/           # Go REST API
│   ├── cmd/           # エントリポイント
│   ├── internal/      # ハンドラー, モデル, サービス
│   └── migrations/    # DBマイグレーション
├── web/               # React フロントエンド
│   ├── src/
│   │   ├── components/  # UI コンポーネント
│   │   ├── pages/       # ページ
│   │   ├── stores/      # Zustand ストア
│   │   └── services/    # API クライアント
│   └── public/
├── mobile/            # React Native モバイルアプリ
│   ├── src/
│   │   ├── screens/     # 画面
│   │   ├── stores/      # Zustand ストア
│   │   ├── services/    # API, 位置情報, 通知
│   │   └── navigation/  # React Navigation
│   ├── ios/             # Xcode プロジェクト
│   └── android/         # Android プロジェクト
├── docs/              # ドキュメント
└── docker-compose.yml # PostgreSQL + PostGIS
```
