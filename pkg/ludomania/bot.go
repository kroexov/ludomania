package ludomania

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-pg/pg/v10"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	patternConfirm             = "confirm:"
	patternPapikSlots          = "papikSlots"
	patternMayatinRoulette     = "mayatinRoulette"
	patternMayatinRouletteBet  = "mayatinBet"
	patternPovyshevExams       = "povyshevExams"
	patternBuyBack             = "buyback"
	playersRating              = "rating"
	patternBuyBackHouse        = "BuyBackHouse"
	initialBalance             = 1000000
	patternMayatinRouletteBetN = "_n"
	patternMayatinRouletteBetP = "_p"
	patternMayatinRouletteBetB = "_b"
	patternMayatinRouletteBetU = "_u"

	patternAddWatch = "add"
)

var p = message.NewPrinter(language.German)

var slotsResults = [7]string{
	"https://i.ibb.co/1YqJpXwW/photo-2025-03-21-18-45-11.jpg",
	"https://i.ibb.co/jPJ6TJ7Q/photo-2025-03-21-18-45-14.jpg",
	"https://i.ibb.co/Z6PhZ8jh/photo-2025-03-21-18-45-17.jpg",
	"https://i.ibb.co/qYLRLcN0/photo-2025-03-21-18-45-19.jpg",
	"https://i.ibb.co/m5Ykp15w/photo-2025-03-21-18-45-22.jpg",
	"https://i.ibb.co/pBYcBbDJ/photo-2025-03-21-18-45-25.jpg",
	"https://i.ibb.co/rRBVsQJC/photo-2025-03-21-18-45-27.jpg",
}

var marketingSlots = []string{
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
	`||Самые вкусные вакансии - в [StudUp Jobs](https://t.me/studup_jobs)\! Не упусти свой шанс залутать первый опыт коммерческой разработки\!||`,
}

var jackPotPapikyan = "https://i.ibb.co/3yPD09VM/image.png"
var jackPotITMO = "https://i.ibb.co/QWq12Yq/image.png"

type MayatinRouletteCategory struct {
	CategoryName string
	CategoryPic  string
	WinSum       int
}

var mayatinCategories = map[string]MayatinRouletteCategory{
	patternMayatinRouletteBetN: {
		CategoryName: "Надежность",
		CategoryPic:  "https://i.ibb.co/mCxMpSdk/image.png",
		WinSum:       1500000,
	},
	patternMayatinRouletteBetP: {
		CategoryName: "Производительность",
		CategoryPic:  "https://i.ibb.co/Zpqh23VB/image.png",
		WinSum:       1500000,
	},
	patternMayatinRouletteBetB: {
		CategoryName: "Безопасность",
		CategoryPic:  "https://i.ibb.co/WNbKBsrp/image.png",
		WinSum:       1500000,
	},
	patternMayatinRouletteBetU: {
		CategoryName: "Уважаемый коллега",
		CategoryPic:  "https://i.ibb.co/DPjH6ym5/image.png",
		WinSum:       5000000,
	},
}

type BotService struct {
	embedlog.Logger
	db db.DB

	cr                      db.CommonRepo
	mayatinRouletteBets     *sync.Map
	mu                      sync.Mutex
	isMayatinRouletteActive bool
	mayatinRouletteUsers    map[int]struct{}
	mayatinCategoriesVotes  map[string]int

	limitByBack    int
	papikyanLock   map[int]struct{}
	lastClick      sync.Map
	blackjackGames *sync.Map
	buyBackLock    map[int]struct{}
}

func (bs *BotService) SetLimitByBack(newLimit int) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.limitByBack = newLimit
	bs.Logger.Printf("New limit : %d", bs.limitByBack)
}
func NewBotService(logger embedlog.Logger, dbo db.DB) *BotService {
	return &BotService{Logger: logger, db: dbo, cr: db.NewCommonRepo(dbo), mayatinRouletteBets: new(sync.Map), papikyanLock: make(map[int]struct{}), buyBackLock: make(map[int]struct{}), limitByBack: 10, blackjackGames: new(sync.Map)}

}

