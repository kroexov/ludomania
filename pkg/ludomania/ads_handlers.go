package ludomania

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"math/rand"
	"strconv"
	"strings"
)

const (
	patternAddWatch = "add"
	// limit to watch ads
	adWatchLimit = 10
)

var ads = []string{
	`||Лучший мини апп [полка](https://t.me/polkabot_news) зарелизился и ждет клиентов\. Будьте первыми\!||`,
	`||Выберите себе фильм или сериал посмотреть с [PopcornBro](https://t.me/PopcornBroBot)\!
Алгоритм подстроится под ваши желания и посоветует самые лушие опции\!||`,
	`||Узнайте свой грейд с @tvoy\_grade\_bot \!
Просто введите название бота в чате и лудоманьте свою зарплату\!
В качестве бонуса вы получите возможность накрутить себе песню undefined\!||`,
	`||Эй\, джун\, а ты уже улучшил своё резюме с [resumeup](https://resumeup.ru/consultancy)\? Заходи и почитай\, как его прокачать\!||`,
	`||Подписывайся на [лучший канал про бэкенд](https://t.me/andrey_threads) от дядюшки Андрея\!||`,
	`||А вы знали\, что у мейнтейнера бота есть свой [канал в тг](https://t.me/+RPsESmkpZFY3OTIy)\? Подпишись и читай про ИТМО и разработку с точки зрения первокурсника\!||`,
	`||Если ищешь работу\, чекай [ITMO Careers](https://t.me/careercentreitmo)\! Подпишись и смотри топ вакансии по своему профилю\!||`,
	`||Самые вкусные вакансии \- в [StudUp Jobs](https://t.me/studup_jobs)\! Не упусти свой шанс залутать первый опыт коммерческой разработки\!||`,
	`||Советую классный сайт \- [Easyoffer](https://easyoffer.ru/)\! Там можно чекнуть записи собесов\, топ вопросы по вакансиям\, и другие штуки для ускорения поиска работы\!||`,
}

func (bs *BotService) rollAd(userId int) models.InlineQueryResult {
	text := `Рекламная интеграция\!
` + ads[rand.Intn(len(ads))]
	if rand.Intn(11) == 10 {
		text = fmt.Sprintf(`Рекламная интеграция\!
||А _ТЫ_ уже поставил звёздочку на наш [репозиторий](https://github.com/kroexov/ludomania/)\? На данный момент там всего %d звёздочек\!||`, bs.limitByBack-defaultLimitBuyBack)
	}
	return &models.InlineQueryResultGif{
		ID:                "6",
		Title:             "Реклама!",
		Caption:           "Реклама!",
		GifURL:            "https://media.tenor.com/QttOudwaS4kAAAAM/ohhp.gif",
		ThumbnailURL:      "https://media.tenor.com/QttOudwaS4kAAAAM/ohhp.gif",
		ThumbnailMimeType: "image/gif",
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					models.InlineKeyboardButton{
						Text:         "Получить 500К за просмотр рекламы",
						CallbackData: patternAddWatch + "_" + strconv.Itoa(userId),
					},
				},
			}},
		InputMessageContent: &models.InputTextMessageContent{
			MessageText: text,
			ParseMode:   models.ParseModeMarkdown,
		}}
}

func (bs *BotService) AddWatch(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("len(parts) < 2")
		return
	}
	userId, err := strconv.Atoi(parts[1])
	if err != nil {
		bs.Errorf("%v", err)
	}

	ludoman, err := bs.cr.LudomanByID(ctx, userId)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	if ludoman.AdsWatched > adWatchLimit {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Вы уже посмотрели слишком много рекламы на сегодня, попробуйте завтра :)",
			ShowAlert:       true,
		})
		if err != nil {
			bs.Errorf("%v", err)
		}
		return
	}

	ludoman.AdsWatched += 1

	if _, err = bs.cr.UpdateLudoman(ctx, ludoman); err != nil {
		bs.Errorf("%v", err)
		return
	}

	if err = bs.updateBalance(500000, []int{userId}, true); err != nil {
		bs.Errorf("%v", err)
		return
	}

	_, err = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		ReplyMarkup:     nil,
	})
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
}
