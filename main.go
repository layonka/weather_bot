package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type WeatherResp struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
	} `json:"main"`

	Weather []struct {
		Desc string `json:"description"`
	} `json:"weather"`

	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`

	Name string `json:"name"`
}

func getWeather(city string, apikey string) (string, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric&lang=ru", city, apikey)

	req, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	var data WeatherResp
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
	}

	if data.Name == "" {
		return "Введите название городка", nil
	}

	temp := data.Main.Temp
	desc := data.Weather[0].Desc
	name := data.Name
	feelsLike := data.Main.FeelsLike
	wind := data.Wind.Speed

	return fmt.Sprintf("Город: %s\nТемпература: %.1f°C\nОписание: %s\nОщущается как: %.1f°C\nСкорость ветра: %.1f м/с", name, temp, desc, feelsLike, wind), nil
}

func main() {
	err := godotenv.Load("file.env")
	if err != nil {
		log.Println("Файл .env не загружен:", err)
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	apikey := os.Getenv("WEATHER_API_KEY")

	if token == "" || apikey == "" {
		log.Fatal(".env файлы не найдены")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Бот авторизован")

	bot.Debug = true

	updateCfg := tgbotapi.NewUpdate(0)
	updateCfg.Timeout = 60

	upd := bot.GetUpdatesChan(updateCfg)

	for updates := range upd {
		if updates.Message.Text == "" {
			continue
		}

		if updates.Message == nil {
			continue
		}

		if updates.Message.IsCommand() && updates.Message.Command() == "start" {
			message := tgbotapi.NewMessage(updates.Message.Chat.ID, "💋 Напишите название города. Покажу погоду в нём.")
			bot.Send(message)
			continue
		}

		city := updates.Message.Text

		weather, err := getWeather(city, apikey)
		if err != nil {
			message := tgbotapi.NewMessage(updates.Message.Chat.ID, "Либо город неправильный, либо пробуй позже, либо с сайтом беда.")
			bot.Send(message)
			continue
		}

		message := tgbotapi.NewMessage(updates.Message.Chat.ID, weather)
		bot.Send(message)
	}
}

//hello