func (bs *BotService) RegisterBotHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternPapikSlots, bot.MatchTypePrefix, bs.PapikRouletteHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternMayatinRoulette, bot.MatchTypePrefix, bs.MayatinRouletteHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternMayatinRouletteBet, bot.MatchTypePrefix, bs.MayatinRouletteBetHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, playersRating, bot.MatchTypePrefix, bs.PlayersRatingHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternBuyBack, bot.MatchTypePrefix, bs.BuyBackHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternBuyBackHouse, bot.MatchTypePrefix, bs.BuybackHouseHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternConfirm, bot.MatchTypePrefix, bs.handleCallbackQueryTransaction)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternBlackjack, bot.MatchTypePrefix, bs.BlackjackHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternAddWatch, bot.MatchTypePrefix, bs.AddWatch)
}

func (bs *BotService) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil && update.Message.Document != nil {
		println(update.Message.Document.FileName, "|", update.Message.Document.FileID)
	}
	if update.Message != nil && update.Message.ViaBot != nil && update.Message.Chat.Type == models.ChatTypeSupergroup && update.Message.ViaBot.ID == 7672429736 && update.Message.MessageThreadID != 8388 {
		_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: update.Message.Chat.ID, MessageID: update.Message.ID})
		if err != nil {
			bs.Errorf("%v", err)
			return
		}
	}
	if update.InlineQuery != nil && update.InlineQuery.From != nil {
		if err := bs.answerInlineQuery(ctx, b, update); err != nil {
			bs.Errorf("%v", err)
		}
		return
	}
	return
}
func (bs *BotService) Transaction(ctx context.Context, userFrom db.Ludoman, userTo db.Ludoman, amount int, dbo db.DB) error {
	if userFrom.Balance < amount {
		return fmt.Errorf("недостаточно средств: нужно %d, а есть %d", amount, userFrom.Balance)
	}

	err := dbo.RunInTransaction(ctx, func(tx *pg.Tx) error {
		query1 := `UPDATE ludomans SET balance = balance -(?0) WHERE "ludomanId" = ?1`
		if _, err := tx.Exec(query1, amount, userFrom.ID); err != nil {
			return err
		}
		query2 := `UPDATE ludomans SET balance = balance +(?0) WHERE "ludomanId" = ?1`
		if _, err := tx.Exec(query2, amount, userTo.ID); err != nil {
			return err
		}

		txRepo := bs.cr.WithTransaction(tx)

		_, err := txRepo.AddTransaction(ctx, &db.Transaction{
			FromLudomanID: userFrom.ID,
			ToLudomanID:   userTo.ID,
			Amount:        amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("ошибка транзакции: %w", err)
	}

	return nil
}

func (bs *BotService) isUserFromBot(ctx context.Context, nickname string) bool {
	search := &db.LudomanSearch{LudomanNickname: &nickname}
	user, err := bs.cr.OneLudoman(ctx, search)
	return err == nil && user != nil
}
func (bs *BotService) transferInlineQuery(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	if update.InlineQuery == nil {
		return false
	}

	userInput := update.InlineQuery.Query
	parts := strings.SplitN(userInput, " ", 2)

	if len(parts) != 2 {
		return false
	}

	firstPart := parts[0]
	if len(firstPart) > 1 {
		firstPart = firstPart[1:]
	}

	secondPart := parts[1]
	value, err := strconv.Atoi(secondPart)
	if err != nil {
		fmt.Printf("Ошибка преобразования строки в число: %v\n", err)
		return false
	}
	fmt.Println("value =", value)

	if value >= 100000 && bs.isUserFromBot(ctx, firstPart) {
		username := update.InlineQuery.From.Username

		userFrom, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &username})
		if err != nil || userFrom == nil {
			fmt.Println("Юзера-отправителя не существует или ошибка БД")
			return false
		}

		fmt.Println("баланс и value =", userFrom.Balance, value)
		if userFrom.Balance >= value {
			keyboard := &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{{
					{
						Text:         fmt.Sprintf("Подтвердить перевод %d для %s", value, firstPart),
						CallbackData: fmt.Sprintf("confirm:%s:%s:%d", username, firstPart, value),
					},
				}},
			}

			result := &models.InlineQueryResultArticle{
				ID:    "1",
				Title: "Подтвердите перевод",
				InputMessageContent: &models.InputTextMessageContent{
					MessageText: fmt.Sprintf("Перевести %d пользователю %s?", value, firstPart),
				},
				ReplyMarkup: keyboard,
			}

			b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
				InlineQueryID: update.InlineQuery.ID,
				Results:       []models.InlineQueryResult{result},
			})

		}
	}

	return true
}

