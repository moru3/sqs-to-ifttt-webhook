package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const settingFile = "./setting.yml"

// Settings from yaml
type Settings struct {
	AwsRegion string `yaml:"awsRegion"`
	QueueURL  string `yaml:"queueURL"`
	IftttURL  string `yaml:"iftttURL"`
	Interval  int    `yaml:"interval"`
}

var settings Settings
var svc *sqs.SQS

func main() {
	// load setting file
	buf, err := ioutil.ReadFile(settingFile)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(buf, &settings)
	if err != nil {
		panic(err)
	}

	// set up SQS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(settings.AwsRegion)},
	)

	if err != nil {
		log.Fatal(err)
	}
	svc = sqs.New(sess)

	log.Println("start")

	for {
		retrieveMessage()
		time.Sleep(time.Duration(settings.Interval) * time.Second)
	}
}

func retrieveMessage() error {
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(settings.QueueURL),
		MaxNumberOfMessages: aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(20),
	}
	resp, err := svc.ReceiveMessage(params)

	if err != nil {
		return err
	}

	log.Printf("messages count: %d", len(resp.Messages))

	if len(resp.Messages) == 0 {
		log.Println("empty queue")
		return nil
	}

	var wg sync.WaitGroup
	for _, m := range resp.Messages {
		wg.Add(1)
		go func(msg *sqs.Message) {
			defer wg.Done()
			log.Println(msg.Body)
			httpPost(settings.IftttURL, *msg.Body)
			if err := deleteMessage(msg); err != nil {
				log.Println(err)
			}
		}(m)
	}

	wg.Wait()

	return nil
}

// delete messages
func deleteMessage(msg *sqs.Message) error {
	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(settings.QueueURL),
		ReceiptHandle: aws.String(*msg.ReceiptHandle),
	}
	_, err := svc.DeleteMessage(params)

	if err != nil {
		return err
	}
	return nil
}

func httpPost(url, value1 string) error {
	jsonStr := `{"value1":"` + value1 + `"}`

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}
