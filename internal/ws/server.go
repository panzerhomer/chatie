package ws

import (
	"chatie/internal/apperror"
	"chatie/internal/models"
	"chatie/internal/repository"
	"chatie/internal/services"
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/redis/go-redis/v9"
)

const PubSubGeneralChannel = "general"

var ctx = context.Background()

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	wsChats    map[*WsChat]bool
	users      []models.User
	// userService services.UserServices
	chatRepository repository.ChatRepository
	userRepository services.UserRepository
	redis          *redis.Client
	// sync.RWMutex
}

// NewWebsocketServer creates a new WsServer type
func NewWsServer(
	chatRepository repository.ChatRepository,
	userRepository services.UserRepository,
	redis *redis.Client,
) *WsServer {
	wsServer := &WsServer{
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan []byte),
		wsChats:        make(map[*WsChat]bool),
		chatRepository: chatRepository,
		userRepository: userRepository,
		redis:          redis,
	}

	var err error
	wsServer.users, err = userRepository.GetAllUsers(context.Background())
	if err != nil {
		if errors.Is(err, apperror.ErrUserNotFound) {
			log.Println("", err)
		} else {
			log.Println("can't extract users from database: ", err)
		}
	}

	return wsServer
}

func (server *WsServer) Run() {
	go server.listenPubSubChannel()
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
			// case message := <-server.broadcast:
			// 	server.broadcastToClients(message)
		}
	}
}

func (server *WsServer) publishClientJoined(client *Client) {
	message := &WebsocketMessage{
		Action: UserJoinedAction,
		Sender: clientToUser(client),
	}

	if err := server.redis.Publish(ctx, PubSubGeneralChannel, message.encode()).Err(); err != nil {
		log.Println(err)
	}
}

func (server *WsServer) publishClientLeft(client *Client) {
	message := &WebsocketMessage{
		Action: UserLeftAction,
		Sender: clientToUser(client),
	}

	if err := server.redis.Publish(ctx, PubSubGeneralChannel, message.encode()).Err(); err != nil {
		log.Println(err)
	}
}

func (server *WsServer) listenPubSubChannel() {
	pubsub := server.redis.Subscribe(ctx, PubSubGeneralChannel)

	ch := pubsub.Channel()

	for msg := range ch {
		var message WebsocketMessage
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("Error on unmarshal JSON message %s", err)
			return
		}

		switch message.Action {
		case UserJoinedAction:
			server.handleUserJoined(message)
		case UserLeftAction:
			server.handleUserLeft(message)
		case JoinChatPrivateAction:
			server.handleUserJoinPrivate(message)
		}
	}
}

func (server *WsServer) handleUserJoined(message WebsocketMessage) {
	// Add the user to the slice
	if message.Sender != nil {
		server.users = append(server.users, *message.Sender)
	}
	// server.users = append(server.users, *message.Sender)
	server.broadcastToClients(message.encode())
}

func (server *WsServer) handleUserLeft(message WebsocketMessage) {
	for i, user := range server.users {
		if message.Sender != nil && user.GetID() == message.Sender.GetID() {
			server.users[i] = server.users[len(server.users)-1]
			server.users = server.users[:len(server.users)-1]
		}
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) handleUserJoinPrivate(message WebsocketMessage) {
	targetClient := server.findClientByName(message.Target)
	if targetClient != nil {
		client := server.findClientByName(message.Sender.Name)
		targetClient.joinChat(message.Target, client)
	}
}

func (server *WsServer) registerClient(client *Client) {
	server.userRepository.InsertUser(ctx, clientToUser(client))

	// server.notifyClientJoined(client)
	// Publish user in PubSub
	server.publishClientJoined(client)
	// server.listOnlineClients(client)
	server.clients[client] = true
	log.Println("ws server join", server.clients)
}

func (server *WsServer) unregisterClient(client *Client) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
		// server.notifyClientLeft(client)
		server.publishClientLeft(client)
	}
	log.Println("ws server left", server.clients)
}

func (server *WsServer) notifyClientJoined(client *Client) {
	message := &WebsocketMessage{
		Action: UserJoinedAction,
		Sender: clientToUser(client),
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) notifyClientLeft(client *Client) {
	message := &WebsocketMessage{
		Action: UserLeftAction,
		Sender: clientToUser(client),
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) listOnlineClients(client *Client) {
	for existingClient := range server.clients {
		message := &WebsocketMessage{
			Action: UserJoinedAction,
			Sender: clientToUser(existingClient),
		}
		client.send <- message.encode()
	}
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		client.send <- message
	}
}

func (server *WsServer) findChatByName(name string) *WsChat {
	var chat *WsChat
	for c := range server.wsChats {
		if c.GetName() == name {
			chat = c
			break
		}
	}

	return chat
}

func (server *WsServer) findRoomByID(ID string) *WsChat {
	var chat *WsChat
	for c := range server.wsChats {
		if c.GetID() == ID {
			chat = c
			break
		}
	}

	return chat
}

func (server *WsServer) createChat(name string, private bool) *WsChat {
	chat := NewChat(server, name, private)
	go chat.Run()
	server.wsChats[chat] = true

	return chat
}

func (server *WsServer) findClientByID(ID string) *Client {
	var foundClient *Client
	for client := range server.clients {
		if client.GetID() == ID {
			foundClient = client
			break
		}
	}

	return foundClient
}

func (server *WsServer) findClientByName(ID string) *Client {
	var foundClient *Client
	for client := range server.clients {
		if client.GetName() == ID {
			foundClient = client
			break
		}
	}

	return foundClient
}

func clientToUser(client *Client) *models.User {
	user := &models.User{
		// BaseModel: models.BaseModel{
		// 	ID: client.GetID(),
		// },
		Name: client.GetName(),
	}
	return user
}