func (bs *BotService) handleCallbackQueryTransaction(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	data := update.CallbackQuery.Data

	parts := strings.SplitN(data, ":", 4)
	if len(parts) != 4 {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Неверный формат подтверждения.",
			ShowAlert:       true,
		})
		bs.deleteCallbackMessage(ctx, b, update)
		return
	}
	initiatorNick := parts[1]
	targetNick := parts[2]
	value, err := strconv.Atoi(parts[3])

	clickerNick := update.CallbackQuery.From.Username

	if clickerNick != initiatorNick {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Это не ваш автомат! Только @" + initiatorNick + " может подтвердить перевод.",
			ShowAlert:       true,
		})
		return
	}

	if err != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Некорректная сумма.",
			ShowAlert:       true,
		})
		bs.deleteCallbackMessage(ctx, b, update)
		return
	}
	fromUsername := update.CallbackQuery.From.Username
	userFrom, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &fromUsername})
	if err != nil || userFrom == nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка поиска отправителя.",
			ShowAlert:       true,
		})
		bs.deleteCallbackMessage(ctx, b, update)
		return
	}
	if userFrom.LudomanNickname != update.CallbackQuery.From.Username {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Это не ваш автомат! Нажмите на название бота и тоже сможете сыграть :)",
			ShowAlert:       true,
		})
		return
	}

	userTo, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &targetNick})
	if err != nil || userTo == nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Получатель не найден.",
			ShowAlert:       true,
		})
		bs.deleteCallbackMessage(ctx, b, update)
		return
	}
	if userFrom.Balance < value {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Недостаточно средств.",
			ShowAlert:       true,
		})
		bs.deleteCallbackMessage(ctx, b, update)
		return
	}
	err = bs.Transaction(ctx, *userFrom, *userTo, value, bs.db)
	if err != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка транзакции.",
			ShowAlert:       true,
		})
		bs.deleteCallbackMessage(ctx, b, update)
		return
	}
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Платеж успешно выполнен.",
		ShowAlert:       true,
	})

	if update.CallbackQuery.InlineMessageID != "" {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			InlineMessageID: update.CallbackQuery.InlineMessageID,
			Text:            fmt.Sprintf("Пользователь @%s успешно перевел %d I$ coins пользователю @%s", fromUsername, value, targetNick),
		})
	}
}

func (bs *BotService) deleteCallbackMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery.Message.Message != nil {
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		messageID := update.CallbackQuery.Message.Message.ID

		if _, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: messageID,
		}); err != nil {
			bs.Errorf("не удалось удалить сообщение: %v", err)
		}
		return
	}

	if update.CallbackQuery.InlineMessageID != "" {
		if _, err := b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
			InlineMessageID: update.CallbackQuery.InlineMessageID,
			ReplyMarkup:     nil,
		}); err != nil {
			bs.Errorf("не удалось удалить сообщение InlineMessage : %v", err)
		}
	}
}

