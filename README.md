
DeepL API Freeを使用し、SlugにはTITLEで書かれている日本語をインプットにして、端的でわかりやすい英語にしたい。

# mt2hugo
Movable TypeのエクスポートファイルからHugo用のファイルを作成するツールです。

## 概要
本ツールは、Movable Typeエクスポートファイルを解析し、各記事ごとに日付に基づいたディレクトリ構造でHugoのコンテンツファイル（index.md）を生成します。本文のHTMLは、Markdownへ自動変換されます（変換に失敗した場合は元のHTMLが出力されます）。

## 依存関係
- Golang（推奨バージョン：1.XX以降）
- [github.com/JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown)  
  ※ HTMLからMarkdownへの変換に使用します。

## 使用方法

1. Golangをインストールします。
2. このリポジトリをクローンします。
3. 必要な外部パッケージを取得します。
   ```sh
   go get github.com/JohannesKaufmann/html-to-markdown
   ```
4. Movable Typeのエクスポートファイルを準備します。  
   ※ エクスポートファイルには必ず「TITLE」と「DATE」フィールドを含めてください。  
   ※ 「DATE」フィールドは以下のいずれかの形式で記述してください:
   - MM/DD/YYYY HH:MM:SS
   - YYYY-MM-DD HH:MM:SS
   - MM/DD/YY HH:MM:SS
   - DD/MM/YYYY HH:MM:SS
   - YYYY/MM/DD HH:MM:SS
   - MM/DD/YYYY HH:MM
5. ツールを実行します。
   ```sh
   go run main.go <Movable_Typeエクスポートファイルのパス>
   ```
6. 変換が成功すると、`output`ディレクトリ内に各記事ごとに日付形式（YYYY/MM/DD/HHMM）のディレクトリが作成され、その中に`index.md`が生成されます。

## 注意事項

- Movable Typeエクスポートファイルには必ず「TITLE」と「DATE」フィールドが含まれている必要があります。  
  「DATE」フィールドが欠けている場合や日付形式が不正な場合、変換処理はエラーとなります。
- 本ツールは記事本文のHTMLをMarkdownへ変換しますが、変換中にエラーが発生した場合は、元のHTMLがそのまま出力されます。
- ファイルやディレクトリの作成権限に問題がある場合、出力が失敗する可能性があります。

## movabletypeパッケージについて

- Movable Typeエクスポートファイルのパース処理は、`movabletype.go`に実装されています。
- 本ツールは、Movable Typeエクスポートファイルから記事情報を抽出し、`createHugoFiles`関数でHugo用のコンテンツファイルを作成します。

## デバッグ情報

- 実行中に記事の「DATE」フィールドの解析結果やエラー情報が標準出力に表示されますので、処理中のログを確認してください.
