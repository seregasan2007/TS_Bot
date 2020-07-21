package main

import (
	"log"
	"time"
	"strconv"
	tb "gopkg.in/tucnak/telebot.v2" 
	cron "github.com/robfig/cron/v3" 
)

// Function to add and count your time shift

func addTS(tstime, id int, users map[int]int, sender tb.Recipient) (sum int, uresp string) {
					
	ts := users[id]
	newts := tstime
	sum = ts
	if newts < 10 {
		sum = ts + newts*60
	} else {
		sum = ts + newts
	}
	

	if sum > 480 {
		sum = 0
		uresp = "⚠️ Трудозатраты не могут превышать 480 минут в день"
		return
	} else if sum < 0 {
		sum = 0
		uresp = "⚠️ Трудозатраты не могут быть меньше нуля"
		return
	} else {
		users[id] = sum
		hour := users[id] / 60
		hour_min := users[id] - hour*60
		balance := 480 - users[id]
		balance_hour := 8-hour
		balance_min := hour_min
		if balance_min != 0 {
			balance_hour += -1
			balance_min = 60-balance_min
		}
		uresp = "Списано: " + strconv.Itoa(users[id]) + " Минут" + "\nВ часах: " + strconv.Itoa(hour) + ":" + strconv.Itoa(hour_min) + "\nОсталось: " + strconv.Itoa(balance) + "\nОсталось в часах: " + strconv.Itoa(balance_hour) + ":" + strconv.Itoa(balance_min)
		return
	}
}