func (bs *BotService) answerInlineQuery(ctx context.Context, b *bot.Bot, update *models.Update) error {
	if bs.transferInlineQuery(ctx, b, update) {
		return nil
	}
	username := update.InlineQuery.From.Username
	tgID := int(update.InlineQuery.From.ID)
	user, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &username})
	if err != nil {
		return err
	}
	if user == nil {
		newUser, err := bs.cr.AddLudoman(ctx, &db.Ludoman{
			LudomanNickname: username,
			Balance:         initialBalance,
			TgID:            tgID,
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
									Text:         "Слоты Папикяна",
									CallbackData: patternPapikSlots + "_" + strconv.Itoa(newUser.ID) + "_1",
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Рулетка Маятина",
									CallbackData: patternMayatinRoulette + "_" + strconv.Itoa(newUser.ID),
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Экзамен Повышева",
									CallbackData: patternPovyshevExams + "_" + strconv.Itoa(newUser.ID),
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
					ID:           "2",
					Title:        "Выберите игру!",
					ThumbnailURL: "https://i.ibb.co/Xfx3C5wH/image-1.jpg",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Блекджек с Даней Казанцевым",
									CallbackData: patternBlackjack + "_" + strconv.Itoa(user.ID),
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Слоты Папикяна",
									CallbackData: patternPapikSlots + "_" + strconv.Itoa(user.ID) + "_1",
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Рулетка Маятина",
									CallbackData: patternMayatinRoulette + "_" + strconv.Itoa(user.ID),
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Экзамен Повышева (в разработке)",
									CallbackData: patternPovyshevExams + "_" + strconv.Itoa(user.ID),
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nВаш баланс: %s I$Coins\nВыбирайте игру и побеждайте!", username, p.Sprintf("%d", user.Balance)),
					}},
				&models.InlineQueryResultArticle{
					ID:           "3",
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
				&models.InlineQueryResultArticle{
					ID:           "4",
					Title:        "Правила",
					ThumbnailURL: "https://casino.ru/wp-content/uploads/articles/poker/poker-1-400x266.jpg",
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nВот список наших развлечений:\n1. Слоты Папикяна. Вход 100.000, шанс на выигрыш 1/7, размер выигрыша 500.000\n2. Рулетка Маятина. Вход 100.000, шансы на выигрыш: 3/10 с возвратом 300.000, либо 1/10 с возвратом 1.000.000\n3. Экзамен Повышева (в разработке). Вход 100.000, шансы на выигрыш 1/6 в размере 500.000, либо взять седьмой \"удачный билет\" с шансом 50/50 и выигрышем 500.000, но ставкой 300.000\n\nВо всех автоматах есть 1/100 шанс на Гигавыигрыш в размере 10.000.000! (в разработке)", username),
					}},
				&models.InlineQueryResultArticle{
					ID:           "5",
					Title:        "Особые опции 🤭",
					ThumbnailURL: "https://linda.nyc3.cdn.digitaloceanspaces.com/370_npd_webp-o_18/sticker-fan_11513288_o.webp",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Выкупить квартиру 2М",
									CallbackData: patternBuyBackHouse + "_" + strconv.Itoa(user.ID),
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("🤭🤭🤭🤭🤭🤭🤭"),
					}},
				&models.InlineQueryResultGif{
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
									CallbackData: patternAddWatch + "_" + strconv.Itoa(user.ID),
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: `Рекламная интеграция\!
` + marketingSlots[rand.Intn(len(marketingSlots))],
						ParseMode: models.ParseModeMarkdown,
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
	if len(parts) < 3 {
		bs.Errorf("len(parts) < 3")
		return
	}

	koef, err := strconv.Atoi(parts[2])
	if err != nil {
		bs.Errorf("%v", err)
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

	if _, ok := bs.papikyanLock[user.ID]; ok {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Автомат отдыхает, и вы немного отдохните :)",
			ShowAlert:       true,
		})
		return
	}

	if user.LudomanNickname != update.CallbackQuery.From.Username {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Это не ваш автомат! Нажмите на название бота и тоже сможете сыграть :)",
			ShowAlert:       true,
		})
		return
	}

	if user.Balance < 100000*koef {
		if user.Balance < 100000 {
			bs.lossHandler(ctx, b, update, parts[1])
			return
		}
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "У вас недостаточно денег для этой ставки :/",
			ShowAlert:       true,
		})
		return
	}

	bs.papikyanLock[user.ID] = struct{}{}
	defer delete(bs.papikyanLock, user.ID)

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaVideo{
			Media:     "https://media.tenor.com/_yoDqyYP8aYAAAAM/casino77-slot-machine.gif",
			Caption:   "Крутимся...",
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
	})

	time.Sleep(5 * time.Second)

	num := rand.Intn(len(slotsResults))
	var res string
	switch num {
	case 0:
		err = bs.updateBalance(400000*koef, []int{user.ID}, false)
		if err != nil {
			bs.Errorf("%v", err)
			return
		}
		res = fmt.Sprintf("@%s, Победа! Вы получаете +%s I$Coins. Ваш текущий баланс: %s I$Coins", update.CallbackQuery.From.Username, p.Sprintf("%d", 500000*koef), p.Sprintf("%d", user.Balance+400000*koef))
	default:
		err = bs.updateBalance(-100000*koef, []int{user.ID}, false)
		if err != nil {
			bs.Errorf("%v", err)
			return
		}
		res = fmt.Sprintf("@%s, Неудача! Ваш текущий баланс: %s I$Coins", update.CallbackQuery.From.Username, p.Sprintf("%d", user.Balance-100000*koef))
	}

	pic := slotsResults[num]

	//if rand.Intn(667) == 666 {
	//	err = bs.updateBalance(100000000*koef, []int{user.ID}, false)
	//	if err != nil {
	//		bs.Errorf("%v", err)
	//		return
	//	}
	//	res = fmt.Sprintf("@%s, ГИГАДЖЕКПОТ! С прошедшим 125-м днём рождения нашего любимого ВУЗа!\nВы получаете +%s I$Coins. Ваш текущий баланс: %s I$Coins", update.CallbackQuery.From.Username, p.Sprintf("%d", 100000000*koef), p.Sprintf("%d", 100000000*koef+user.Balance))
	//	pic = jackPotITMO
	//}

	if rand.Intn(112) == 111 {
		err = bs.updateBalance(10000000*koef, []int{user.ID}, false)
		if err != nil {
			bs.Errorf("%v", err)
			return
		}
		res = fmt.Sprintf("@%s, ДЖЕКПОТ! Вы получаете +%s I$Coins. Ваш текущий баланс: %s I$Coins", update.CallbackQuery.From.Username, p.Sprintf("%d", 10000000*koef), p.Sprintf("%d", 10000000*koef+user.Balance))
		pic = jackPotPapikyan
	}

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
					Text:         "Сыграть на 100k",
					CallbackData: patternPapikSlots + "_" + parts[1] + "_1",
				},
			},
			{
				models.InlineKeyboardButton{
					Text:         "Сыграть на 500k",
					CallbackData: patternPapikSlots + "_" + parts[1] + "_5",
				},
			},
			{
				models.InlineKeyboardButton{
					Text:         "Сыграть на 1m",
					CallbackData: patternPapikSlots + "_" + parts[1] + "_10",
				},
			},
		}},
	})

	if err != nil {
		if !strings.Contains(err.Error(), "cannot unmarshal bool") {
			bs.Errorf("%v", err)
		}
		if strings.Contains(err.Error(), "retry_after") {
			retryAfter := strings.Split(err.Error(), " ")
			retryAfterTime := retryAfter[len(retryAfter)-1]
			retryTime, err := strconv.Atoi(retryAfterTime)
			if err != nil {
				bs.Errorf("%v", err)
				return
			}
			time.Sleep(time.Duration(retryTime) * time.Second)
			errorMsg := fmt.Sprintf("Извините, бот задерживается из-за перегруза запросов.\nЗадержка:%ds", retryTime)

			b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
				InlineMessageID: update.CallbackQuery.InlineMessageID,
				Media: &models.InputMediaPhoto{
					Media:     pic,
					Caption:   res + "\n" + errorMsg,
					ParseMode: models.ParseModeHTML,
					//HasSpoiler: true,
				},
				ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						models.InlineKeyboardButton{
							Text:         "Сыграть на 100k",
							CallbackData: patternPapikSlots + "_" + parts[1] + "_1",
						},
					},
					{
						models.InlineKeyboardButton{
							Text:         "Сыграть на 500k",
							CallbackData: patternPapikSlots + "_" + parts[1] + "_5",
						},
					},
					{
						models.InlineKeyboardButton{
							Text:         "Сыграть на 1m",
							CallbackData: patternPapikSlots + "_" + parts[1] + "_10",
						},
					},
				}},
			})
		}
	}
}
func (bs *BotService) lossHandler(ctx context.Context, b *bot.Bot, update *models.Update, userId string) {
	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaVideo{
			Media:     "https://media.tenor.com/aSkdq3IU0g0AAAAM/laughing-cat.gif",
			Caption:   "Вы израходовали свой баланс!",
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				models.InlineKeyboardButton{
					Text:         "Хочу откупиться!",
					CallbackData: patternBuyBack + "_" + userId,
				},
			},
		}},
	})
}

