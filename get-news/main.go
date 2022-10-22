package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/mmcdole/gofeed"
	"github.com/slack-go/slack"
)

/// リクエスト
type RequestBody struct {
	Challenge string `json:"challenge"`
	Token     string `json:"token"`
	Type      string `json:"type"`
	Event     Event  `json:"event"`
}

type Event struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

/// レスポンス
type ChallengeResponseBody struct {
	Challenge string `json:"challenge"`
}

type ResponseBody struct {
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

type NewsItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type NewsItems []NewsItem

/// lambdaハンドラー
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	requestBody, err := getRequestBody(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	event := requestBody.Event
	challenge := requestBody.Challenge

	// challenge認証
	if challenge != "" {
		responseBody, err := json.Marshal(ChallengeResponseBody{
			Challenge: challenge,
		})

		_ = err
		return events.APIGatewayProxyResponse{
			Body:       string(responseBody),
			StatusCode: 200,
		}, nil
	}

	if request.Headers["X-Slack-Signature"] != "" {
		// slackからの投稿の場合、SlackBotに取得結果をPOSTする

		channel := event.Channel
		textArray := strings.Split(event.Text, "\n")

		log.Print(textArray)
		if len(textArray) < 2 {
			// パラメータが不足
			sendSlackMessage(channel, "検索ワードを検出できませんでした")

			responseBody, err := json.Marshal(ResponseBody{
				Status: 1,
				Detail: "NG",
			})

			_ = err
			return events.APIGatewayProxyResponse{
				Body:       string(responseBody),
				StatusCode: 200,
			}, nil
		}

		query := textArray[1]

		// RSSフィードでニュースを取得
		log.Print("検索ワード：" + query)
		urlText := "https://news.google.com/rss/search?hl=ja&gl=JP&ceid=JP:ja&q=" + url.QueryEscape(query)
		fp := gofeed.NewParser()
		feed, _ := fp.ParseURL(urlText)

		message := query + " に関するニュースをお届けします! \n\n" + feed.Items[0].Title + " \n" + feed.Items[0].Link + " \n\n"

		sendSlackMessage(channel, message)
	}

	responseBody, err := json.Marshal(ResponseBody{
		Status: 0,
		Detail: "OK",
	})

	return events.APIGatewayProxyResponse{
		Body:       string(responseBody),
		StatusCode: 200,
	}, nil
}

/// Slackへメッセージを投稿する
func sendSlackMessage(channel, text string) {
	// パラメータストアより認証キーを取得
	res, err := fetchParameterStore(getEnv("SSM_SLACK_AUTH_KEY_NAME", ""))
	if err != nil {
		log.Print(err)
		return
	}

	client := slack.New(res)
	_, _, err = client.PostMessage(channel, slack.MsgOptionText(text, true))
	if err != nil {
		// エラーログを出力
		log.Print(err.Error())
	}
}

/// APIGatewayからのリクエストBodyを処理用に変換する
func getRequestBody(bodyText string) (*RequestBody, error) {
	var body RequestBody
	err := json.Unmarshal([]byte(bodyText), &body)
	if err != nil {
		return nil, err
	}
	return &body, nil
}

/// パラメータストアから値を取得する
func fetchParameterStore(paramName string) (string, error) {

	sess := session.Must(session.NewSession())
	svc := ssm.New(
		sess,
		aws.NewConfig().WithRegion(getEnv("SSM_REGION", "")),
	)

	res, err := svc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "SSM: パラメータ取得失敗", err
	}

	value := *res.Parameter.Value
	return value, nil
}

/// Lambdaに設定した環境変数を参照する
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	lambda.Start(handler)
}
