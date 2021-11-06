package controllers

import (
	"banking-app/configs"
	"banking-app/models"
	"banking-app/services"
	"banking-app/types"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var form models.User
		defer cancel()
		if err := c.BindJSON(&form); err != nil {
			c.JSON(400,
				gin.H{
					"message": "please make sure all of the fields are filled",
				})
			return
		}
		if form.Age < 18 {
			c.JSON(
				401,
				gin.H{
					"message": "You must be at least 18 years old to open an account with us",
				})
			return
		}
		form.Password, _ = services.HashPassword(form.Password)
		result, err := userCollection.InsertOne(ctx, form)
		if err != nil {
			c.JSON(
				500,
				gin.H{
					"message": "something went wrong!",
				})
			return
		}
		c.JSON(
			201,
			gin.H{
				"message": "User has been successfuly created",
				"data":    result,
			})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var login types.Login
		var user models.User
		defer cancel()
		err := c.ShouldBind(&login)
		if err != nil {
			c.JSON(
				400,
				gin.H{
					"message": "Make sure you have entered email and password",
				})
		}
		err = userCollection.FindOne(ctx, bson.M{"email": login.Email}).Decode(&user)
		if err != nil {
			c.JSON(
				404,
				gin.H{
					"message": "User has not been found",
				})
		}
		hashedPassword := services.CheckPasswordHash(login.Password, user.Password)
		if hashedPassword {
			c.JSON(
				200,
				gin.H{
					"message": "you have been succesfully logged in",
				})
		} else {
			c.JSON(
				401,
				gin.H{
					"message": "make sure you have entered correct email and password",
				})
		}
	}
}

func UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var user types.User
		defer cancel()
		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.JSON(
				500,
				gin.H{
					"message": "something went wrong",
				})
			return
		}
		userId, _ := primitive.ObjectIDFromHex(user.Id)
		filter := bson.M{"_id": userId}
		update := user.UpdateUser()
		if update == nil {
			c.JSON(
				400,
				gin.H{
					"message": "There're no fields to update, please make sure you entered fields",
				})
		}
		result, err := userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(
				400,
				gin.H{
					"message": "something went wrong",
				})
			return
		}
		var updatedUser types.User
		if result.MatchedCount == 1 {
			err := userCollection.FindOne(ctx, filter).Decode(&updatedUser)
			if err != nil {
				c.JSON(
					404,
					gin.H{
						"message": "updated user has not been found",
					})
			}
		}
	}
}

func OpenCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var rc types.RegisterCard
		defer cancel()
		err := c.ShouldBind(&rc)
		if err != nil {
			c.JSON(
				400,
				gin.H{
					"message": "Please make sure you filled out all of the fields",
				})
			return
		}
		userId, _ := primitive.ObjectIDFromHex(rc.Id)
		filter := bson.M{"_id": userId}
		var usr models.User
		err = userCollection.FindOne(ctx, filter).Decode(&usr)
		if err != nil {
			c.JSON(
				404,
				gin.H{
					"message": "No user found by this ID",
				})
			return
		}
		result := rc.CardValidation()
		if !result {
			c.JSON(
				400,
				gin.H{
					"message": "Please select a valid card option between 0 - Master Card, 1 - Visa",
				})
			return
		}
		var user models.Card = rc.CreateCard()
		res, error := userCollection.InsertOne(ctx, user)
		if error != nil {
			c.JSON(
				500,
				gin.H{
					"message": error,
				})
			return
		}
		c.JSON(
			201,
			gin.H{
				"message": "You have succesfully opened the card",
				"data":    res,
			})
	}
}

func Deposit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var d types.Deposit
		var card models.Card
		defer cancel()
		if err := c.ShouldBind(&d); err != nil {
			c.JSON(
				400,
				gin.H{
					"message": "Please make sure you provided: card number and the amount correctly",
				})
			return
		};
		d.Validate()
		response := d.DepositCash()
		if response == -1 {
			c.JSON(
				400,
				gin.H{
					"message": "Bad request, please make sure you are deposeting to the right account",
				})
			return
		}
		if err := userCollection.FindOne(ctx, bson.M{"number": d.Number}).Decode(&card); err != nil {
			c.JSON(
				404,
				gin.H{
					"message": err.Error(),
				})
			return
		}
		c.JSON(
			200,
			gin.H{
				"card number":      card.Number,
				"deposited amount": d.Amount,
				"new balance":      card.Balance,
			})
	}
}

func Transfer() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var t types.Transfer
		err := c.ShouldBind(&t)
		if err != nil {
			c.JSON(
				400,
				gin.H{
					"message": "Please make sure you filled all of the fields",
				})
			return
		}
		validate := t.Validate()
		if validate == -1 {
			c.JSON(
				400,
				gin.H{
					"message": "Please make sure you entered the right card info and have sufficient balance",
				})
			return
		}
		response := t.Finalize()
		if response == -1 {
			c.JSON(
				400,
				gin.H{
					"message": "Please make sure you entered the right card info and have sufficient balance!!!",
				})
			return
		}
		c.JSON(
			200,
			gin.H{
				"message": "You have successfully transferred the money",
			})
	}
}
