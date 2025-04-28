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
			}},
		InputMessageContent: &models.InputTextMessageContent{
			MessageText: fmt.Sprintf("🤭🤭🤭🤭🤭🤭🤭"),
		}}
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

	err = bs.updateBalance(-2000000, []int{user.ID}, true)
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

	err = bs.updateBalance(-100000000, []int{user.ID}, true)
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
