package ludomania

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	patternBlackjack       = "bj"
	patternBlackjackBet    = "bjBet"
	patternBlackjackAction = "bjAct"
)

var (
	suits = []string{"♠️", "♥️", "♦️", "♣️"}
	ranks = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
)

type BlackjackGame struct {
	UserID      int
	PlayerHand  string
	DealerHand  string
	Bet         int
	Deck        string
	IsDoubled   bool
	IsCompleted bool

	//надо добавить обработку
	mu sync.Mutex
}

func (bs *BotService) BlackjackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	fmt.Println("callback  == ", parts)
	if len(parts) < 2 {
		bs.Errorf("invalid blackjack data: %s", update.CallbackQuery.Data)
		return
	}

	action := parts[0]
	userID := -1
	if len(parts) == 3 {
		userID, _ = strconv.Atoi(parts[2])
	} else {
		userID, _ = strconv.Atoi(parts[1])
	}
	//userID, _ := strconv.Atoi(parts[1])
	if userID == -1 {
		bs.Errorf("invalid blackjack data: %s", update.CallbackQuery.Data)
		return
	}
	fmt.Println("id юзера == ", userID)

	switch {
	case action == patternBlackjack:
		bs.handleBlackjackStart(ctx, b, update, userID)
	case strings.HasPrefix(action, patternBlackjackBet):
		bs.handleBlackjackBet(ctx, b, update, userID, parts)
	case strings.HasPrefix(action, patternBlackjackAction):
		bs.handleBlackjackAction(ctx, b, update, userID, parts)
	}
}

func (bs *BotService) handleBlackjackStart(ctx context.Context, b *bot.Bot, update *models.Update, userID int) {

	markup := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "100K", CallbackData: fmt.Sprintf("bjBet_1_%d", userID)},
				{Text: "500K", CallbackData: fmt.Sprintf("bjBet_5_%d", userID)},
				{Text: "1M", CallbackData: fmt.Sprintf("bjBet_10_%d", userID)},
			},
		},
	}

	_, err := b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     "https://ibb.co/SDxYjTgD",
			Caption:   "Выберите ставку для Блекджека с Даней Казанцевым!",
			ParseMode: models.ParseModeHTML,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		bs.Errorf("%v", err)
	}
}

func (bs *BotService) handleBlackjackBet(ctx context.Context, b *bot.Bot, update *models.Update, userID int, parts []string) {
	bet, _ := strconv.Atoi(parts[1])
	//userID, _ := strconv.Atoi(parts[2])
	user, err := bs.cr.LudomanByID(ctx, userID)
	if err != nil {
		bs.Errorf("Error getting user: %v", err)
		return
	}
	if user == nil {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Пользователь не найден!")
		return
	}

	betAmount := bet * 100000

	if user.Balance < betAmount {
		bs.respondToCallback(ctx, b, update.CallbackQuery.ID, "Недостаточно средств!")
		return
	}

	deck := generateShuffledDeck()
	game := &BlackjackGame{
		UserID:     userID,
		Bet:        bet,
		PlayerHand: drawCards(&deck, 2),
		DealerHand: drawCards(&deck, 2),
		Deck:       strings.Join(deck, " "),
	}

	bs.blackjackGames.Store(userID, game)
	bs.updateBalance(-betAmount, []int{userID}, false)
	bs.renderGameState(ctx, b, update.CallbackQuery.InlineMessageID, userID, game, false)
}

func generateShuffledDeck() []string {
	deck := make([]string, 0, 52)
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, rank+suit)
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func formatCard(card string) string {
	for _, suit := range suits {
		if strings.HasSuffix(card, suit) {
			rank := strings.TrimSuffix(card, suit)
			return fmt.Sprintf("%s%s", rank, suit)
		}
	}
	return card
}

func drawCards(deck *[]string, n int) string {
	if len(*deck) < n {
		newDeck := generateShuffledDeck()
		*deck = append(*deck, newDeck...)
	}
	hand := strings.Join((*deck)[:n], " ")
	*deck = (*deck)[n:]
	return hand
}

func (bs *BotService) renderGameState(ctx context.Context, b *bot.Bot, inlineMsgID string, userID int, game *BlackjackGame, showDealer bool) {
	playerValue := calculateHandValue(game.PlayerHand)

	dealerHand := "?"
	if !showDealer {
		cards := strings.Split(game.DealerHand, " ")
		if len(cards) > 0 {
			dealerHand = formatCard(cards[0]) + " ?"
		}
	} else {
		dealerHand = formatHand(game.DealerHand)
	}

	caption := fmt.Sprintf("Ваши карты: %s (%d)\nДилер: %s",
		formatHand(game.PlayerHand),
		playerValue,
		dealerHand,
	)
	if calculateHandValue(game.PlayerHand) >= 21 {
		game.IsCompleted = true
	}
	var buttons [][]models.InlineKeyboardButton
	if !game.IsCompleted {
		buttons = [][]models.InlineKeyboardButton{
			{
				{Text: "Взять", CallbackData: fmt.Sprintf("bjAct_hit_%d", userID)},
				{Text: "Стоп", CallbackData: fmt.Sprintf("bjAct_stand_%d", userID)},
			},
		}
		if user, _ := bs.cr.LudomanByID(ctx, userID); user.Balance >= game.Bet*100000 && len(game.PlayerHand) < 20 {
			fmt.Println("DOUBLE")
			buttons[0] = append(buttons[0], models.InlineKeyboardButton{
				Text:         "Удвоить",
				CallbackData: fmt.Sprintf("bjAct_double_%d", userID),
			})
		}
	}

	_, err := b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: inlineMsgID,
		Media: &models.InputMediaPhoto{
			Media:     "https://ibb.co/SDxYjTgD",
			Caption:   caption,
			ParseMode: models.ParseModeHTML,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: buttons},
	})
	if err != nil {
		bs.Errorf("%v", err)
	}
}