func (bs *BotService) BuybackHouseHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("invalid callback data: %s", update.CallbackQuery.Data)
		return
	}

	userID, err := strconv.Atoi(parts[1])
	if err != nil {
		bs.Errorf("invalid user id: %v", err)
		return
	}

	user, err := bs.cr.LudomanByID(ctx, userID)
	if err != nil {
		bs.Errorf("failed to get user: %v", err)
		return
	}

	if user.LudomanNickname != update.CallbackQuery.From.Username {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Это не ваше окно выкупа квартиры !")
		return
	}

	if user.Balance < 2000000 {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Недостаточно i$ coins для выкупа квартиры обратно :(")
		return
	}

	if user.Losses <= 0 {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Вам нечего выкупать, пора сыграть в i$ казик")
		return
	}

	bs.updateBalance(-2000000, []int{user.ID}, true)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
	user.Losses = user.Losses - 1
	_, err = bs.cr.UpdateLudoman(ctx, user, db.WithColumns(db.Columns.Ludoman.Losses))
	if err != nil {
		bs.Errorf("failed to update user: %v", err)
		return
	}

	bs.respondToCallback(ctx, b, update.CallbackQuery.ID, fmt.Sprintf("Вы успешно выкупили квартиру обратно!\nКоличество проданных квартир: %d", user.Losses))
}

