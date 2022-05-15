package main

import (
	"os"
	"strings"
	"strconv"
	"log"
	"net/url"
	"encoding/json"
	"github.com/mmcdole/gofeed"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/slack-go/slack"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

/// リクエスト
type RequestBody struct {
	Challenge string `json:"challenge"`
	Token string `json:"token"`
	Type string `json:"type"`
	Event Event `json:"event"`
}

type Event struct {
	Text string `json:"text"`
	Channel string `json:"channel"`
}

/// レスポンス
type ChallengeResponseBody struct {
	Challenge string `json:"challenge"`
}

type ResponseBody struct {
	Status int `json:"status"`
	Detail string `json:"detail"`
}

type NewsItem struct {
	Title  string `json:"title"`
	Link   string `json:"link"`
}

type NewsItems []NewsItem

/// lambdaハンドラー
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// リクエストBody 取得
	requestBody, err := getRequestBody(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: err.Error(),
			StatusCode: 500,
		}, err
	}

	// challenge認証
	if requestBody.Challenge != "" {
		responseBody, err := json.Marshal(ChallengeResponseBody{
			Challenge: requestBody.Challenge,
		})

		_ = err
		return events.APIGatewayProxyResponse{
			Body: string(responseBody),
			StatusCode: 200,
		}, nil
	}

	if request.Headers["X-Slack-Signature"] != "" {
		// slackからの投稿の場合、SlackBotに取得結果をPOSTする
		postChannel(requestBody)
	}

	// レスポンスBody 生成
	responseBody, err := json.Marshal(ResponseBody{
		Status: 0,
		Detail: "正常終了",
	})

	return events.APIGatewayProxyResponse{
		Body: string(responseBody),
		StatusCode: 200,
	}, nil
}

func postChannel(body *RequestBody) {
	log.Print(body.Event.Text)
	
	// POST値チェック
	channel := body.Event.Channel
	textArray := strings.Split(body.Event.Text, "\n")
	limitStr := ""
	log.Print(textArray)
	if len(textArray) >= 3 {
		// 正常
		limitStr = textArray[2]
	} else if len(textArray) == 2 {
		// 正常（limitにデフォルト値を設定）
		limitStr = getEnv("DEFAULT_LIMIT_CNT", "1")
	} else {
		// 異常
		sendSlackMessage(channel, "検索ワードを入力してね")
		return
	} 

	query := textArray[1]
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		// 変換に失敗
		log.Print(err.Error())
		sendSlackMessage(channel, "取得件数は1以上の整数で入力してね")
		return
	}

	log.Print(query)
	log.Print(limit)

	// RSSフィードを取得
	urlText := "https://news.google.com/rss/search?hl=ja&gl=JP&ceid=JP:ja&q=" + url.QueryEscape(query)
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(urlText)

	text := query + " に関するニュースをお届けします! \n\n"
	for cnt, item := range feed.Items {
		text += item.Title + " \n" + item.Link + " \n\n"

		if cnt + 1 == limit {
			break
		}
	}

	// slackへメッセージをPOST
	sendSlackMessage(channel, text)
}

func sendSlackMessage(channel, text string) {
	log.Print(text)

	res, err := fetchParameterStore("SlackGolangApp-SlackAccessToken")
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getRequestBody(bodyText string) (*RequestBody, error) {
	var body RequestBody
	err := json.Unmarshal([]byte(bodyText), &body)
	if err != nil {
		return nil, err
	}
	return &body, nil
}

func fetchParameterStore(paramName string) (string, error) {

	sess := session.Must(session.NewSession())
	svc := ssm.New(
		sess,
		aws.NewConfig().WithRegion(getEnv("SSM_REGION", "")),
	)

	res, err := svc.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "SSM：パラメータ取得失敗", err
	}

	value := *res.Parameter.Value
	return value, nil
}

func main() {
	lambda.Start(handler)
}