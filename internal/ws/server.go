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
	clients    map[*Client]bool
	rooms      map[*Room]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	users      map[*models.User]bool
	sync.RWMutex
	redis          *redis.Client
	roomRepository repository.RoomRepository
	userRepository repository.UserRepository
}

func NewWsServer(redis *redis.Client, roomRepository repository.RoomRepository, userRepository repository.UserRepository) *WsServer {
	wsServer := &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[*Room]bool),
		// users:          make(map[*models.User]bool),
		roomRepository: roomRepository,
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

func (server *WsServer) findRoomByName(name string) *Room {
	server.Lock()
	defer server.Unlock()

	var room *Room
	for r := range server.rooms {
		if r.GetName() == name {
			room = r
			break
		}
	}

	if room == nil {
		room = server.runRoomFromRepository(name)
	}

	return room
}

func (server *WsServer) createRoom(name string, private bool) *Room {
	server.Lock()
	defer server.Unlock()

	room := NewRoom(name, private)
	go room.Run()
	server.rooms[room] = true

	return room
}

func (server *WsServer) runRoomFromRepository(name string) *Room {
	// server.Lock()
	// defer server.Unlock()

	var room *Room
	repoRoom, _ := server.roomRepository.GetRoomByName(ctx, name)
	if repoRoom != nil {
		room = NewRoom(repoRoom.GetName(), repoRoom.GetPrivate())
		room.ID, _ = uuid.Parse(repoRoom.GetId())
		go room.Run()
		server.rooms[room] = true
	}

	return room
}

func (server *WsServer) findRoomByID(ID string) *Room {
	server.RLock()
	defer server.RUnlock()

	var room *Room
	for r := range server.rooms {
		if r.GetId() == ID {
			room = r
			break
		}
	}

	return room
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
