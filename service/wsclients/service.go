package wspusher

import (
	"context"
	"errors"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type Service struct {
	config          Config
	log             *logrus.Logger
	app             *fiber.App
	storage         *storage.UserService
	userConnections map[int64][]*websocket.Conn
	registerChan    chan WsInfo
	unregisterChan  chan WsInfo
	sendMessage     chan model.WsEvent
}

type WsInfo struct {
	UserId     int64
	Connection *websocket.Conn
}

func New(config Config, storage *storage.UserService, log *logrus.Logger) (*Service, error) {

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) { // Returns true if the client requested upgrade to the WebSocket protocol
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	s := &Service{
		config:          config,
		log:             log,
		app:             app,
		storage:         storage,
		userConnections: make(map[int64][]*websocket.Conn),
		registerChan:    make(chan WsInfo, 10),
		unregisterChan:  make(chan WsInfo, 10),
		sendMessage:     make(chan model.WsEvent),
	}
	app.Get("/ws/:id", websocket.New(s.WsConnectionHandler))
	return s, nil
}

func (s *Service) WsConnectionHandler(c *websocket.Conn) {
	id := c.Params("id")
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	wsInfo := WsInfo{
		UserId:     userId,
		Connection: c,
	}
	defer func() {
		// Send message to chan for closing connection
		s.unregisterChan <- wsInfo
		c.Close()
	}()

	// Register the client TODO strategy for force disconnect or use array of connections (and deliver to it)
	s.registerChan <- wsInfo

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.log.Errorf("connection user_id %d %+v read error: %s", wsInfo.UserId, wsInfo.Connection, err)
			}

			return // Closes deferred connection on close
		}

		if messageType == websocket.PingMessage || messageType == websocket.PongMessage {
			// TODO Ping/Pong handling
		}
		if messageType == websocket.TextMessage {
			// Broadcast the received message
			s.log.Infof("received text message %s for user_id %d connection %+v", string(message), wsInfo.UserId, wsInfo.Connection)
		} else {
			s.log.Infof("received unsupported message type %d for user_id %d connection %+v", messageType, wsInfo.UserId, wsInfo.Connection)
		}
	}
}

// SendEventToClient Send message for client
func (s *Service) SendEventToClient(event model.WsEvent) error {
	s.log.Infof("try to deliver message for user_id %d: %s", event.UserId, event.Message)
	cs, ok := s.userConnections[event.UserId]
	if !ok {
		s.log.Infof("cant deliver message for user_id %d: %s", event.UserId, event.Message)
		return nil
	}
	for _, c := range cs {
		err := c.WriteMessage(websocket.TextMessage, []byte(event.Message))

		if err != nil {
			s.log.Errorf("cant deliver message for user_id %d connection %+v: %s", event.UserId, c, err)
			continue
		}
		s.log.Errorf("success deliver message for user_id %d connection %+v", event.UserId, c)
	}
	return nil
}

func (s *Service) RegisterChanHandlers() {
	for {
		select {
		case connection := <-s.registerChan:
			_, ok := s.userConnections[connection.UserId]
			if !ok {
				cs := []*websocket.Conn{connection.Connection}

				s.userConnections[connection.UserId] = cs

			} else {
				s.userConnections[connection.UserId] = append(s.userConnections[connection.UserId], connection.Connection)
			}

			s.log.Infof("connection %+v registered", connection)
		case connection := <-s.unregisterChan:
			// Remove the client from the hub

			_, ok := s.userConnections[connection.UserId]
			if ok {
				for i, other := range s.userConnections[connection.UserId] {
					if other.Conn == connection.Connection.Conn {
						x := append(s.userConnections[connection.UserId][:i], s.userConnections[connection.UserId][i+1:]...)
						if len(x) > 0 {
							s.userConnections[connection.UserId] = x
						} else {
							delete(s.userConnections, connection.UserId)
						}
						break
					}
				}
				s.log.Infof("connection %+v unregistered", connection)
				continue
			}
			s.log.Errorf("connection %+v not found in registered", connection)
		}
	}

}

// Run Group task
func (s *Service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.WithField("address", s.config.Listen).Info("Start listening")
	defer func() {
		logger.Info("Stop listening")
	}()

	go s.RegisterChanHandlers()

	if err := s.app.Listen(s.config.Listen); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		logger.WithError(err).Error("Failed start listening")
		return err
	}

	return nil

}

// Shutdown Group task gracefull shutdown
func (s *Service) Shutdown(ctx context.Context) error {
	logger := log.Ctx(ctx)

	if err := s.app.Shutdown(); err != nil {
		logger.WithError(err).Error("Failed shutdown")
		return err
	}
	// TODO gracefull shutdown of all connections
	// TODO shutdown rabbitmq connection
	return nil
}