func (bs *BotService) respondToCallback(ctx context.Context, b *bot.Bot, callbackID, text string) {
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackID,
		Text:            text,
		ShowAlert:       true,
	})
	if err != nil {
		bs.Errorf("failed to answer callback query: %v", err)
	}
}
func (bs *BotService) PlayersRatingHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	topWins, err := bs.cr.LudomenByFilters(
		ctx,
		&db.LudomanSearch{},
		db.Pager{PageSize: 10},
		db.WithSort(db.NewSortField(db.Columns.Ludoman.TotalWon, true)),
	)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
	topLosses, err := bs.cr.LudomenByFilters(
		ctx,
		&db.LudomanSearch{},
		db.Pager{PageSize: 10},
		db.WithSort(db.NewSortField(db.Columns.Ludoman.TotalLost, true)),
	)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	topBalance, err := bs.cr.LudomenByFilters(
		ctx,
		&db.LudomanSearch{},
		db.Pager{PageSize: 10},
		db.WithSort(db.NewSortField(db.Columns.Ludoman.Balance, true)),
	)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	// Функция для добавления 1 к индексу (так как индексация с 0)
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatDigit": func(in int) string {
			return p.Sprintf("%d", in)
		},
	}

	winsTemplate := `{{- range $index, $ludoman := . }}
{{- printf "\n%d. Никнейм: @%s, Выиграно: %s, Квартир продано: %d" (add $index 1) $ludoman.LudomanNickname (formatDigit $ludoman.TotalWon) $ludoman.Losses}}
{{- end }}`

	lossesTemplate := `{{- range $index, $ludoman := . }}
{{- printf "\n%d. Никнейм: @%s, Проиграно: %s, Квартир продано: %d" (add $index 1) $ludoman.LudomanNickname (formatDigit $ludoman.TotalLost) $ludoman.Losses}}
{{- end }}`

	balanceTemplate := `{{- range $index, $ludoman := . }}
{{- printf "\n%d. Никнейм: @%s, Баланс: %s, Квартир продано: %d" (add $index 1) $ludoman.LudomanNickname (formatDigit $ludoman.Balance) $ludoman.Losses}}
{{- end }}`

	var bufWins bytes.Buffer
	var bufLosses bytes.Buffer
	var bufBalance bytes.Buffer

	tmplWins, err := template.New("winsList").Funcs(funcMap).Parse(winsTemplate)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
	tmplLosses, err := template.New("lossesList").Funcs(funcMap).Parse(lossesTemplate)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
	tmplBalance, err := template.New("balanceList").Funcs(funcMap).Parse(balanceTemplate)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	if err = tmplWins.Execute(&bufWins, topWins); err != nil {
		bs.Errorf("%v", err)
		return
	}
	if err = tmplLosses.Execute(&bufLosses, topLosses); err != nil {
		bs.Errorf("%v", err)
		return
	}
	if err = tmplBalance.Execute(&bufBalance, topBalance); err != nil {
		bs.Errorf("%v", err)
		return
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Text: "Список топ игроков по суммарным выигрышам:" + bufWins.String() +
			"\n\nСписок топ игроков по суммарным проигрышам:" + bufLosses.String() +
			"\n\nСписок топ игроков по балансу:" + bufBalance.String(),
	})
	if err != nil {
		bs.Errorf("%v", err)
	}
}

