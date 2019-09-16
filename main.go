package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/whatupdave/mcping"
)

var buildDate string

func main() {
	token := os.Getenv("TG_TOKEN")
	if token == "" {
		log.Panic("No token")
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		log.Panic("No host")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panicf("Can't create bot: %v", err)
	}

	if os.Getenv("DEBUG") == "1" {
		bot.Debug = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	build, err := strconv.ParseInt(buildDate, 10, 64)
	var buildString string
	if err != nil {
		log.Println(err)
		buildString = "I don't know️, but this is probably a debug build of some kind."
	} else {
		buildString = time.Unix(build, 0).Format("Mon Jan 2 15:04:05 MST 2006")
	}

	queryResultTitle := "SERVER STATS"

	me, err := bot.GetMe()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("I am %s\n", me.String())

	for update := range updates {
		if update.InlineQuery != nil {
			stats, err := queryServer(serverHost)
			if err != nil {
				log.Println(err)
				continue
			}

			t := strconv.FormatInt(time.Now().Unix(), 10)
			if bot.Debug {
				stats += fmt.Sprintf("\nTIME: %s", t)
			}

			answer := tgbotapi.NewInlineQueryResultArticle(t, queryResultTitle, stats)

			bot.AnswerInlineQuery(tgbotapi.InlineConfig{
				InlineQueryID:     update.InlineQuery.ID,
				Results:           []interface{}{answer},
				CacheTime:         0,
				IsPersonal:        false,
				NextOffset:        "",
				SwitchPMText:      "",
				SwitchPMParameter: "",
			})

			continue
		}

		if update.Message != nil {
			stats, err := queryServer(serverHost)
			if err != nil {
				log.Println(err)
				continue
			}

			cmd := update.Message.Command()
			switch cmd {
			case "mc":
				answer := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					stats,
				)

				bot.Send(answer)
				continue
			case "version":
				answer := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Build date: "+buildString,
				)

				bot.Send(answer)
				continue
			}
		}
	}
}

func queryServer(host string) (string, error) {
	res, err := mcping.Ping(host)
	if err != nil {
		return "STATUS: Offline", err
	}

	var players []string
	for _, v := range res.Sample {
		players = append(players, v.Name)
	}

	message := fmt.Sprintf("STATUS: Online\nPLAYERS: %s", strings.Join(players, ", "))

	return message, nil
}
