# FXTester

## 環境構築

### インストール
```
$ brew install go-task
```

## リリース手順

  - 1. ソースをダウンロードする
    ```
    $ git clone https://github.com/kzkapr1129/FXTester.git
    ```
  - 2. FxtesterWebを初期化し、ビルドを実行する
    ```
    $ cd FXTester/FxtesterWeb
    $ task init
    $ task init:npm
    $ npm run build
    ```
  - 3. FxtesterServerを初期化し、dockerコンテナを起動する
    ```
    $ cd ../FXTesterServer
    $ task init
    $ task dc:up
    ```