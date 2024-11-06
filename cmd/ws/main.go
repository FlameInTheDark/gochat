package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/middleware/logger"
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/FlameInTheDark/gochat/cmd/ws/config"
	"github.com/FlameInTheDark/gochat/internal/helper"
)

var rabbitConn *amqp.Connection

type Message struct {
	ID          int64        `json:"id"`
	ChannelID   int64        `json:"channel_id"`
	AuthorID    Author       `json:"author_id"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
}

type Author struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Discriminator string `json:"discriminator"`
}

type Attachment struct {
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
	Height      *int   `json:"height"`
	Width       *int   `json:"width"`
	URL         string `json:"url"`
	Size        int    `json:"size"`
}

func publishMessage(msg Message) error {
	ch, err := rabbitConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"gochat.messages", // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return err
	}

	routingKey := fmt.Sprintf("channel.%d", msg.ChannelID) // use channel_id as the routing key
	messageBody, _ := json.Marshal(msg)

	return ch.Publish(
		"gochat.messages", // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
}

func wsHandler(c *websocket.Conn) {
	if !c.Locals("allowed").(bool) {
		err := c.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}
	channelId := c.Locals("channel_id").(int64)
	//user := c.Locals("user").(helper.JWTUser)
	defer func() {
		err := c.Close()
		if err != nil && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			log.Println("Error closing WebSocket:", err)
		}
	}()

	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Println("Failed to open channel:", err)
		return
	}

	defer func() {
		err := ch.Close()
		if err != nil {
			log.Println("Error closing RabbitMQ channel:", err)
		}
	}()

	err = ch.ExchangeDeclare(
		"gochat.messages",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to declare exchange:", err)
		return
	}

	queue, err := ch.QueueDeclare(
		"",    // let RabbitMQ generate a unique name
		true,  // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Println("Failed to declare queue:", err)
		return
	}

	err = ch.QueueBind(
		queue.Name,                           // queue name
		fmt.Sprintf("channel.%d", channelId), // routing key (channel ID)
		"gochat.messages",                    // exchange
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to bind queue:", err)
		return
	}

	go func() {
		msgs, err := ch.Consume(
			queue.Name,
			"",
			true,  // auto-ack
			true,  // exclusive
			false, // no-local
			false, // no-wait
			nil,
		)
		if err != nil {
			log.Println("Failed to register consumer:", err)
			return
		}

		for msg := range msgs {
			err := c.WriteMessage(websocket.TextMessage, msg.Body)
			if err != nil {
				log.Println("WebSocket send failed:", err)
				return
			}
		}
	}()

	var (
		mt  int
		msg []byte
	)

	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			continue
		}
		switch mt {
		case websocket.TextMessage:
			log.Printf("Received text message: %s", msg)
			// Handle text message (e.g., parse JSON)

		case websocket.BinaryMessage:
			log.Printf("Received binary message of length %d", len(msg))
			// Handle binary data (e.g., process file or image data)

		case websocket.PingMessage:
			// Optionally respond with a pong message
			err := c.WriteMessage(websocket.PongMessage, nil)
			if err != nil {
				log.Println("Error responding to ping:", err)
			}

		case websocket.PongMessage:
			log.Println("Received pong")
			// Optional: track pong responses for connection health

		case websocket.CloseMessage:
			// Respond with a close message and close the connection gracefully
			closeCode := websocket.CloseNormalClosure
			closeMessage := "Closing connection as requested"
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, closeMessage))
			if err != nil {
				log.Println("Failed to send close response:", err)
			}
			// Exit the loop after sending the close response
			return
		}
	}

}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	rabbitConn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.RabbitMQUsername, cfg.RabbitMQPassword, cfg.RabbitMQHost, cfg.RabbitMQPort))
	if err != nil {
		logger.Error("unable to connect to RabbitMQ", slog.String("error", err.Error()))
		os.Exit(1)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(flog.New())
	app.Use(recm.New())
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(cfg.AuthSecret)},
	}))

	app.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", false)
			channelID := c.Query("channel_id")
			id, err := strconv.ParseInt(channelID, 10, 64)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "unable to parse channel id")
			}
			user, err := helper.GetUser(c)
			if err != nil {
				return fiber.NewError(fiber.StatusUnauthorized, "unable to get user")
			}
			c.Locals("user", user)
			c.Locals("channel_id", id)
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	wscfg := websocket.Config{
		RecoverHandler: func(conn *websocket.Conn) {
			if err := recover(); err != nil {
				err := conn.WriteJSON(fiber.Map{"customError": "error occurred"})
				if err != nil {
					logger.Error("failed to send error", slog.String("error", err.Error()))
				}
			}
		},
	}
	app.Get("/subscribe", websocket.New(wsHandler, wscfg))

	log.Println("Server starting on :3100")
	log.Fatal(app.Listen(":3100"))
}
