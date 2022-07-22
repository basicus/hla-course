package handlers

import (
	"fmt"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Handlers struct {
	Logger  *logrus.Logger
	Storage storage.UserService
	Config  Config
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

	validated, err := h.Storage.ValidateUser(c.UserContext(), login, pass)
	if err != nil || !validated {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Error on login", "data": "User not exist or invalid password"})
	}

	user, err := h.Storage.GetByLogin(c.UserContext(), login)
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

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "data": nil, "token": token})

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

func jwtToken(user model.User, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Login
	claims["user_id"] = user.UserId
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	return token.SignedString([]byte(secret))
}