func main() {
	bot_token := "1071824688:AAE73MLfgtHHi4Y2_Re6gO-r8mkHXBvJyaM"
	b, err := tb.NewBot(tb.Settings{
		Token: bot_token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return
	}

	senders := make(map[int]*tb.User)   // Map [User ID int] info for send message
	users := make(map[int]int)          // Map [User ID int] sum ts int
	last_change := make(map[int]int)
	if_new := make(map[int]bool)

	log.Printf("Authorized on account TS_Bot")

	// The function of sending a reminder of the balance of ts at 18:00
	c := cron.New()
	c.AddFunc("CRON_TZ=Europe/Moscow 0 18 * * *", func() { 
		for key := range senders {
			balance := 480 - users[key]
			if balance != 0 {
				uresp := "Осталось списать: " + strconv.Itoa(balance) 
				b.Send(senders[key], uresp)
			}
		} 
		log.Println("Remaind sended")
	})
	c.Start()


	//Add buttons
	thirty_min := tb.InlineButton{
		Unique: "FM",
		Text:   "30 мин",
	}

	one_hour := tb.InlineButton{
		Unique: "OH",
		Text:   "1 час",
	}

	hour_and_half := tb.InlineButton{
		Unique: "HAH",
		Text:   "1,5 часа",
	}

	two_hour := tb.InlineButton{
		Unique: "TH",
		Text:   "2 часа",
	}

	delete_last_ts := tb.InlineButton{
		Unique: "DLT",
		Text:   "❌ Удалить последний TS",
	}

	delete_all_ts := tb.InlineButton{
		Unique: "DAT",
		Text:   "⭕ Обнулить",
	}

	BackToMain := tb.InlineButton{
		Unique: "BTM",
		Text:   "⬅️ Назад",
	}
	
	// Collect buttons on group
	mainInline := [][]tb.InlineButton{
		[]tb.InlineButton{thirty_min, one_hour},
		[]tb.InlineButton{hour_and_half, two_hour},
		[]tb.InlineButton{delete_last_ts, delete_all_ts},
	}

	helpInline := [][]tb.InlineButton{
		[]tb.InlineButton{BackToMain},
	}

	b.Handle("/start", func(m *tb.Message) {

			senders[m.Sender.ID] = m.Sender

			uresp := "Привет! 🗳️\nЯ могу считать списанные трудозатраты!\nДля учета трудозатрат пришли количество списанных минут или выбери подходящую кнопку."
			b.Send(m.Sender, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})

			b.Handle("/help", func(m *tb.Message) {

				log.Println(m.Sender.Username, ": ", m.Text)

				uresp := "Для счета трудозатрат достаточно просто отправить количество списанных минут или выбрать соответствующую кнопку.\nЕсли необходимо отнять трудозатраты, можно отправить отрицательное значение или нажать кнопку (❌ Удалить последний TS).\nЭта кнопка удалит последние учтенные ботом ваши трудозатраты.\nСтоит учитывать, что бот запоминает только последнее учтенное сообщение даже если его значение было отрицательным.\nP.s. В соседнем боте можно узнать о погоде - @pogo_go_bot."
				b.Send(m.Sender, uresp, &tb.ReplyMarkup{
					InlineKeyboard: helpInline,
				})
			})

			b.Handle(tb.OnText, func(m *tb.Message) {

				log.Println(m.Sender.Username, ": ", m.Text)
	
				newts, err := strconv.Atoi(m.Text)
				if err != nil {
					return
				}
				if_new[m.Sender.ID] = true

				users[m.Sender.ID], uresp = addTS(newts, m.Sender.ID, users, m.Sender)

				last_change[m.Sender.ID] = newts

				b.Send(m.Sender, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

			})

			b.Handle(&thirty_min, func(c *tb.Callback) {

				log.Println(c.Sender.Username, ": thirty_min")

				if_new[c.Sender.ID] = true

				users[c.Sender.ID], uresp = addTS(30, c.Sender.ID, users, c.Sender)

				last_change[c.Sender.ID] = 30

				b.Edit(c.Message, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

			})

			b.Handle(&one_hour, func(c *tb.Callback) {

				log.Println(c.Sender.Username, ": one_hour")

				if_new[c.Sender.ID] = true

				users[c.Sender.ID], uresp = addTS(60, c.Sender.ID, users, c.Sender)

				last_change[c.Sender.ID] = 60

				b.Edit(c.Message, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

			})

			b.Handle(&hour_and_half, func(c *tb.Callback) {

				log.Println(c.Sender.Username, ": hour_and_half")

				if_new[c.Sender.ID] = true

				users[c.Sender.ID], uresp = addTS(90, c.Sender.ID, users, c.Sender)

				last_change[c.Sender.ID] = 90

				b.Edit(c.Message, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

			})

			b.Handle(&two_hour, func(c *tb.Callback) {

				log.Println(c.Sender.Username, ": two_hour")

				if_new[c.Sender.ID] = true

				users[c.Sender.ID], uresp = addTS(120, c.Sender.ID, users, c.Sender)

				last_change[c.Sender.ID] = 120

				b.Edit(c.Message, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

			})

			b.Handle(&delete_last_ts, func(c *tb.Callback) {
				
				lts := last_change[c.Sender.ID]

				log.Println(c.Sender.Username, ": -", lts)

				if if_new[c.Sender.ID] == true {

					if_new[c.Sender.ID] = false
					users[c.Sender.ID] += -lts
					balance := 480 - users[c.Sender.ID]
					hour := users[c.Sender.ID] / 60
					hour_min := users[c.Sender.ID] - hour*60
					balance_hour := 8-hour
					balance_min := hour_min
					if balance_min != 0 {
						balance_hour += -1
						balance_min = 60-balance_min
					}
					uresp = "Списано: " + strconv.Itoa(users[c.Sender.ID]) + " Минут" + "\nВ часах: " + strconv.Itoa(hour) + ":" + strconv.Itoa(hour_min) + "\nОсталось: " + strconv.Itoa(balance) + "\nОсталось в часах: " + strconv.Itoa(balance_hour) + ":" + strconv.Itoa(balance_min)
					
					b.Edit(c.Message, uresp, &tb.ReplyMarkup{
						InlineKeyboard: mainInline,
					})
					
				} else {
					
					balance := 480 - users[c.Sender.ID]

					uresp := "⚠️ Последние трудозатраты уже отменены.\nДля вычета трудозатрат вы можете прислать отрицательное значение.\n" + "Списано: " + strconv.Itoa(users[c.Sender.ID]) + "\nОсталось: " + strconv.Itoa(balance)
					b.Edit(c.Message, uresp, &tb.ReplyMarkup{
						InlineKeyboard: mainInline,
					})
				}
				b.Respond(c, &tb.CallbackResponse{})
			})

			b.Handle(&delete_all_ts, func(c *tb.Callback) {
				
				log.Println(c.Sender.Username, ": -", users[c.Sender.ID])

				last_change[c.Sender.ID] = - users[c.Sender.ID]
				users[c.Sender.ID] = 0
				balance := 480 - users[c.Sender.ID]
				hour := users[c.Sender.ID] / 60
				hour_min := users[c.Sender.ID] - hour*60
				balance_hour := 8-hour
				balance_min := hour_min
				if balance_min != 0 {
					balance_hour += -1
					balance_min = 60-balance_min
				}
				uresp = "Списано: " + strconv.Itoa(users[c.Sender.ID]) + " Минут" + "\nВ часах: " + strconv.Itoa(hour) + ":" + strconv.Itoa(hour_min) + "\nОсталось: " + strconv.Itoa(balance) + "\nОсталось в часах: " + strconv.Itoa(balance_hour) + ":" + strconv.Itoa(balance_min)
				
				b.Edit(c.Message, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

				b.Respond(c, &tb.CallbackResponse{})
			})

			b.Handle(&BackToMain, func(c *tb.Callback) {
				balance := 480 - users[c.Sender.ID]
				hour := users[c.Sender.ID] / 60
				hour_min := users[c.Sender.ID] - hour*60
				balance_hour := 8-hour
				balance_min := hour_min
				if balance_min != 0 {
					balance_hour += -1
					balance_min = 60-balance_min
				}
				uresp = "Списано: " + strconv.Itoa(users[c.Sender.ID]) + " Минут" + "\nВ часах: " + strconv.Itoa(hour) + ":" + strconv.Itoa(hour_min) + "\nОсталось: " + strconv.Itoa(balance) + "\nОсталось в часах: " + strconv.Itoa(balance_hour) + ":" + strconv.Itoa(balance_min)
				
				b.Edit(c.Message, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline})
				b.Respond(c, &tb.CallbackResponse{})
			})
			
	})

	b.Start() // Start the bot
	
	c.Stop()  // Stop the scheduler (does not stop any jobs already running).
}