package helper

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JWTUser struct {
	Id   int64
	Name string
}

func GetUser(c *fiber.Ctx) (*JWTUser, error) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("could not find user in context")
	}
	return GetUserFromToken(user)
}

func GetUserFromToken(token *jwt.Token) (*JWTUser, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("could not get claims")
	}
	var name string
	var id int64

	switch claims["name"].(type) {
	case string:
		name = claims["name"].(string)
	default:
		return nil, fmt.Errorf("could not get name")
	}

	switch claims["id"].(type) {
	case int64:
		id = claims["id"].(int64)
	case float64:
		id = int64(claims["id"].(float64))
	case string:
		i, err := strconv.ParseInt(claims["id"].(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not get id: %w", err)
		}
		id = i
	}
	return &JWTUser{
		Id:   id,
		Name: name,
	}, nil
}
