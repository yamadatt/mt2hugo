# mt2hugo
movable typeのエクスポートファイルからhugoのファイルを作成する

## 使用方法

1. Golangをインストールします。
2. このリポジトリをクローンします。
3. `main.go`を実行して、movable typeのエクスポートファイルをhugoのファイルに変換します。

```sh
go run main.go <path_to_movable_type_export_file>
```

4. 変換されたhugoファイルは、各記事ごとにディレクトリが作成され、その中に保存されます。

## 注意事項

- Movable Typeエクスポートファイルには必ず「TITLE」フィールドが含まれている必要があります。
- 「TITLE」フィールドが欠けている場合、エラーが発生します。
- Movable Typeエクスポートファイルに空行が含まれている場合、それらは無視されます。

## パーサーパッケージ

- Movable Typeエクスポートファイルのパース処理は、`parser`パッケージに分離されています。
- `parser`パッケージは、`parser/parser.go`ファイルに実装されています。
