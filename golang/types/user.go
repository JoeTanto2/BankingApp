package types

import (
	"banking-app/configs"
	"banking-app/models"
	"context"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

type User struct {
	Id      string `json:"id" binding:"required"`
	Name    string `json:"name"`
	Address string `json:"addres"`
	Age     int    `json:"age"`
	Country string `json:"country"`
	State   string `json:"state"`
}

func (u *User) UpdateUser() bson.M {
	update := bson.M{}
	if u.Name != "" {
		update["name"] = u.Name
	}
	if u.Address != "" {
		update["address"] = u.Address
	}
	if u.Age != 0 {
		update["age"] = u.Age
	}
	if u.Country != "" {
		update["country"] = u.Country
	}
	if u.State != "" {
		update["state"] = u.State
	}
	if len(update) == 0 {
		return nil
	}
	return bson.M{"$set": update}
}

type Login struct {
	Email    string `json:"email,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
}

type RegisterCard struct {
	Id       string `json:"id,omitempty"  validate:"required"`
	TypeCard int    `json:"type,omitempty" validage:"required"`
}

func (rc *RegisterCard) CardValidation() bool {
	if rc.TypeCard < 0 || rc.TypeCard > 1 {
		return false
	}
	return true
}

func (rc *RegisterCard) CardNumberGenerator(n int) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var letterRunes = []rune("1234567890")
	var card models.Card
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	err := userCollection.FindOne(ctx, bson.M{"number": string(b)}).Decode(&card)
	log.Println(card)
	if err == nil {
		log.Println("It has restarted")
		return rc.CardNumberGenerator(configs.CardNumberLength)
	}
	log.Println(string(b))
	return string(b)
}

func (rc *RegisterCard) CreateCard() models.Card {
	var card models.Card
	var history models.TransactionHistory
	now := time.Now()
	date := now.AddDate(4, 0, 0)
	history.ActionDate = now.Format(time.ANSIC)
	history.Action = "Card has been opened"
	options := []string{"Master Card", "Visa Card"}
	cardNumber := rc.CardNumberGenerator(configs.CardNumberLength)
	card.User = rc.Id
	card.Number = cardNumber
	card.SecurityCode = rc.CardNumberGenerator(configs.SecurityCodeLength)
	card.Network = options[rc.TypeCard]
	card.ExpirationDate = date.Format("2006-01-02")
	card.Balance = 0.0
	card.TransactionHistory = append(card.TransactionHistory, history)
	return card
}

type Deposit struct {
	Id     string  `json:"id,omitempty" validate:"required"`
	Number string  `json:"number,omitempty" validate:"required"`
	Amount float32 `json:"amount,omitempty" validate:"required"`
}

func (d *Deposit) Validate() {
	if d.Amount <= 0 {
		panic("Amount is less than minimal")
	}
	var card models.Card
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"number": d.Number}
	err := userCollection.FindOne(ctx, filter).Decode(&card)
	if err != nil {
		panic("You dont't have card with the number of " + d.Number)
	}
}

func (d *Deposit) DepositCash() int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := userCollection.UpdateOne(ctx, bson.M{"user": d.Id, "number": d.Number}, bson.M{"$set": bson.M{"balance": d.Amount}})
	if err != nil {
		return -1
	}
	return 1
}

type Transfer struct {
	From         string  `json:"from,omitempty" validae:"required"`
	To           string  `json:"to,omitempty" validate:"required"`
	Amount       float32 `json:"amount,omitempty" validate:"required"`
	SecurityCode string  `json:"security,omitempty" validate:"required"`
}

func (t *Transfer) Validate() int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var card models.Card
	err := userCollection.FindOne(ctx, bson.M{"number": t.From}).Decode(&card)
	if err != nil {
		return -1
	}
	if t.SecurityCode != card.SecurityCode {
		return -1
	}
	if t.Amount > card.Balance {
		return -1
	}
	err = userCollection.FindOne(ctx, bson.M{"number": t.To}).Decode(&card)
	if err != nil {
		return -1
	}
	return 1
}

func (t *Transfer) Finalize() int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var cardFrom models.Card
	var cardTo models.Card
	_ = userCollection.FindOne(ctx, bson.M{"number": t.From}).Decode(&cardFrom)
	_ = userCollection.FindOne(ctx, bson.M{"number": t.To}).Decode(&cardTo)
	cardFromNewAmount := cardFrom.Balance - t.Amount
	cardToNewAmount := cardTo.Balance + t.Amount
	_, err := userCollection.UpdateOne(ctx, bson.M{"number": t.From}, bson.M{"$set": bson.M{"balance": cardFromNewAmount}})
	if err != nil {
		return -1
	}
	_, err = userCollection.UpdateOne(ctx, bson.M{"number": t.To}, bson.M{"$set": bson.M{"balance": cardToNewAmount}})
	if err != nil {
		return -1
	}
	return 1
}
