package bot

import (
	"context"
	"fmt"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"
	"math"
	"math/rand"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	scholarPercent  = 100
	brokePercent    = 250
	internPercent   = 350
	juniorPercent   = 650
	middlePercent   = 850
	seniorPercent   = 950
	teamLeadPercent = 980
	ceoPercent      = 995
	papikPercent    = 1000
)

var salariesMap = map[int]string{
	scholarPercent:  `Ты школота\, копишь на обеды\, прогаешь только домашнее задание 🫵😹 Зарплату получаешь от мамы\, премию \- от бабушки\.`,
	brokePercent:    `Ты безработный 🫵😹 Стипендии и денег родителей пока что хватает на еду\, но я тебе не завидую :/`,
	internPercent:   `Тебя взяли стажёром в твою первую IT\-галеру 🔥 Теперь ты \- настоящий программист\! Правда\, придётся ишачить 2 года\, чтобы получить повышение\.\.\.`,
	juniorPercent:   `Ты получаешь гордое звание джуна\! 🧑‍💻 Таких как ты \- абсолютное большинство\. Удачи пробиться на Хедхантере :\)`,
	middlePercent:   `Ты \- мидл\! 🤠 Молодец\, немногие сюда добираются\. А теперь настало время выгорания на работе 🔥🔥🔥`,
	seniorPercent:   `Ты \- сеньор\! 🤑 Можешь работать по 3 часа в день\, хрюшам все равно дороже искать замену`,
	teamLeadPercent: `Ты \- тимлид\! 👨‍💼 Можешь вообще не работать\, а сидеть на созвонах и важных встречах целый день`,
	ceoPercent:      `Ты \- CEO\! 😎 Пока эти лошпеды тратят нервы на кодинг\, ты получаешь все сливки с их трудов\. Все потому что ты \- лучше\, чем они\. Не забывай напоминать им об этом\!`,
	papikPercent:    `Ты \- Папикян Сергей Седракович\, легенда ИТМО и самый богатый человек в мире\. Ты победил в этой жизни\, все тебе завидуют\.`,
}

type BotService struct {
	embedlog.Logger
	db db.DB
}

func NewBotService(logger embedlog.Logger, db db.DB) *BotService {
	return &BotService{Logger: logger, db: db}
}

func (bs BotService) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.InlineQuery != nil && update.InlineQuery.From != nil {
		if err := bs.answerInlineQuery(ctx, b, update); err != nil {
			bs.Errorf("%v", err)
		}
		return
	}
	return
}

func (bs BotService) answerInlineQuery(ctx context.Context, b *bot.Bot, update *models.Update) error {

	var salary int
	var ending string
	percents := rand.Intn(1000)
	switch {
	case percents <= scholarPercent:
		salary = rand.Intn(1000)
		ending = salariesMap[scholarPercent]
		break
	case percents <= brokePercent:
		salary = 10000 + rand.Intn(20000)
		ending = salariesMap[brokePercent]
		break
	case percents <= internPercent:
		salary = 20000 + (rand.Intn(25000)/1000)*1000
		ending = salariesMap[internPercent]
		break
	case percents <= juniorPercent:
		salary = 40000 + (rand.Intn(40000)/1000)*1000
		ending = salariesMap[juniorPercent]
		break
	case percents <= middlePercent:
		salary = 80000 + (rand.Intn(200000)/5000)*5000
		ending = salariesMap[middlePercent]
		break
	case percents <= seniorPercent:
		salary = 280000 + (rand.Intn(320000)/10000)*10000
		ending = salariesMap[seniorPercent]
		break
	case percents <= teamLeadPercent:
		salary = 600000 + (rand.Intn(500000)/50000)*50000
		ending = salariesMap[teamLeadPercent]
		break
	case percents <= ceoPercent:
		salary = 1100000 + (rand.Intn(10000000)/100000)*100000
		ending = salariesMap[ceoPercent]
		break
	case percents <= papikPercent:
		ending = salariesMap[papikPercent]
		salary = math.MaxInt32
		break
	}

	// send answer to the query
	results := []models.InlineQueryResult{
		&models.InlineQueryResultArticle{
			ID:           "1",
			Title:        "Твоя зп",
			ThumbnailURL: "https://cdn.vectorstock.com/i/500p/79/20/emoticon-with-dollars-vector-2287920.jpg",
			ReplyMarkup: models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						models.InlineKeyboardButton{
							Text:                         "Узнать свою",
							SwitchInlineQueryCurrentChat: " ",
						},
					},
				}},
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: fmt.Sprintf("Зарплата @%s: ||%d₽\n%s||", update.InlineQuery.From.Username, salary, ending),
				ParseMode:   models.ParseModeMarkdown,
			}},
	}

	_, err := b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		//Button: &models.InlineQueryResultsButton{
		//	Text:           "Оставить фидбек",
		//	StartParameter: "1",
		//},
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
		IsPersonal:    true,
		CacheTime:     1,
	})

	return err
}