func (bs *BotService) handleBlackjackAction(ctx context.Context, b *bot.Bot, update *models.Update, userID int, parts []string) {
	action := parts[1]
	gameInterface, _ := bs.blackjackGames.Load(userID)
	game := gameInterface.(*BlackjackGame)

	fmt.Println("hand size == ", game.PlayerHand)
	fmt.Println("hand size == ", len(game.PlayerHand))

	switch action {
	case "hit":
		time.Sleep(3 * time.Second)
		deck := strings.Split(game.Deck, " ")
		if len(deck) == 0 {
			bs.Errorf("Deck is empty")
			return
		}
		newCard := deck[0]
		game.PlayerHand += " " + newCard
		game.Deck = strings.Join(deck[1:], " ")
		if calculateHandValue(game.PlayerHand) >= 21 {
			game.IsCompleted = true
		}

	case "stand":
		time.Sleep(3 * time.Second)
		game.IsCompleted = true

	case "double":
		if user, _ := bs.cr.LudomanByID(ctx, game.UserID); user.Balance >= game.Bet*100000 && len(game.PlayerHand) < 20 {
			time.Sleep(3 * time.Second)
			game.IsDoubled = true
			deck := strings.Split(game.Deck, " ")
			newCard := deck[0]
			game.PlayerHand += " " + newCard
			game.Deck = strings.Join(deck[1:], " ")
			game.IsCompleted = true
			bs.updateBalance(-game.Bet*100000, []int{game.UserID}, false)
		}
	}

	if game.IsCompleted {
		bs.finalizeGame(ctx, b, update.CallbackQuery.InlineMessageID, userID, game)
	} else {
		bs.blackjackGames.Store(userID, game)
		bs.renderGameState(ctx, b, update.CallbackQuery.InlineMessageID, userID, game, false)
	}
}

func (bs *BotService) finalizeGame(ctx context.Context, b *bot.Bot, inlineMsgID string, userID int, game *BlackjackGame) {

	dealerHand := game.DealerHand
	deck := strings.Split(game.Deck, " ")

	for calculateHandValue(dealerHand) < 17 && len(deck) > 0 {
		dealerHand += " " + deck[0]
		deck = deck[1:]
	}

	playerValue := calculateHandValue(game.PlayerHand)
	dealerValue := calculateHandValue(dealerHand)
	multiplier := 1
	if game.IsDoubled {
		multiplier = 2
	}

	var result string
	var resultImage string
	defaultImage := "https://ibb.co/SDxYjTgD"

	switch {
	case playerValue == 21 && len(game.PlayerHand) < 20:
		result = "БЛЕКДЖЕК! Поздравляем, выплата 3:2 !"
		resultImage = "https://ibb.co/Vc0WKybS" // Игрок выиграл
		bs.updateBalance(game.Bet*250000*multiplier, []int{userID}, false)
	case playerValue > 21:
		result = "Перебор! Вы проиграли"
		resultImage = "https://ibb.co/B50vbT3R" // Диллер выиграл
	case dealerValue > 21 || playerValue > dealerValue:
		result = "Вы выиграли!"
		resultImage = "https://ibb.co/Vc0WKybS" // Игрок выиграл
		bs.updateBalance(game.Bet*200000*multiplier, []int{userID}, false)
	case playerValue == dealerValue:
		result = "Ничья! Возврат ставки"
		resultImage = defaultImage
		bs.updateBalance(game.Bet*100000*multiplier, []int{userID}, false)
	default:
		result = "Вы проиграли"
		resultImage = "https://ibb.co/B50vbT3R" // Диллер выиграл
	}

	if resultImage == "" {
		resultImage = defaultImage
	}

	user, _ := bs.cr.LudomanByID(ctx, userID)
	caption := fmt.Sprintf("%s\nВаши карты: %s (%d)\nКарты дилера: %s (%d)\nБаланс: %s",
		result,
		formatHand(game.PlayerHand),
		playerValue,
		formatHand(dealerHand),
		dealerValue,
		p.Sprintf("%d", user.Balance),
	)

	markup := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Сыграть снова", CallbackData: fmt.Sprintf("bj_%d", userID)},
			},
		},
	}

	_, err := b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: inlineMsgID,
		Media: &models.InputMediaPhoto{
			Media:     resultImage,
			Caption:   caption,
			ParseMode: models.ParseModeHTML,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		bs.Errorf("%v", err)
	}
	bs.blackjackGames.Delete(userID)
}

func calculateHandValue(hand string) int {
	value, aces := 0, 0
	cards := strings.Split(hand, " ")

	for _, card := range cards {
		card = strings.TrimSpace(card)
		if card == "" {
			continue
		}

		for _, suit := range suits {
			if strings.HasSuffix(card, suit) {
				rank := strings.TrimSuffix(card, suit)
				switch rank {
				case "J", "Q", "K":
					value += 10
				case "A":
					value += 11
					aces++
				default:
					if num, err := strconv.Atoi(rank); err == nil {
						value += num
					}
				}
				break
			}
		}
	}

	for value > 21 && aces > 0 {
		value -= 10
		aces--
	}
	return value
}

func formatHand(hand string) string {
	cards := strings.Split(hand, " ")
	var formatted []string
	for _, card := range cards {
		card = strings.TrimSpace(card)
		if card != "" {
			formatted = append(formatted, formatCard(card))
		}
	}
	return strings.Join(formatted, " ")
}
