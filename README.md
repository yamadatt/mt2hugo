# mt2hugo
Movable TypeのエクスポートファイルからHugo用のファイルを作成するツールです。

## 概要
本ツールは、Movable Typeエクスポートファイルを解析し、各記事ごとに日付に基づいたディレクトリ構造でHugoのコンテンツファイル（index.md）を生成します。本文のHTMLは、Markdownへ自動変換されます（変換に失敗した場合は元のHTMLが出力されます）。また、HTMLをそのまま出力する際に、見やすく整形することも可能です。

## 機能
- Movable Typeエクスポートファイルの解析
- 記事ごとのHugoフォーマットのファイル生成
- HTMLからMarkdownへの自動変換
- HTMLの整形出力（インデント付き階層構造）
- 柔軟な日付形式対応
- メタデータ（タグ、カテゴリなど）の保持

## 依存関係
- Golang（推奨バージョン：1.20以降）
- [github.com/JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTMLからMarkdownへの変換
- [github.com/yosssi/gohtml](https://github.com/yosssi/gohtml) - HTML整形に使用

## インストール方法

```sh
# リポジトリのクローン
git clone https://github.com/yourusername/mt2hugo.git
cd mt2hugo

# 依存パッケージのインストール
go get -u github.com/JohannesKaufmann/html-to-markdown
go get -u github.com/yosssi/gohtml
```

## 使用方法

### 基本的な使い方

```sh
go run main.go <Movable_Typeエクスポートファイルのパス>
```

### コマンドラインオプション

```sh
go run main.go [オプション] <Movable_Typeエクスポートファイルのパス>
```

**利用可能なオプション:**

- `--no-markdown`: HTMLをMarkdownに変換せず、そのまま出力します
- `--format-html`: HTMLを階層構造でフォーマットして出力します (--no-markdownと共に使用)
- `--output <ディレクトリ>`: 出力先ディレクトリを指定します (デフォルト: "output")

### 例

1. 標準的な変換（HTMLをMarkdownに変換）:
   ```sh
   go run main.go export.txt
   ```

2. HTMLをそのまま出力（Markdown変換なし）:
   ```sh
   go run main.go --no-markdown export.txt
   ```

3. HTMLをフォーマットして出力:
   ```sh
   go run main.go --no-markdown --format-html export.txt
   ```

4. 出力先ディレクトリの指定:
   ```sh
   go run main.go --output hugo-content export.txt
   ```

## 入力ファイル形式

Movable Typeエクスポートファイルは以下の要件を満たす必要があります：

- 必須フィールド: `TITLE`, `DATE`
- サポートしている日付形式:
  - MM/DD/YYYY HH:MM:SS
  - YYYY-MM-DD HH:MM:SS
  - MM/DD/YY HH:MM:SS
  - DD/MM/YYYY HH:MM:SS
  - YYYY/MM/DD HH:MM:SS
  - MM/DD/YYYY HH:MM

## 出力形式

1. **ディレクトリ構造**:
   ```
   output/
   └── YYYY/
       └── MM/
           └── DD/
               └── HHMM/
                   └── index.md
   ```

2. **メタデータ形式**:
   ```markdown
   ---
   title: "記事タイトル"
   date: YYYY-MM-DDThh:mm:ss-07:00
   slug: 記事スラグ
   category:
     - カテゴリ名
   tags:
     - タグ1
     - タグ2
   draft: false
   showtoc: false
   ---

   本文内容
   ```

## HTML整形機能について

`--no-markdown --format-html` オプションを使用すると、HTMLを美しく整形して出力します：

- インデントによる階層構造表示
- タグや属性の整理
- 読みやすいスペーシングの適用

この機能は、HTMLを直接編集したい場合や、Markdown変換の結果が望ましくない場合に便利です。

## プロジェクト構成

```
mt2hugo/
  ├── main.go                # エントリーポイント
  ├── fs/                    # ファイルシステム操作
  │   └── filesystem.go
  ├── converter/             # HTML→Markdown変換、HTML整形
  │   └── converter.go
  ├── movabletype/           # Movable Type関連
  │   ├── models.go          # 構造体定義
  │   └── parser.go          # パース処理
  ├── hugo/                  # Hugo関連
  │   ├── models.go          # 構造体定義
  │   └── converter.go       # 変換処理
  └── util/                  # ユーティリティ
      └── date.go            # 日付処理など
```

## 注意事項

- Movable Typeエクスポートファイルには必ず「TITLE」と「DATE」フィールドが含まれている必要があります。
- 「DATE」フィールドが欠けている場合や日付形式が不正な場合、変換処理はエラーとなります。
- 変換中にエラーが発生した場合でも、可能な限り処理を継続し、問題のある記事だけをスキップします。
- HTMLのフォーマット機能は、`--no-markdown`オプションと共に使用する必要があります。

## 今後の開発予定

- DeepL API Freeを使用し、記事タイトルからSlug用の英語テキストを生成する機能
- より多くの日付フォーマットのサポート
- 画像ファイルの取り扱い改善
- テスト範囲の拡充

## ライセンス

このプロジェクトは[MITライセンス](LICENSE)の下で提供されています。

## 開発者

Your Name - [your-email@example.com]
