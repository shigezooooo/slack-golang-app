# slack-golang-app

## about project

goのサンプルプロジェクト(`sam init --runtime go1.x --app-template hello-world`)をベースに作成。
フォルダ構成を下記のようにいじりました

```
$ tree .
.
├── Makefile
├── README.md
├── get-news              ← [Modify] 名称を変えました
│   ├── go.mod
│   ├── go.sum
│   ├── main.go           ← lambda関数を実装
├── openapi.yaml          ← [New] lambda関数をAPIとして利用するためにAPIGatewayと紐づける。OpenApi形式で定義を記述。
└── template.yaml         ← SAMの設定やその他リソース情報(RoleであったりLogの定義など)をCloudFormationテンプレート形式で定義
```


## setup

SAMを使ってローカル開発を進めていくには準備が必要です。下記ツールをインストールします。

- <a href="https://docs.aws.amazon.com/ja_jp/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html" target="_blank">AWS SAM CLI</a>
- <a href="https://www.docker.com/products/docker-desktop" target="_blank">Docker</a>（作成したアプリをローカルで実行するのに必要）
- <a href="https://go.dev/" target="_blank">Go</a>（Go言語で開発しているためビルドに必要。<a href="https://go.dev/dl/" target="_blank">公式</a>からダウンロードしましょう。）


## build

```
sam build
```

## execute, test

```
# By lambda
sam local invoke

# By API Gateway
sam local start-api
```

## deploy

```
sam deploy --guided
```

