package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	mailgun "github.com/mailgun/mailgun-go"
	"github.com/mmcdole/gofeed"
)

func main() {
	var dir string
	flag.StringVar(&dir, "d", "./", "base directory")
	flag.Parse()
	err := godotenv.Load(fmt.Sprintf("%v/.env", dir))
	if err != nil {
		log.Fatal(err)
	}

	const dateForm = "Mon, 2 Jan 2006  03:04:05 -0700"
	t, err := readLastDate(fmt.Sprintf("%v/last_date", dir), dateForm)
	if err != nil {
		log.Fatal(err)
	}
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://lowlyj.com/feed/")
	if err != nil {
		log.Fatal(err)
	}
	newItems := []gofeed.Item{}
	for _, item := range feed.Items {
		if t.Before(*item.PublishedParsed) {
			newItems = append(newItems, *item)
		}
	}

	if len(newItems) < 1 {
		return
	}

	mg := mailgun.NewMailgun(os.Getenv("MG_DOMAIN"), os.Getenv("MG_SECRET"))

	sender := "lowly-feed@dymanticdesign.com"
	subject := fmt.Sprintf("New Lowly Updates %v/%v", time.Now().Day(), time.Now().Month())
	body := mailBody(newItems)
	recipient := "joyner.michael@gmail.com"

	message := mg.NewMessage(sender, subject, body, recipient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err = mg.Send(ctx, message)

	if err != nil {
		log.Fatal(err)
	}

	err = saveLastDate(time.Now())
	if err != nil {
		log.Fatal(err)
	}
}

func readLastDate(filepath, dateFormat string) (time.Time, error) {
	st, err := ioutil.ReadFile(filepath)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(dateFormat, string(st))
}

func saveLastDate(date time.Time) error {
	dateString := date.Format("Mon, 2 Jan 2006  03:04:05 -0700")
	return ioutil.WriteFile("last_date", []byte(dateString), 0644)
}

func mailBody(items []gofeed.Item) string {
	st := "There are some new things to read.\n\n"

	for _, item := range items {
		st = st + fmt.Sprintf("%v\n%v\n\n", item.Title, item.Link)
	}

	return st
}
