package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	token, textPath, err := parseFlags()
	if err != nil {
		fmt.Printf("parse flags error: %s\n", err)
		os.Exit(1)
	}

	textFile, err := os.Open(textPath)
	if err != nil {
		fmt.Printf("open comment text file error: %s\n", err)
		os.Exit(1)
	}
	defer textFile.Close()

	builder := new(strings.Builder)
	if _, err := io.Copy(builder, textFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	commentText := builder.String()

	client := NewClient(token)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("Fetching promo timeline...")

			timeline, err := client.SelfPromoTimeline()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			nowDt := time.Now()
			for _, item := range timeline.Result.Items {
				itemDt := time.Unix(item.Data.Date, 0)

				row := fmt.Sprintf("title=\"%s\" date=%s", item.Data.Title, itemDt)

				if nowDt.Year() != itemDt.Year() || nowDt.YearDay() != itemDt.YearDay() {
					fmt.Printf("%s: wrong date\n", row)
					continue
				}

				if item.Data.Title != "Субботний самопиар на vc.ru" {
					fmt.Printf("%s: title is not \"Субботний самопиар на vc.ru\"\n", row)
					continue
				}

				fmt.Printf("%s: matched.\n", row)

				if err := client.AddComment(item.Data.ID, commentText); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				fmt.Println("Comment was posted.")

				return
			}
		}
	}
}

func parseFlags() (token string, textPath string, err error) {
	flag.StringVar(&token, "token", "", "vc.ru API token")
	flag.StringVar(&textPath, "text-path", "", "comment text file path")

	flag.Parse()

	if token == "" {
		err = errors.New("token not passed")
		return
	}

	if textPath == "" {
		err = errors.New("comment text file path not passed")
		return
	}

	return
}