func (bs *BotService) BuyBackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("parts === ", update.CallbackQuery.Data)
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
		if err != nil {
			bs.Errorf("%v", err)
		}
		return
	}

	if _, ok := bs.buyBackLock[user.ID]; ok {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Автомат отдыхает, и вы немного отдохните :)",
			ShowAlert:       true,
		})
		return
	}

	bs.buyBackLock[user.ID] = struct{}{}
	defer delete(bs.buyBackLock, user.ID)

	if user.Losses >= bs.limitByBack {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Вы превысили лимит по продажам квартир. Чтобы повысить лимит, поставьте звездочку в гитхабе")
		return
	}

	user.Balance = initialBalance
	if user.TgID == 0 {
		user.TgID = int(update.CallbackQuery.From.ID)
	}

	user.Losses += 1
	_, err = bs.cr.UpdateLudoman(ctx, user, db.WithColumns(db.Columns.Ludoman.Balance, db.Columns.Ludoman.Losses, db.Columns.Ludoman.TgID))
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     "https://i.ibb.co/6R0Cz78Q/image-4.jpg",
			Caption:   fmt.Sprintf("Вы откупились! Счетчик ваших проданных квартир: %d\nНажмите на название бота и проиграйте всё снова, или может быть сегодня вам повезет попасть в топ рейтинга?)\n\n ваш текущий лимит выкупов: %d / %d \n\n Чтобы увеличить лимит продаж квартир, поставьте звездочку в гитхабе 👉👈 https://github.com/kroexov/ludomania", user.Losses, user.Losses, bs.limitByBack),
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
	})
}

func (bs *BotService) MayatinRouletteHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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
		bs.lossHandler(ctx, b, update, parts[1])
		return
	}

	if bs.isMayatinRouletteActive {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Рулетка уже идет! Присоединяйтесь к текущей!",
			ShowAlert:       true,
		})
		return
	}

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaVideo{
			Media:     "https://i.pinimg.com/originals/32/37/bf/3237bf1e172a6089e0c437ffd3b28010.gif",
			Caption:   fmt.Sprintf("Рулетка Маятина началась! Выбирайте ваш слот в рулетке!\nСтавка 500.000, слот 'Уважаемый коллега дает 10x выигрыш, но выпадает реже'\nОсталось 20 секунд!"),
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				models.InlineKeyboardButton{
					Text:         fmt.Sprintf("Надёжность! (0 ставок)"),
					CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetN,
				},
			},
			{
				models.InlineKeyboardButton{
					Text:         fmt.Sprintf("Производительность! (0 ставок)"),
					CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetP,
				},
			},
			{
				models.InlineKeyboardButton{
					Text:         fmt.Sprintf("Безопасность! (0 ставок)"),
					CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetB,
				},
			},
			{
				models.InlineKeyboardButton{
					Text:         fmt.Sprintf("Уважаемый коллега 😎 (10x выигрыш, 0 ставок)"),
					CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetU,
				},
			},
		}},
	})

	bs.mayatinRouletteBets = new(sync.Map)
	bs.isMayatinRouletteActive = true
	bs.mayatinRouletteUsers = make(map[int]struct{})
	bs.mayatinCategoriesVotes = make(map[string]int)

	var errorMsg string

	for i := 0; i < 10; i++ {
		bs.mu.Lock()
		_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{
			Caption:         fmt.Sprintf("Рулетка Маятина началась! Выбирайте ваш слот в рулетке!\nСтавка 500.000, слот 'Уважаемый коллега дает 10x выигрыш, но выпадает реже'\nОсталось %d секунд!\n%s", (10-i)*2, errorMsg),
			InlineMessageID: update.CallbackQuery.InlineMessageID,
			ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Надёжность! (%d ставок)", bs.mayatinCategoriesVotes[patternMayatinRouletteBetN]),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetN,
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Производительность! (%d ставок)", bs.mayatinCategoriesVotes[patternMayatinRouletteBetP]),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetP,
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Безопасность! (%d ставок)", bs.mayatinCategoriesVotes[patternMayatinRouletteBetB]),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetB,
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Уважаемый коллега 😎 (10x выигрыш, %d ставок)", bs.mayatinCategoriesVotes[patternMayatinRouletteBetU]),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetU,
					},
				},
			}},
		})
		if err != nil {
			if !strings.Contains(err.Error(), "cannot unmarshal bool") {
				bs.Errorf("%v", err)
			}

			if strings.Contains(err.Error(), "retry_after") {
				retryAfter := strings.Split(err.Error(), " ")
				retryAfterTime := retryAfter[len(retryAfter)-1]
				retryTime, err := strconv.Atoi(retryAfterTime)
				if err != nil {
					bs.Errorf("%v", err)
					return
				}
				time.Sleep(time.Duration(retryTime) * time.Second)
				errorMsg = fmt.Sprintf("Извините, бот задерживается из-за перегруза запросов.\nЗадержка:%ds", retryTime)
			}
		}
		bs.mu.Unlock()
		time.Sleep(2 * time.Second)
	}

	i := rand.Intn(100)
	var selectedCategory string
	switch {
	case i <= 30:
		selectedCategory = patternMayatinRouletteBetP
	case i <= 60:
		selectedCategory = patternMayatinRouletteBetB
	case i <= 90:
		selectedCategory = patternMayatinRouletteBetN
	default:
		selectedCategory = patternMayatinRouletteBetU
	}
	cat := mayatinCategories[selectedCategory]

	var winners []int
	// Iterating over sync.Map
	bs.mayatinRouletteBets.Range(func(key, value interface{}) bool {
		println(key.(int), value.(string))
		if value.(string) == selectedCategory {
			winners = append(winners, key.(int))
		}
		return true
	})

	err = bs.updateBalance(-500000, intKeys(bs.mayatinRouletteUsers), false)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	var result string
	if len(winners) == 0 {
		result = `Победителей нет 🫵😹`
	} else {
		winUsers, err := bs.cr.LudomenByFilters(ctx, &db.LudomanSearch{IDs: winners}, db.PagerNoLimit)
		if err != nil {
			bs.Errorf("%v", err)
		}
		result = "Список победителей: "
		for _, winUser := range winUsers {
			result += "@" + winUser.LudomanNickname + " "
		}
		result += fmt.Sprintf("\nПобедителям начислено: %s", p.Sprintf("%d", cat.WinSum))

		err = bs.updateBalance(cat.WinSum, db.Ludomans(winUsers).IDs(), false)
		if err != nil {
			bs.Errorf("%v", err)
			return
		}
	}

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     cat.CategoryPic,
			Caption:   fmt.Sprintf("Рулетка Маятина завершена! Выпало: %s!\n%s", cat.CategoryName, result),
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
	})

	bs.isMayatinRouletteActive = false
}

