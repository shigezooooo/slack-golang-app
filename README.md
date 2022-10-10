# slack-golang-app

## build

```
make
```

## execute, test

```
# by lambda
sam local invoke

# by API Gateway
sam local start-api
```

## deploy

```
sam deploy --guided
```



## 1. 環境構築

SAMを使ってローカル開発を進めていくには、いろいろ準備が必要です。下記のツールや実行環境をインストールします。

- <a href="https://docs.aws.amazon.com/ja_jp/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html">AWS SAM CLI</a>
- <a href="https://www.docker.com/products/docker-desktop">Docker</a>（作成したアプリをローカルで実行するのに必要）
- <a href="https://go.dev/">Go</a>（Go言語で開発しているためビルドに必要。<a href="https://github.com/syndbg/goenv">goenv</a>でインストールするのがおすすめ）

エディタは基本なんでもいいですが、拡張が豊富なvscodeが無難だと思います。



上記がインストールできたら、初めにGoのSAMサンプルプロジェクトを作成します。

```
# 対話式でサンプルプロジェクトを作成
$ sam init

# goサンプルプロジェクトの場合
$ sam init --runtime go1.x --app-template hello-world --name [プロジェクト名]
```



作成されたプロジェクトは以下のようなディレクトリ構成になっています。

```
$ tree .
.
├── Makefile
├── README.md
├── hello-world
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── main_test.go
└── template.yaml
```



上記はあくまでHelloWorldプロジェクトなのでAPI開発用に下記のように変更しました。

```
$ tree .
.
├── Makefile
├── README.md
├── get-news-list         ← [Modify] 名称を変えました
│   ├── go.mod
│   ├── go.sum
│   ├── main.go           ← lambda関数を実装
│   └── main_test.go
├── openapi.yaml          ← [New] APIをOpenApi形式で定義を記述していきます
└── template.yaml         ← SAMの設定やその他リソース情報(RoleであったりDBの定義など)をCloudFormationテンプレート形式で定義
```

`get-news-list/main.go` にアプリケーションで実行するlambda関数を記述して、`openapi.yaml`には実装するlambda関数と紐付けるAPI Gatewayの詳細な定義を**OpenAPI**（`swagger`）のお作法で定義していきます。`template.yaml`には、SAMのメインであるサーバーレスアプリケーションで管理されるリソースの定義を記述していきます。



以上で、環境構築とプロジェクトの作成は完了です！
