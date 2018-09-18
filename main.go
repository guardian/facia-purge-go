package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// See https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html
func main() {
	lambda.Start(handle)
}

// https://github.com/guardian/frontend-lambda/blob/master/facia-purger/src/main/scala/com/gu/purge/facia/Lambda.scala
// from what I can see it:
// - listens for S3 events and then purges Fastly accordingly
func handle(ctx context.Context, event events.S3Event) error {
	fastlyAPIKey := os.Getenv("FASTLY_API_KEY")
	fastlyServiceID := os.Getenv("FASTLY_SERVICE_ID")

	for _, record := range event.Records {
		key := record.S3.Object.Key

		if front := extractFront(key); front != "" {
			softPurge(fastlyServiceID, fastlyAPIKey, key)
		}
	}

	return nil
}

func extractFront(path string) string {
	re := regexp.MustCompile("/frontsapi/pressed/live/(.+)/fapi/pressed.v2.json")
	matches := re.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}

	log.Printf("Unable to get front from path %s", path)
	return ""
}

func softPurge(serviceID string, apiKey string, contentID string) bool {
	contentPath := "/" + contentID
	surrogateKey := md5.Sum([]byte(contentPath))

	url := fmt.Sprintf("https://api.fastly.com/service/%s/purge/%s", serviceID, surrogateKey)

	client := http.Client{}
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Add("FASTLY_KEY", apiKey)
	req.Header.Add("Fastly-Soft-Purge", "1")
	_, err := client.Do(req)

	if err != nil {
		log.Printf("Fastly purge request for %s failed, %s", contentID, err.Error())
		return false
	}

	return true
}
