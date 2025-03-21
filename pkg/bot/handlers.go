package bot

import (
	"bytes"
	"context"
	"fmt"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"
	"math/rand"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	patternPapikRoulette = "papikRoulette"
	playersRating        = "rating"
)

var slotsResults = [7]string{
	"https://i.ibb.co/1YqJpXwW/photo-2025-03-21-18-45-11.jpg",
	"https://i.ibb.co/jPJ6TJ7Q/photo-2025-03-21-18-45-14.jpg",
	"https://i.ibb.co/Z6PhZ8jh/photo-2025-03-21-18-45-17.jpg",
	"https://i.ibb.co/qYLRLcN0/photo-2025-03-21-18-45-19.jpg",
	"https://i.ibb.co/m5Ykp15w/photo-2025-03-21-18-45-22.jpg",
	"https://i.ibb.co/pBYcBbDJ/photo-2025-03-21-18-45-25.jpg",
	"https://i.ibb.co/rRBVsQJC/photo-2025-03-21-18-45-27.jpg",
}

type BotService struct {
	embedlog.Logger
	db db.DB

	cr db.CommonRepo
}

func NewBotService(logger embedlog.Logger, dbo db.DB) *BotService {
	return &BotService{Logger: logger, db: dbo, cr: db.NewCommonRepo(dbo)}
}

func (bs *BotService) RegisterBotHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternPapikRoulette, bot.MatchTypePrefix, bs.PapikRouletteHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, playersRating, bot.MatchTypePrefix, bs.PlayersRatingHandler)
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
	username := update.InlineQuery.From.Username
	user, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &username})
	if err != nil {
		return err
	}
	if user == nil {
		newUser, err := bs.cr.AddLudoman(ctx, &db.Ludoman{
			LudomanNickname: username,
			Balance:         1000000,
		})
		if err != nil {
			return err
		}
		// send answer to the query
		_, err = b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
			//Button: &models.InlineQueryResultsButton{
			//	Text:           "Оставить фидбек",
			//	StartParameter: "1",
			//},
			InlineQueryID: update.InlineQuery.ID,
			Results: []models.InlineQueryResult{
				&models.InlineQueryResultArticle{
					ID:           "1",
					Title:        "Добро пожаловать!",
					ThumbnailURL: "https://i.ibb.co/Xfx3C5wH/image-1.jpg",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Рулетка Папикяна",
									CallbackData: patternPapikRoulette + "_" + strconv.Itoa(newUser.ID),
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nВам начислен 1.000.000 I$Coins за первый визит. Выбирайте игру и побеждайте!", username),
					}},
			},
			IsPersonal: true,
			CacheTime:  1,
		})
	} else {
		_, err = b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
			//Button: &models.InlineQueryResultsButton{
			//	Text:           "Оставить фидбек",
			//	StartParameter: "1",
			//},
			InlineQueryID: update.InlineQuery.ID,
			Results: []models.InlineQueryResult{
				&models.InlineQueryResultArticle{
					ID:           "1",
					Title:        "Выберите игру!",
					ThumbnailURL: "https://i.ibb.co/Xfx3C5wH/image-1.jpg",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Рулетка Папикяна",
									CallbackData: patternPapikRoulette + "_" + strconv.Itoa(user.ID),
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nВыбирайте игру и побеждайте!", username),
					}},
				&models.InlineQueryResultArticle{
					ID:           "2",
					Title:        "Рейтинг игроков!",
					ThumbnailURL: "https://russia-rating.ru/wp-content/uploads/2024/09/567.jpg",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Узнать рейтинг",
									CallbackData: playersRating,
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nНажмите кнопку ниже, чтобы узнать рейтинг игроков!", username),
					}},
			},
			IsPersonal: true,
			CacheTime:  1,
		})
	}

	return err
}

func (bs *BotService) PapikRouletteHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("len(parts) < 2")
		return
	}

	// find user
	userId, err := strconv.Atoi(parts[1])
	if err != nil {
		bs.Errorf("%v", err)
	}
	user, err := bs.cr.LudomanByID(ctx, userId)
	if err != nil {
		bs.Errorf("%v", err)
	}

	if user.LudomanNickname != update.CallbackQuery.From.Username {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Это не ваш автомат! Нажмите на название бота и тоже сможете сыграть :)",
			ShowAlert:       true,
		})
		return
	}

	if user.Balance < 100000 {
		_, err = b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
			InlineMessageID: update.CallbackQuery.InlineMessageID,
			Media: &models.InputMediaPhoto{
				Media:     "https://i.ibb.co/8C2G9X9/image.png",
				Caption:   "Вы израходовали свой баланс! Попробуйте в другой раз...",
				ParseMode: models.ParseModeHTML,
				//HasSpoiler: true,
			},
		})
		if err != nil {
			bs.Errorf("%v", err)
		}
		return
	}

	_, err = b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaAnimation{
			Media:     "https://media.tenor.com/_yoDqyYP8aYAAAAM/casino77-slot-machine.gif",
			Caption:   "Крутимся...",
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
	})
	if err != nil {
		bs.Errorf("%v", err)
	}

	time.Sleep(3 * time.Second)

	num := rand.Intn(len(slotsResults))
	var res string
	switch num {
	case 0:
		user.Balance += 500000
		res = fmt.Sprintf("@%s, Победа! Вы получаете +500.000 I$Coins. Ваш текущий баланс: %d I$Coins", update.CallbackQuery.From.Username, user.Balance)
	default:
		user.Balance -= 100000
		res = fmt.Sprintf("@%s, Неудача! Ваш текущий баланс: %d I$Coins", update.CallbackQuery.From.Username, user.Balance)
	}

	pic := slotsResults[num]

	_, err = b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     pic,
			Caption:   res,
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				models.InlineKeyboardButton{
					Text:         "Сыграть ещё раз",
					CallbackData: patternPapikRoulette + "_" + parts[1],
				},
			},
		}},
	})
	if err != nil {
		bs.Errorf("%v", err)
	}

	_, err = bs.cr.UpdateLudoman(ctx, user, db.WithColumns(db.Columns.Ludoman.Balance))
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
}

func (bs *BotService) PlayersRatingHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	players, err := bs.cr.LudomenByFilters(ctx, &db.LudomanSearch{}, db.Pager{PageSize: 10}, db.WithSort(db.NewSortField(db.Columns.Ludoman.Balance, true)))
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	// Шаблон для вывода списка
	listTemplate := `Список игроков:
{{- range $index, $ludoman := . }}
{{- printf "\n%d. Никнейм: @%s, Баланс: %d" (add $index 1) $ludoman.LudomanNickname $ludoman.Balance }}
{{- end }}
`
	// Функция для добавления 1 к индексу (так как индексация с 0)
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Создаем шаблон и парсим его
	tmpl, err := template.New("list").Funcs(funcMap).Parse(listTemplate)
	if err != nil {
		bs.Errorf("%v", err)
	}

	var buf bytes.Buffer

	// Выполняем шаблон и выводим результат
	err = tmpl.Execute(&buf, players)
	if err != nil {
		bs.Errorf("%v", err)
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Text:            buf.String(),
	})
	if err != nil {
		bs.Errorf("%v", err)
	}
}

func Pointer[T any](in T) *T {
	return &in
}