func (bs *BotService) MayatinRouletteBetHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("len(parts) < 2")
		return
	}
	userBet := parts[1]

	// find user
	user, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &update.CallbackQuery.From.Username})
	if err != nil {
		bs.Errorf("%v", err)
	} else if user == nil {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Вас еще нет в нашей базе данных :( Попробуйте сначала зарегаться в боте",
			ShowAlert:       true,
		})
		return
	}

	if _, ok := bs.mayatinRouletteUsers[user.ID]; ok {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Вы уже поставили на рулетку! Теперь ждите и молитесь :)",
			ShowAlert:       true,
		})
		return
	}

	if user.Balance < 500000 {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "У вас недостаточно денег для этой ставки :/",
			ShowAlert:       true,
		})
		return
	}

	if bs.mayatinRouletteUsers == nil {
		return
	}

	bs.mu.Lock()
	bs.mayatinRouletteUsers[user.ID] = struct{}{}
	bs.mayatinCategoriesVotes["_"+userBet]++
	bs.mu.Unlock()

	bs.mayatinRouletteBets.Store(user.ID, "_"+userBet)
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

	err = bs.updateBalance(500000, []int{userId}, true)
	if err != nil {
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

func Pointer[T any](in T) *T {
	return &in
}

func intKeys(in map[int]struct{}) []int {
	out := make([]int, 0, len(in))
	for v := range in {
		out = append(out, v)
	}
	return out
}

func (bs *BotService) updateBalance(sum int, ids []int, balanceOnly bool) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
	UPDATE ludomans
	SET balance = balance + ?0,
		"totalWon" = CASE
	WHEN ?2 = False AND ?0 > 0 THEN COALESCE("totalWon", 0) + ?0
	ELSE "totalWon"
	END,
		"totalLost" = CASE
	WHEN ?2 = False AND ?0 <= 0 THEN COALESCE("totalLost", 0) + ABS(?0)
	ELSE "totalLost"
	END
	WHERE "ludomanId" in (?1)
	`

	_, err := bs.db.Exec(query, sum, pg.In(ids), balanceOnly)
	return err
}
