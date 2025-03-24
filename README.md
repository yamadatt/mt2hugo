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
- [github.com/stretchr/testify](https://github.com/stretchr/testify) - テスト用ユーティリティ

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
  │   ├── html_formatter.go    # HTML整形用ユーティリティ関数
  │   ├── html_to_markdown.go  # HTMLからMarkdownへの変換
  │   ├── constants.go         # 定数定義
  │   └── converter_test.go    # 変換機能のテスト
  ├── models/                  # データモデル定義
  │   ├── movabletype.go       # Movable Type記事モデル
  │   └── hugo.go              # Hugo記事モデル
  ├── parser/                  # 入力ファイル解析
  │   └── mtparser/            # Movable Type解析
  │       ├── parser.go        # MTエクスポートファイル解析機能
  │       └── parser_test.go   # パーサーのテスト
  ├── transformer/             # 変換ロジック
  │   ├── mt2hugo.go           # MT→Hugo変換ロジック
  │   └── transformer_test.go  # 変換ロジックのテスト
  ├── generator/               # 出力ロジック
  │   ├── filegenerator.go     # ファイル生成機能
  │   └── generator_test.go    # 生成機能のテスト
  └── util/                    # ユーティリティ
      ├── date.go              # 日付処理
      └── slug.go              # スラグ生成
```

## 各パッケージの責務

### models パッケージ
- データモデルの定義
- **movabletype.go**: Movable Type記事のデータ構造
- **hugo.go**: Hugo記事のデータ構造

### parser パッケージ
- 入力ファイルの解析
- **mtparser/parser.go**: Movable Typeエクスポートファイルを解析し、記事モデルに変換

### converter パッケージ
- コンテンツ変換を担当
- **html_to_markdown.go**: HTMLからMarkdownへの変換処理
- **html_formatter.go**: HTML整形機能

### transformer パッケージ
- **mt2hugo.go**: Movable Type記事からHugo記事への変換ロジック
- データの検証
- メタデータの変換と最適化
- 本文の処理（MarkdownへのHTML変換など）

### generator パッケージ
- **filegenerator.go**: 出力ファイルの生成
- 出力先ディレクトリの作成
- テンプレートの適用
- ファイル書き込み処理
- ファイル命名規則の管理

### util パッケージ
- 共通ユーティリティ機能
- **date.go**: 日付処理機能
- **slug.go**: スラグ生成と変換

### fs パッケージ
- ファイルシステム操作の抽象化
- 実際のファイルシステムとテストモックの両方に対応

## テンプレートについて

mt2hugoでは、Hugoの記事生成に使用するテンプレートファイルが必要です。
デフォルトでは、`templates/hugo.tmpl`が使用されます。

このファイルが存在しない場合、プログラムはエラーを表示して終了します。
テンプレートファイルは必ず用意してください。

テンプレートの例は[ドキュメント](docs/template.md)を参照してください。

## 注意事項

- Movable Typeエクスポートファイルには必ず「TITLE」と「DATE」フィールドが含まれている必要があります。
- 「DATE」フィールドが欠けている場合や日付形式が不正な場合、変換処理はエラーとなりますが、他の記事の処理は継続します。
- 変換中にエラーが発生した場合でも、可能な限り処理を継続し、問題のある記事だけをスキップします。
- HTMLのフォーマット機能は、`--no-markdown`オプションと共に使用する必要があります。

## 今後の開発予定

- DeepL API Freeを使用し、記事タイトルからSlug用の英語テキストを生成する機能
- より多くの日付フォーマットのサポート
- 画像ファイルの自動抽出と整理機能 
- HTML整形オプションのカスタマイズ機能
- テスト範囲のさらなる拡充
- 並行処理の導入: 多数の記事を処理する場合、goroutinesを使用した並行処理で処理速度を向上できます。
- メモリ効率の改善: 大きなファイルを処理する際のメモリ使用量を最適化します。
- プログレスレポーティングの改善: 処理終了時間の予測表示などの機能を追加します。
- エラー回復機能の強化: 一部の記事で問題が発生しても、できるだけ多くの記事を処理できるようにします。

## リファクタリングの予定

- エラーハンドリングの一貫性向上
  - 現在、一部のエラーが単純に文字列として結合されており、他の部分では%wを使ってエラーをラップしています。%wを一貫して使うことで、errors.Isやerrors.Asを使って元のエラーを検査できるようになります。
- 設定オプションの集約
- テストカバレッジの向上
- コードスタイルの統一
- ログ機能の一元化
  - 現在の実装では、fmt.Printfやreporterインターフェースを通じてログを出力しています。構造化ログ（例：zerologやzap）を使用することで、より一貫性のあるログ出力ができます：
- 設定ファイル対応
  - 現在はコマンドライン引数のみで設定を行っていますが、設定ファイル（YAML/JSON/TOML）からオプションを読み込める機能を追加すると便利です



1. 複数の責務が混在している
現在のmain.goは以下の複数の責務を担っています：

コマンドライン引数の解析と検証
初期化処理（複数のコンポーネントの作成）
ファイル処理のビジネスロジック
エラー処理とエラーメッセージの整形
進捗管理と表示
メモリ使用状況の監視と表示
統計情報の収集と表示
これらの責務は明確に分離されるべきです。

2. 高い循環的複雑度
記事処理ループ内のロジックは条件分岐が多く、循環的複雑度が高いです。特にエラー処理の部分（記事の日付やタイトルの有無によるエラーメッセージの組み立て）は複雑です。

3. 設定とロジックが混在
コマンドライン引数の解析と実際の処理ロジックが同じファイルに混在しています。設定関連のコードは独立したパッケージに分離すべきです。

4. 初期化コードの肥大化
各種コンポーネント（fileSystem, htmlConverter, transformer, fileGenerator など）の初期化コードがmain.goに直接書かれており、ファイルが肥大化しています。

5. エラー処理の一貫性がない
エラー処理が各所に散在しており、一貫したアプローチになっていません。特に以下の点で問題があります：

一部の箇所では早期リターン
他の箇所ではエラーを収集して後で表示
エラーロギングとエラーハンドリングが混在
6. テスト容易性の低さ
現在の構造ではmain.goの機能をユニットテストすることが難しいです。コマンドライン処理、実際のビジネスロジック、出力処理などが密結合しているためです。

7. ハードコードされた文字列の多用
エラーメッセージやログメッセージが直接コード内にハードコードされており、将来的な国際化や文言変更が難しくなっています。

8. メモリ管理機能が主ロジックと混在
メモリ使用状況のレポート機能が主要なロジックと混在しています。これは独立した監視モジュールとして分離すべきです。

9. 拡張性の制約
新機能を追加する場合、現在の構造ではmain.goを大幅に変更する必要があります。これは保守性を低下させ、バグ混入のリスクを高めます。

10. コンポーネント間の密結合
依存関係の注入が直接的で、main.goがすべてのコンポーネントの初期化と連携を担当しています。これにより、コンポーネントの交換や再利用が困難になっています。