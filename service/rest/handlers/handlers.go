package handlers

import (
	"fmt"
	chat_api "github.com/basicus/hla-course/grpc/chats"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/service/queue"
	"github.com/basicus/hla-course/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"time"
)

type Handlers struct {
	Logger      *logrus.Logger
	Storage     storage.UserService
	AuthService storage.UserService
	Config      Config
	Queue       *queue.Service
	ChatApi     chat_api.ChatServiceClient
}

// Register Регистрация пользователя
func (h *Handlers) Register(c *fiber.Ctx) error {
	user := new(model.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})

	}

	err := user.ValidateRegister()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err.Error()})
	}
	oldUser, err := h.Storage.GetByLogin(c.UserContext(), user.Login)

	if err == nil && user.Login == oldUser.Login {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Duplicate login", "data": nil})
	}

	newUser, err := h.Storage.Create(c.UserContext(), *user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Couldn't create user", "data": err})
	}

	newUser.Settings = model.UserSettings{WSConnect: strings.Replace(h.Config.WsUserString, "{RK}", newUser.ShardId, 1)}

	token, err := jwtToken(newUser, h.Config.JwtSecret)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Created user", "data": newUser, "token": token})
}

func (h *Handlers) Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var input LoginInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on login request", "data": err})
	}
	login := input.Login
	pass := input.Password
	var validator storage.UserService

	if h.AuthService != nil {
		validator = h.AuthService
	} else {
		validator = h.Storage
	}
	validated, err := validator.ValidateUser(c.UserContext(), login, pass)
	if err != nil || !validated {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Error on login", "data": "User not exist or invalid password"})
	}

	user, err := validator.GetByLogin(c.UserContext(), login)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Error on username", "data": err})
	}

	token, err := jwtToken(user, h.Config.JwtSecret)
	if err != nil {
		return err
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "data": model.UserSettings{WSConnect: strings.Replace(h.Config.WsUserString, "{RK}", user.ShardId, 1)}, "token": token})

}

func (h *Handlers) UsersGet(c *fiber.Ctx) error {
	//user := c.Locals("user").(*jwt.Token)
	//claims := user.Claims.(jwt.MapClaims)
	//userId := int64(claims["user_id"].(float64))
	//search := c.Params("search")
	users, err := h.Storage.GetUsers(c.UserContext(), map[string]string{"name": "Ser%"},
		map[string]string{"user_id": "ASC"}, 0, 1000)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on get user list", "data": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Success user listing", "data": users})
}

func (h *Handlers) UpdateProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	//userName := claims["username"].(string)

	userUpdate := new(model.User)
	if err := c.BodyParser(userUpdate); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})

	}
	userUpdate.UserId = userId

	upd, err := userUpdate.ValidateUpdate()
	update, err := h.Storage.Update(c.UserContext(), *userUpdate, upd)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Update problem", "data": err})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Update ok", "data": update})
}

func (h *Handlers) UserInfo(c *fiber.Ctx) error {
	id := c.Params("id")
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	var userId int64
	var err error
	if id != "" {
		userId, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "User id must be number", "data": err})
		}
	} else {
		userId = int64(claims["user_id"].(float64))
	}

	userData, err := h.Storage.GetById(c.UserContext(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Cant get user", "data": err})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Get User data ok ", "data": userData})

}

func (h *Handlers) AddFriend(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	id := c.Params("id")
	var friendId int64
	var err error

	friendId, err = strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Friend id must be number", "data": err})
	}
	status, err := h.Storage.AddFriend(c.UserContext(), userId, friendId)
	if err != nil || !status {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Friend add error", "data": err})
	}
	// Обновляем ленту постов после добавления друга (добавляем в очередь)
	_ = h.Queue.UpdateFeed(c.UserContext(), userId)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Friend add successfully", "data": nil})
}

func (h *Handlers) DeleteFriend(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	id := c.Params("id")
	var friendId int64
	var err error

	friendId, err = strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Friend id must be number", "data": err})
	}
	status, err := h.Storage.DelFriend(c.UserContext(), userId, friendId)
	if err != nil || !status {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Friend del error", "data": err})
	}

	// Обновляем ленту постов после удаления друга (добавляем в очередь)
	_ = h.Queue.UpdateFeed(c.UserContext(), userId)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Friend del successfully", "data": nil})
}

func (h *Handlers) GetFriends(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))

	friends, err := h.Storage.GetFriends(c.UserContext(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get friends error", "data": err})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Friend del successfully", "data": friends})
}

func (h *Handlers) SearchProfile(c *fiber.Ctx) error {
	//user := c.Locals("user").(*jwt.Token)
	//claims := user.Claims.(jwt.MapClaims)
	//userId := int64(claims["user_id"].(float64))
	searchFirst := c.Query("first")
	searchSecond := c.Query("second")
	searchLimit := c.Query("limit")
	searchOffset := c.Query("offset")
	searchOrder := c.Query("order")

	orderBy := make(map[string]string)

	offset := 0
	limit := 0

	if searchLimit != "" {
		atoi, err := strconv.Atoi(searchLimit)
		if err != nil {
			limit = 0
		} else {
			limit = atoi
		}
	}
	if searchOrder != "" {
		orderBy[searchOrder] = "ASC"
	} else {
		orderBy["user_id"] = "ASC"
	}

	if searchOffset != "" {
		atoi, err := strconv.Atoi(searchOffset)
		if err != nil {
			offset = 0
		} else {
			offset = atoi
		}
	}

	filter := make(map[string]string)
	if searchSecond != "" {
		filter["surname"] = fmt.Sprintf("%s%s", searchSecond, "%")
	}
	if searchFirst != "" {
		filter["name"] = fmt.Sprintf("%s%s", searchFirst, "%")
	}

	users, err := h.Storage.GetUsers(c.UserContext(), filter,
		orderBy, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on get user list", "data": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Success user listing", "data": users})
}

