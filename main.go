package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	inow   time.Time
	now    time.Time
	cday   int
	cmonth time.Month
)

type centers struct {
	Centers []struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		Pincode  int    `json:"pincode"`
		Sessions []struct {
			Date              string `json:"date"`
			AvailableCapacity int    `json:"available_capacity"`
			MinAgeLimit       int    `json:"min_age_limit"`
			Vaccine           string `json:"vaccine"`
			AvailableCapDose1 int    `json:"available_capacity_dose1"`
			AvailableCapDose2 int    `json:"available_capacity_dose2"`
		} `json:"sessions"`
	} `json:"centers"`
}

func main() {
	// Initialize chat bot
	telegramBotToken := os.Getenv("BOT_TOKEN")
	cowinUrl := os.Getenv("COWIN_URL")
	c, _ := strconv.Atoi(os.Getenv("BOT_CHAT_ID"))
	chatId := int64(c)
	fmt.Println(chatId)
	if telegramBotToken == "" || cowinUrl == "" {
		fmt.Println("cannot read url/token")
		os.Exit(0)
	}

	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		time.Sleep(3 * time.Second)
		fmt.Println(err.Error())
		return
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	i := 0 // counter that helps incrementing date

	for {
		d, m := getIterationDayAndMonth(i)

		client := &http.Client{}
		url := fmt.Sprintf(cowinUrl, d, m)
		//fmt.Println(url)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Test")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(3 * time.Second)
			fmt.Println(err.Error())
			continue
		}

		found := false
		c := &centers{}
		body, _ := ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(body, c)

		for _, ctr := range c.Centers {
			for _, sess := range ctr.Sessions {
				if sess.MinAgeLimit >= 18 && sess.AvailableCapacity > 0 {
					fmt.Println("Slot Found!!\n", ctr.Name, ctr.Address, ctr.Pincode)
					fmt.Printf("day=%v month=%v", d, m)
					found = true

					// Send notification to telegram
					text := fmt.Sprintf("Slot Found 18+!!\nName: %v\nAddress: %v\nPincode: %v\nDate: %v\n"+
						"Total Available Shots: %v \nDose 1: %v\nDose 2: %v\nVaccine: %v", ctr.Name, ctr.Address, ctr.Pincode,
						sess.Date, sess.AvailableCapacity, sess.AvailableCapDose1, sess.AvailableCapDose2, sess.Vaccine)
					msg := tgbotapi.NewMessage(chatId, text)
					bot.Send(msg)

					break
				}
			}
			if found {
				break
			}
		}

		if found {
			i++
			fmt.Println("Slot Found")
			time.Sleep(3 * time.Second)
			continue
		}

		i++
		fmt.Println("Slot Not Found")
		time.Sleep(5 * time.Second)
	}
}

func getIterationDayAndMonth(i int) (d, m string) {
	if i%19 == 0 {
		loc, _ := time.LoadLocation("Asia/Kolkata")
		now = time.Now().In(loc)
		inow = time.Now().In(loc)
		cday = now.Day()
		cmonth = now.Month()
		d = strconv.Itoa(cday)
		m = strconv.Itoa(int(cmonth))
	} else {
		inow = inow.Add(24 * time.Hour)
		cday = inow.Day()
		cmonth = inow.Month()
		d = strconv.Itoa(cday)
		m = strconv.Itoa(int(cmonth))
	}

	if cday/10 == 0 && d[0] != 0 {
		d = "0" + d
	}

	if cmonth/10 == 0 && m[0] != 0 {
		m = "0" + m
	}

	return d, m
}
