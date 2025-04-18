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
- 記事内の画像を自動ダウンロード（オプション）
- フォーマット変換エラーに対する複数の対応方法

## 依存関係
- Golang（推奨バージョン：1.20以降）
- [github.com/JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTMLからMarkdownへの変換
- [github.com/yosssi/gohtml](https://github.com/yosssi/gohtml) - HTML整形に使用
- [github.com/stretchr/testify](https://github.com/stretchr/testify) - テスト用ユーティリティ
- [github.com/imroc/req/v3](https://github.com/imroc/req) - HTTPリクエスト（画像ダウンロード）

## インストール方法

```sh
# リポジトリのクローン
git clone https://github.com/yourusername/mt2hugo.git
cd mt2hugo

# 依存パッケージのインストール
go mod download
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


- --no-markdown: HTMLをMarkdownに変換せず、そのまま出力します
- --format-html: HTMLを階層構造でフォーマットして出力します (--no-markdownと共に使用)
- --output <ディレクトリ>: 出力先ディレクトリを指定します (デフォルト: "output")
- --error-mode <モード>: Markdown変換エラー時の挙動を指定します (html, error, partial)
- --verbose: 詳細なログを出力します
- --download-images: 記事内の画像をダウンロードします　画像のダウンロード時、アイキャッチがない場合は最初の画像をアイキャッチにします。
- --max-concurrent <数値>: 同時ダウンロード数を指定します（デフォルト: 5）
- --image-timeout <秒数>: 画像ダウンロードのタイムアウト秒数（デフォルト: 30）
- --help: 使用法を表示します

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

## エラー処理オプション

mt2hugoでは、Markdown変換エラー発生時の挙動を指定できます：

- `--error-mode=html`: エラーが発生した場合、元のHTMLをそのまま出力します（デフォルト）
- `--error-mode=error`: エラーが発生した場合、処理を中止します
- `--error-mode=partial`: エラーが発生した場合、部分的な変換結果を出力します

例えば、厳格な変換を行いたい場合は以下のように実行します：

## 入力ファイル形式

Movable Typeエクスポートファイルは以下の要件を満たす必要があります：

- 必須フィールド: `TITLE`, `DATE`
- サポートしている日付形式:
  - MM/DD/YYYY HH:MM:SS
  - MM/DD/YYYY 00:00:00
  - その他拡張予定

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
- `<br>` タグの正規化

この機能は、github.com/yosssi/gohtml パッケージを使用して実装されており、HTMLを直接編集したい場合や、Markdown変換の結果が望ましくない場合に特に便利です。

## プロジェクト構成

```
mt2hugo/
  ├── main.go                  # エントリーポイント
  ├── fs/                      # ファイルシステム操作
  │   ├── filesystem.go        # ファイル操作インターフェースと実装
  │   └── filesystem_test.go   # ファイルシステムのテスト
  ├── converter/               # コンテンツ変換
  │   ├── converter.go         # 変換インターフェースとメイン実装
  │   ├── html/                # HTML関連変換処理
  │   │   ├── converter.go     # HTML変換実装
  │   │   └── converter_test.go # HTML変換テスト
  ├── downloader/              # 画像ダウンロード機能
  │   └── imagedownloader.go   # 画像ダウンロード実装
  ├── internal/                # 内部パッケージ
  │   ├── config/              # 設定関連
  │   │   ├── config.go        # 設定構造体
  │   │   └── flag.go          # コマンドライン引数解析
  │   ├── errors/              # エラー処理
  │   ├── factory/             # ファクトリパターン実装
  │   └── processor/           # 処理ロジック
  ├── models/                  # データモデル定義
  │   ├── movabletype.go       # Movable Type記事モデル
  │   └── hugo.go              # Hugo記事モデル
  ├── parser/                  # 入力ファイル解析
  │   └── mtparser/            # Movable Type解析
  │       ├── parser.go        # MTエクスポートファイル解析機能
  │       └── parser_test.go   # パーサーのテスト
  ├── reporter/                # 進捗レポート
  │   ├── progress.go          # 進捗表示
  │   ├── mock_reporter.go     # テスト用モック
  │   └── result.go            # 結果レポート
  ├── transformer/             # 変換ロジック
  │   ├── mt2hugo/             # MT→Hugo変換
  │   │   ├── mt2hugo.go       # 変換ロジック本体
  │   │   └── transformer_test.go # 変換ロジックのテスト
  ├── generator/               # 出力ロジック
  │   ├── filegenerator.go     # ファイル生成機能
  │   └── generator_test.go    # 生成機能のテスト
  ├── util/                    # ユーティリティ
  │   ├── date.go              # 日付処理
  │   └── slug.go              # スラグ生成
  ├── validator/               # 入力検証
  │   └── validator.go         # 検証ロジック
  └── templates/               # テンプレート
      └── hugo.tmpl            # Hugo記事テンプレート
```

# パッケージの責務

## models パッケージ
- データモデルの定義
    - `movabletype.go`: Movable Type記事のデータ構造
    - `hugo.go`: Hugo記事のデータ構造

## parser パッケージ
- 入力ファイルの解析
    - `mtparser/parser.go`: Movable Typeエクスポートファイルを解析し、記事モデルに変換

## converter パッケージ
- コンテンツ変換を担当
    - `html/converter.go`: HTMLからMarkdownへの変換処理、HTML整形機能

## downloader パッケージ
- 画像ダウンロード処理
    - `imagedownloader.go`: 画像URL抽出、ダウンロード、パス置換

## transformer パッケージ
- `mt2hugo/mt2hugo.go`: Movable Type記事からHugo記事への変換ロジック
    - データの検証
    - メタデータの変換と最適化
    - 本文の処理（MarkdownへのHTML変換など）

## generator パッケージ
- `filegenerator.go`: 出力ファイルの生成
    - 出力先ディレクトリの作成
    - テンプレートの適用
    - ファイル書き込み処理
    - ファイル命名規則の管理

## util パッケージ
- 共通ユーティリティ機能
    - `date.go`: 日付処理機能
    - `slug.go`: スラグ生成と変換

## fs パッケージ
- ファイルシステム操作の抽象化
    - 実際のファイルシステムとテストモックの両方に対応

## reporter パッケージ
- 処理進捗の報告と表示
- 結果レポートの生成
- テスト用モック実装

## validator パッケージ
- 入力データの検証
- エラーメッセージの生成

## internal/config パッケージ
- 設定管理とコマンドライン引数解析

# テンプレートについて

mt2hugoでは、Hugoの記事生成に使用するテンプレートファイルが必要です。デフォルトでは、`templates/hugo.tmpl`が使用されます。

このファイルが存在しない場合、プログラムはエラーを表示して終了します。テンプレートファイルは必ず用意してください。

# 注意事項

- Movable Typeエクスポートファイルには必ず「TITLE」と「DATE」フィールドが含まれている必要があります。
- 「DATE」フィールドが欠けている場合や日付形式が不正な場合、変換処理はエラーとなりますが、他の記事の処理は継続します。
- 変換中にエラーが発生した場合でも、可能な限り処理を継続し、問題のある記事だけをスキップします。
- HTMLのフォーマット機能は、`--no-markdown`オプションと共に使用する必要があります。
- 画像ダウンロード機能を使用する場合、オリジナルの画像が利用できなくなると参照できなくなる可能性があります。

# 今後の開発予定

- DeepL API Freeを使用し、記事タイトルからSlug用の英語テキストを生成する機能
- より多くの日付フォーマットのサポート
- HTML整形オプションのカスタマイズ機能
- テスト範囲のさらなる拡充
- 並行処理の導入: 多数の記事を処理する場合、goroutinesを使用した並行処理で処理速度を向上できます。
- メモリ効率の改善: 大きなファイルを処理する際のメモリ使用量を最適化します。
- プログレスレポーティングの改善: 処理終了時間の予測表示などの機能を追加します。
- エラー回復機能の強化: 一部の記事で問題が発生しても、できるだけ多くの記事を処理できるようにします。
- 画像処理の拡張: 画像リサイズや最適化機能の追加
- 設定ファイルのサポート: YAMLやJSONによる設定ファイルのサポート
- movalbletypeのカテゴリはhugoのtagに変換する

# リファクタリングの予定

- **エラーハンドリングの一貫性向上**  
    現在、一部のエラーが単純に文字列として結合されており、他の部分では`%w`を使ってエラーをラップしています。`%w`を一貫して使うことで、`errors.Is`や`errors.As`を使って元のエラーを検査できるようになります。
- **設定オプションの集約**
- **テストカバレッジの向上**
- **コードスタイルの統一**
- **ログ機能の一元化**  
    現在の実装では、`fmt.Printf`や`reporter`インターフェースを通じてログを出力しています。構造化ログ（例：zerologやzap）を使用することで、より一貫性のあるログ出力ができます。
- **設定ファイル対応**  
    現在はコマンドライン引数のみで設定を行っていますが、設定ファイル（YAML/JSON/TOML）からオプションを読み込める機能を追加すると便利です。
- **関心の分離をさらに進める**  
    特に`main.go`の肥大化を防ぎ、テスト容易性を向上させる
- **コンポーネント間の依存関係の整理**  
    依存性注入の導入により、コンポーネントの交換や再利用を容易にする