// PublishPost Публикация новой записи
func (h *Handlers) PublishPost(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	//userName := claims["username"].(string)

	post := new(model.PostPojo)
	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})

	}
	post.UserId = userId

	savedPost, err := h.Storage.PublishPost(c.UserContext(), post.UserId, post.Title, post.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Save post problem", "data": err})
	}

	// Добавляем в очередь обновление лент пользователей после добавления нового поста
	_ = h.Queue.NewPost(c.UserContext(), savedPost)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Publish ok", "data": savedPost})
}

func (h *Handlers) PersonalFeed(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))

	friendsPosts, err := h.Storage.GetFriendsPosts(c.UserContext(), userId, h.Config.PostsLimit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get timeline problem", "data": err})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "get feed ok", "data": friendsPosts})
}

// GetPostById Получение поста по его Id
func (h *Handlers) GetPostById(c *fiber.Ctx) error {
	id := c.Params("id")
	var postId int64
	var err error

	postId, err = strconv.ParseInt(id, 10, 64)

	post, err := h.Storage.GetPostById(c.UserContext(), postId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get post problem", "data": err})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "get ok", "data": post})
}

func (h *Handlers) GetUserChats(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	requestId := c.Locals("requestid").(string)
	userChatsResponse, err := h.ChatApi.ListChats(c.UserContext(), &chat_api.ListUserChatsRequest{UserId: userId, RequestId: requestId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get user chats problem", "data": err})
	}
	var chats []model.Chat
	for _, chat := range userChatsResponse.GetChats() {
		chats = append(chats, convertChatInfo2Chat(chat))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "get user chats ok", "data": chats})
}
func convertChatInfo2Chat(chat *chat_api.ChatInfo) model.Chat {
	return model.Chat{
		Id:        chat.GetChatId(),
		Title:     chat.GetTitle(),
		CreatedAt: chat.GetCreatedAt().AsTime(),
		Closed:    chat.GetClosed(),
	}
}

func (h *Handlers) ChatCreate(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	requestId := c.Locals("requestid").(string)
	chatCreate := new(model.ChatCreateDTO)
	if err := c.BodyParser(chatCreate); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})

	}
	chatCreate.Users = append(chatCreate.Users, userId)

	createChatResponse, err := h.ChatApi.CreateChat(c.UserContext(), &chat_api.CreateChatRequest{
		UserId:    userId,
		Title:     chatCreate.Title,
		Users:     chatCreate.Users,
		RequestId: requestId,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Chat create problem", "data": err})
	}
	chat := convertChatInfo2Chat(createChatResponse.Chat)

	// TODO Отправка в очередь уведомления о новом событии
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Chat create ok", "data": chat})
}

func (h *Handlers) ChatPostMessage(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	id := c.Params("id")
	requestId := c.Locals("requestid").(string)

	var err error

	chatId, err := strconv.ParseInt(id, 10, 64)

	message := new(model.MessageSendDTO)
	if err := c.BodyParser(message); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}

	// If chat exists
	_, err = h.ChatApi.Get(c.UserContext(), &chat_api.GetChatRequest{ChatId: chatId, RequestId: requestId})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Chat not found", "data": err})
	}

	// Try to save message to chat
	messageSaved, err := h.ChatApi.PostMessage(c.UserContext(), &chat_api.PostMessageRequest{
		UserId:    userId,
		ChatId:    chatId,
		Message:   message.Message,
		Date:      timestamppb.Now(),
		RequestId: requestId,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Create chat problem", "data": err})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get username problem", "data": err})
	}

	// TODO Отправка в очередь уведомления о новом событии
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Message send ok", "data": model.MessageDTO{
		Id:       messageSaved.Message.MessageId,
		UserFrom: messageSaved.Message.UserFrom,
		Date:     messageSaved.Message.Date.AsTime(),
		Message:  messageSaved.Message.Message,
	}})
}

func (h *Handlers) GetChatMessages(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := int64(claims["user_id"].(float64))
	id := c.Params("id")
	requestId := c.Locals("requestid").(string)

	var err error

	chatId, err := strconv.ParseInt(id, 10, 64)

	// If chat exists
	chat, err := h.ChatApi.Get(c.UserContext(), &chat_api.GetChatRequest{ChatId: chatId})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Chat not found", "data": err})
	}

	// Check if user is participant

	var userIsParticipant bool
	participants := chat.GetUsers()
	for _, v := range participants {
		if v == userId {
			userIsParticipant = true
		}
	}

	if !userIsParticipant {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "User is not participant of this chat", "data": "not participant"})
	}

	// Try get messages for chat
	messageList, err := h.ChatApi.Messages(c.UserContext(), &chat_api.ChatMessagesRequest{
		UserId:    userId,
		ChatId:    chatId,
		RequestId: requestId,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Get messages of chat problem", "data": err})
	}

	m := model.MessageList{
		Chat: convertChatInfo2Chat(chat.Chat),
		List: make([]model.MessageDTO, len(messageList.GetMessages())),
	}
	for i := 0; i < len(messageList.GetMessages()); i++ {
		m.List[i] = model.MessageDTO{
			Id:       messageList.Messages[i].MessageId,
			UserFrom: messageList.Messages[i].UserFrom,
			Date:     messageList.Messages[i].Date.AsTime(),
			Message:  messageList.Messages[i].Message,
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Message send ok", "data": m})
}

func jwtToken(user model.User, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Login
	claims["user_id"] = user.UserId
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	return token.SignedString([]byte(secret))
}
