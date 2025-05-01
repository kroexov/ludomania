package ludomania

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gradebot/pkg/db"
	"strconv"
	"strings"
)

const (
	patternBuyBackHouse = "buyBackHouse"
	patternBuyTicket    = "buyTicket"
	patternCoef1        = "setCoef_1"
	patternCoef10       = "setCoef_10"
	patternCoef100      = "setCoef_100"
	patternCoef200      = "setCoef_200"
	patternCoef500      = "setCoef_500"
	patternOfferCoef    = "offerCoef:"
)

func (bs *BotService) extraOptions(userId int) models.InlineQueryResult {
	return &models.InlineQueryResultArticle{
		ID:           "5",
		Title:        "Особые опции 🤭",
		ThumbnailURL: "https://linda.nyc3.cdn.digitaloceanspaces.com/370_npd_webp-o_18/sticker-fan_11513288_o.webp",
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					models.InlineKeyboardButton{
						Text:         "Выкупить квартиру (2М)",
						CallbackData: patternBuyBackHouse + "_" + strconv.Itoa(userId),
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         "Купить билет на ирл турик по покеру (100M)",
						CallbackData: patternBuyTicket + "_" + strconv.Itoa(userId),
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         "поставить кэф x1 ",
						CallbackData: patternCoef1 + "_" + strconv.Itoa(userId),
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         "поставить кэф x10 ",
						CallbackData: patternCoef10 + "_" + strconv.Itoa(userId),
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         "поставить кэф x100 ",
						CallbackData: patternCoef100 + "_" + strconv.Itoa(userId),
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         "поставить кэф x200 ",
						CallbackData: patternCoef200 + "_" + strconv.Itoa(userId),
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         "поставить кэф x500 ",
						CallbackData: patternCoef500 + "_" + strconv.Itoa(userId),
					},
				},
			}},
		InputMessageContent: &models.InputTextMessageContent{
			MessageText: fmt.Sprintf("🤭🤭🤭🤭🤭🤭🤭"),
		}}
}
func (bs *BotService) OfferBuyCoef(ctx context.Context, b *bot.Bot, inlineMsgID string, userID, coef int) error {
	keyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{
				Text:         "Да",
				CallbackData: fmt.Sprintf("%syes_%d_%d", patternOfferCoef, userID, coef),
			},
			{
				Text:         "Нет",
				CallbackData: fmt.Sprintf("%sno_%d_%d", patternOfferCoef, userID, coef),
			},
		}},
	}
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		InlineMessageID: inlineMsgID,
		Text:            fmt.Sprintf("Открыть коэффициент %d за %d I$ coins?", coef, coef*1_000_000),
		ReplyMarkup:     keyboard,
	})
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (bs *BotService) SetCoefHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("invalid callback data: %s", update.CallbackQuery.Data)
		return
	}

	userID, err := strconv.Atoi(parts[2])
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
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Это не ваше окно !")
		return
	}
	coef, err := strconv.Atoi(parts[1])
	fmt.Println("parts == ", parts)
	if err != nil {
		bs.Errorf("invalid coef: %v", err)
		return
	}

	messageID := update.CallbackQuery.InlineMessageID
	fmt.Println("messageID == ", messageID)
	//messageID := update.CallbackQuery.Message.Message.ID
	fmt.Printf("Inf if")

	switch coef {
	case 1:
		user.Coefficient = 1
	case 10:
		if user.Data.K10 {
			user.Coefficient = 10
		} else {
			_ = bs.OfferBuyCoef(ctx, b, messageID, userID, 10)
			return
		}
	case 100:
		if user.Data.K100 {
			user.Coefficient = 100
		} else {
			_ = bs.OfferBuyCoef(ctx, b, messageID, userID, 100)
			return
		}
	case 200:
		if user.Data.K200 {
			user.Coefficient = 200
		} else {
			_ = bs.OfferBuyCoef(ctx, b, messageID, userID, 200)
			return
		}
	case 500:
		if user.Data.K500 {
			user.Coefficient = 500
		} else {
			_ = bs.OfferBuyCoef(ctx, b, messageID, userID, 500)
			return
		}
	default:
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Недопустимый коэффициент.")
		return
	}

	if _, err := bs.cr.UpdateLudoman(ctx, user,
		db.WithColumns(db.Columns.Ludoman.Coefficient, db.Columns.Ludoman.Data)); err != nil {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Ошибка при сохранении.")
		bs.Errorf("update user coef failed: %v", err)
		return
	}

	bs.respondToCallback(ctx, b, update.CallbackQuery.ID, fmt.Sprintf("Коэффициент установлен: %d", coef))
	_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		InlineMessageID: messageID,
		Text:            fmt.Sprintf("Ваш коэффициент %d активен", coef),
		ReplyMarkup:     nil,
	})
}
func (bs *BotService) HandleOfferCoefCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	data := update.CallbackQuery.Data[len(patternOfferCoef):]
	parts := strings.Split(data, "_")
	fmt.Println(parts)
	if len(parts) != 3 {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Неверный формат данных.")
		return
	}

	action, uidStr, coefStr := parts[0], parts[1], parts[2]
	userID, err1 := strconv.Atoi(uidStr)
	coefVal, err2 := strconv.Atoi(coefStr)
	if err1 != nil || err2 != nil {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Неверные данные.")
		return
	}

	user, err := bs.cr.LudomanByID(ctx, userID)
	if err != nil || user == nil {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Пользователь не найден.")
		return
	}

	if user.LudomanNickname != update.CallbackQuery.From.Username {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Это не ваше предложение.")
		return
	}

	if action == "yes" {
		if user.Balance >= coefVal*1000000 {
			bs.updateBalance(-coefVal*1000000, []int{user.ID}, true, 1)
			switch {
			case coefVal == 10:
				user.Data.K10 = true
				user.Coefficient = 10
			case coefVal == 100:
				user.Data.K100 = true
				user.Coefficient = 100
			case coefVal == 200:
				user.Data.K200 = true
				user.Coefficient = 200
			case coefVal == 500:
				user.Data.K500 = true
				user.Coefficient = 500
			}
			_, err = bs.cr.UpdateLudoman(ctx, user,
				db.WithColumns(db.Columns.Ludoman.Coefficient, db.Columns.Ludoman.Data),
			)
			user.Coefficient = coefVal
			bs.respondToCallback(ctx, b, update.CallbackQuery.ID, fmt.Sprintf("Коэффициент установлен: %d", coefVal))
			if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				InlineMessageID: update.CallbackQuery.InlineMessageID,
				Text:            fmt.Sprintf("Коэффициент %d установлен", coefVal),
			}); err != nil {
				bs.Errorf("failed to edit text: %v", err)
			}
		} else {
			bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Недостаточно I$ coins для покупки коэффициента")
			if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				InlineMessageID: update.CallbackQuery.InlineMessageID,
				Text:            fmt.Sprintf("У вас недостаточно I$ coins для покупки данного коэффициента"),
			}); err != nil {
				bs.Errorf("failed to edit text: %v", err)
			}
		}
	} else {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Вы отказались от покупки коэффициента.")
		if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			InlineMessageID: update.CallbackQuery.InlineMessageID,
			Text:            fmt.Sprintf("Жаль, что вы отказались от покупки коэффициента"),
		}); err != nil {
			bs.Errorf("failed to edit text: %v", err)
		}
	}

	if _, err := b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		ReplyMarkup:     nil,
	}); err != nil {
		bs.Errorf("failed to remove keyboard: %v", err)
	}

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

	err = bs.updateBalance(-2000000, []int{user.ID}, true, 1)
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

func (bs *BotService) BuyTicketHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Не ваше окно покупки билета!")
		return
	}

	if user.Balance < 100000000 {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Недостаточно i$ coins для покупки билета на турик :(")
		return
	}

	err = bs.updateBalance(-100000000, []int{user.ID}, true, 1)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	_, err = bs.db.Exec(`insert into tournamentusers (userid) values (?);`, user.ID)
	if err != nil {
		bs.Errorf("%v", err)
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, err.Error())
		return
	}

	link, err := b.CreateChatInviteLink(ctx, &bot.CreateChatInviteLinkParams{
		ChatID:      bs.tournamentChatId,
		Name:        "Инвайт в турик по покеру ИС.Работа",
		MemberLimit: 1,
	})
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		Text:            "Спасибо за покупку!\nВот ваша ссылка в чат на турик по покеру: " + link.InviteLink,
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		ReplyMarkup:     nil,
	})
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
}
