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
