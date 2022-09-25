package rest_chats

import (
	auth_api "github.com/basicus/hla-course/grpc/auth"
	"github.com/basicus/hla-course/model"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
	"time"
)

func (s *Service) GetUserChats(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(int64)
	s.log.Infof("Request GetUserChats request_id %s", c.Params("requestid"))
	chats, err := s.storage.UserGetChats(c.UserContext(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get user chats problem", "data": err})
	}
	if chats == nil {
		chats = []model.Chat{}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "get user chats ok", "data": chats})
}

func (s *Service) ChatCreate(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(int64)
	s.log.Infof("Request ChatCreate request_id %s", c.Params("requestid"))
	chatCreate := new(model.ChatCreateDTO)
	if err := c.BodyParser(chatCreate); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})

	}
	chatCreate.Users = append(chatCreate.Users, userId)

	chat, err := s.storage.ChatCreate(c.UserContext(), chatCreate.Title, chatCreate.Users...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Chat create problem", "data": err})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Chat create ok", "data": chat})
}

func (s *Service) ChatPostMessage(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(int64)
	id := c.Params("id")
	s.log.Infof("Request ChatPostMessage request_id %s", c.Params("requestid"))
	var err error

	chatId, err := strconv.ParseInt(id, 10, 64)

	message := new(model.MessageSendDTO)
	if err := c.BodyParser(message); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}

	// If chat exists
	_, err = s.storage.GetChat(c.UserContext(), chatId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Chat not found", "data": err})
	}

	// Try to save message to chat
	messageSaved, err := s.storage.MessageSave(c.UserContext(), chatId, userId, time.Now(), message.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Create chat problem", "data": err})
	}

	// Get username
	userName, err := s.authApi.UserName(c.UserContext(), &auth_api.UserNameRequest{UserId: messageSaved.UserFrom})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get username problem", "data": err})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Message send ok", "data": model.MessageDTO{
		Id:       messageSaved.Id,
		UserFrom: userName.UserName,
		Date:     messageSaved.SendAt,
		Message:  messageSaved.Message,
	}})
}

func (s *Service) GetChatMessages(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(int64)
	id := c.Params("id")
	s.log.Infof("Request GetChatMessages request_id %s", c.Params("requestid"))
	var err error

	chatId, err := strconv.ParseInt(id, 10, 64)

	// If chat exists
	chat, err := s.storage.GetChat(c.UserContext(), chatId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Chat not found", "data": err})
	}

	// Check if user is participant

	var userIsParticipant bool
	paricipants, err := s.storage.ChatGetParticipants(c.UserContext(), chatId)
	for _, v := range paricipants {
		if v == userId {
			userIsParticipant = true
		}
	}

	if !userIsParticipant {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "User is not participant of this chat", "data": "not participant"})
	}

	// Try get messages for chat
	messageList, err := s.storage.ChatMessages(c.UserContext(), chatId, 0, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Create chat problem", "data": err})
	}

	m := model.MessageList{
		Chat: chat,
		List: make([]model.MessageDTO, len(messageList)),
	}

	// Get usernames and create dto for user
	users := make(map[int64]string, 10)
	for i := 0; i < len(messageList); i++ {
		_, ok := users[messageList[i].UserFrom]
		if !ok {
			// Get username
			userName, err := s.authApi.UserName(c.UserContext(), &auth_api.UserNameRequest{UserId: messageList[i].UserFrom})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get username problem", "data": err})
			}
			users[messageList[i].UserFrom] = userName.GetUserName()
		}
		m.List[i] = model.MessageDTO{
			Id:       messageList[i].Id,
			UserFrom: users[messageList[i].UserFrom],
			Date:     messageList[i].SendAt,
			Message:  messageList[i].Message,
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Message send ok", "data": m})
}

func (s *Service) Protected(c *fiber.Ctx) error {

	auth := c.Get("Authorization")
	authScheme := "Bearer"
	l := len(authScheme)
	token := ""
	if len(auth) > l+1 && strings.EqualFold(auth[:l], authScheme) {
		token = auth[l+1:]
	}
	checkSession, err := s.authApi.CheckSession(c.UserContext(), &auth_api.CheckSessionRequest{JwtToken: token})
	if err != nil || !checkSession.Ok {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired JWT")
	}

	c.Locals("user_id", checkSession.UserId)
	return c.Next()
}
