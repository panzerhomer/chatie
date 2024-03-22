package ws

import (
	"chatie/internal/models"
	"chatie/internal/repository"
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type WsServer struct {
	clients        map[*Client]bool
	chats          map[*Chat]bool
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	users          map[*models.User]bool
	redis          *redis.Client
	chatRepository repository.ChatRepository
	userRepository repository.UserRepository
	sync.RWMutex
}

func NewWsServer(redis *redis.Client, chatRepository repository.ChatRepository, userRepository repository.UserRepository) *WsServer {
	wsServer := &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		chats:      make(map[*Chat]bool),
		// users:          make(map[*models.User]bool),
		chatRepository: chatRepository,
		userRepository: userRepository,
		redis:          redis,
	}

	repoUsers, _ := wsServer.userRepository.GetAllUsers(ctx)
	wsServer.users = make(map[*models.User]bool, len(repoUsers))
	for i := 0; i < len(repoUsers); i++ {
		wsServer.users[&repoUsers[i]] = true
	}

	return wsServer
}

func (server *WsServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

func clientToUser(c *Client) *models.User {
	user := models.User{
		BaseModel: models.BaseModel{
			ID: c.ID.String(),
		},
		Name: c.Name,
	}
	return &user
}

func (server *WsServer) registerClient(client *Client) {
	server.Lock()
	defer server.Unlock()

	newUser := clientToUser(client)
	server.userRepository.AddUser(ctx, newUser)
	server.clients[client] = true
	server.users[newUser] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	server.Lock()
	defer server.Unlock()

	deleteUser := clientToUser(client)
	server.userRepository.RemoveUser(ctx, deleteUser)
	delete(server.users, deleteUser)
	delete(server.clients, client)
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(server.clients, client)
		}
	}
}

func (server *WsServer) showOnlineClients(client *Client) {
	for registeredClient := range server.users {
		message := &Event{
			Type: UserJoinedEvent,
			Payload: &WsMessage{
				From: registeredClient,
			},
		}
		client.send <- message.encode()
	}
}

func (server *WsServer) findChatByName(name string) *Chat {
	server.Lock()
	defer server.Unlock()

	var chat *Chat
	for c := range server.chats {
		if c.GetName() == name {
			chat = c
			break
		}
	}

	if chat == nil {
		chat = server.runChatFromRepository(name)
	}

	return chat
}

func (server *WsServer) createChat(name string, private bool) *Chat {
	server.Lock()
	defer server.Unlock()

	chat := NewChat(name, private)
	go chat.Run()
	server.chats[chat] = true

	return chat
}

func (server *WsServer) runChatFromRepository(name string) *Chat {
	var chat *Chat
	repoChat, _ := server.chatRepository.GetChatByName(ctx, name)
	if repoChat != nil {
		chat = NewChat(repoChat.GetName(), repoChat.GetPrivate())
		chat.ID, _ = uuid.Parse(repoChat.GetId())
		go chat.Run()
		server.chats[chat] = true
	}
	return chat
}

func (server *WsServer) findChatByID(ID string) *Chat {
	server.RLock()
	defer server.RUnlock()

	var chat *Chat
	for c := range server.chats {
		if c.GetId() == ID {
			chat = c
			break
		}
	}
	return chat
}

func (server *WsServer) findClientByName(name string) *Client {
	server.RLock()
	defer server.RUnlock()

	var client *Client
	for c := range server.clients {
		if c.Name == name {
			client = c
			break
		}
	}
	return client
}

func (server *WsServer) findClientByID(ID string) *Client {
	server.RLock()
	defer server.RUnlock()

	var client *Client
	for c := range server.clients {
		if c.ID.String() == ID {
			client = c
			break
		}
	}
	return client
}
