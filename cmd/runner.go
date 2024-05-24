package main

import (
	"bytes"
	"chatie/internal/models"
	cry "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func getRequest(i int, url string, data []byte, users []models.User, psw string, wg *sync.WaitGroup) {
	defer wg.Done()
	// delay := time.Duration(rand.Intn(100)+50) * time.Millisecond
	// time.Sleep(delay)
	start := time.Now()

	client := &http.Client{}
	// client.Jar.SetCookies()
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error making POST request to", url, ":", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ошибка при чтении ответа:", err)
		return
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		fmt.Println("ошибка при разборе JSON:", err)
		return
	}

	users[i] = user
	users[i].Password = psw

	end := time.Now()

	fmt.Println("ответ от сервера:", users[i], psw, end.Sub(start).Milliseconds())
}

func postRequest(i int, url string, data []byte, users []models.User, psw string, wg *sync.WaitGroup) {
	defer wg.Done()
	// delay := time.Duration(rand.Intn(100)+50) * time.Millisecond
	// time.Sleep(delay)
	start := time.Now()

	client := &http.Client{}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error making POST request to", url, ":", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ошибка при чтении ответа:", err)
		return
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		fmt.Println("ошибка при разборе JSON:", err)
		return
	}

	users[i] = user
	users[i].Password = psw

	end := time.Now()

	fmt.Println("ответ от сервера:", users[i], psw, end.Sub(start).Milliseconds())
}

func postLogin(i int, url string, data []byte, userTokens []tokenResponse, wg *sync.WaitGroup) {
	defer wg.Done()
	// delay := time.Duration(rand.Intn(100)+50) * time.Millisecond
	// time.Sleep(delay)
	start := time.Now()

	client := &http.Client{}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error making POST request to", url, ":", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ошибка при чтении ответа:", err)
		return
	}

	var token tokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		fmt.Println("ошибка при разборе JSON:", err)
		return
	}

	userTokens[i] = token

	end := time.Now()

	fmt.Println("ответ от сервера:", string(body), end.Sub(start).Milliseconds())
}

type tokenResponse struct {
	Access  string `json:"accessToken"`
	Refresh string `json:"refreshToken"`
}

func generateRandomString(length int, charset string) string {
	rand.Seed(time.Now().UTC().UnixNano())

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func generateRandomName() string {
	names := []string{
		"Ivan", "Maria", "John", "Sophia", "Michael", "Olivia", "Alexander", "Emma",
		"Daniel", "Isabella", "David", "Ava", "William", "Mia", "James", "Charlotte",
		"Joseph", "Ella", "Andrew", "Amelia", "Benjamin", "Grace", "Nicholas", "Chloe",
		"Matthew", "Lily", "Gabriel", "Sophie", "Oliver", "Emily", "Samuel", "Anna",
		"Robert", "Julia", "Adam", "Natalie", "Peter", "Victoria",
	}
	return names[rand.Intn(len(names))]
}

func generateRandomPatronymic() string {
	patronymics := []string{
		"Ivanovich", "Petrovich", "Sidorovich", "Andreevna", "Markovich", "Vladimirovna",
		"Nikolaevich", "Dmitrievna", "Alexandrovich", "Ilyich", "Fedorovna", "Sergeevich",
		"Viktorovich", "Olegovna", "Vasilievich", "Yurievna", "Borisovich", "Aleksandrovna",
		"Romanovich", "Andreevna", "Mikhailovich", "Nikolaevna", "Denisovich", "Anatolyevna",
		"Daniilovich", "Ivanovna", "Pavelovich", "Nikolaevna", "Maximovich", "Viktorovna",
		"Artemovich", "Antonovna", "Glebovich", "Aleksandrovna", "Kirillovich", "Ilyinichna",
		"Iosifovich", "Dmitrievna", "Yaroslavovich", "Stanislavovna",
	}
	return patronymics[rand.Intn(len(patronymics))]
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	url := "http://localhost:3000/api/signup"

	N := 250

	users := make([]models.User, N)

	var wg sync.WaitGroup
	rand.Seed(time.Now().Unix())
	for i := 0; i < N; i++ {
		// if i%N == 0 {
		// 	time.Sleep(time.Second)
		// }

		wg.Add(1)
		data := map[string]string{
			"name":       generateRandomName(),
			"lastname":   generateRandomName() + "ov",
			"patronymic": generateRandomPatronymic(),
			"email":      fmt.Sprintf("ivan%v@mail.ru", i),
			"password":   "ivanivan",
		}
		body, _ := json.Marshal(data)

		go postRequest(i, url, body, users, "ivanivan", &wg)
		// time.Sleep(time.Millisecond * 200)
	}

	wg.Wait()
	fmt.Println("All requests completed")
	time.Sleep(time.Second)

	url = "http://localhost:3000/api/login"

	userTokens := make([]tokenResponse, N)

	for i := 0; i < len(users); i++ {
		wg.Add(1)
		data := map[string]string{
			// "name":       users[i].Name,
			// "lastname":   users[i].Lastname,
			// "patronymic": users[i].Patronymic,
			"email":    users[i].Email,
			"password": users[i].Password,
		}
		body, _ := json.Marshal(data)

		go postLogin(i, url, body, userTokens, &wg)
	}

	wg.Wait()

	// url = "http://localhost:3000/api/user/ws"
	// // u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/api/user/ws"}
	// clients := make(map[*websocket.Conn]bool)

	// for i := 0; i < len(userTokens); i++ {
	// 	conn, err := establishWebSocketConnection(url, userTokens[i].Access)
	// 	if err != nil {
	// 		log.Println(err)
	// 	} else {
	// 		clients[conn] = true
	// 	}
	// }

	// chatId := 1
	// for k := range clients {
	// 	chatId++
	// 	wg.Add(2)

	// 	go func() {
	// 		defer wg.Done()

	// 		message := map[string]interface{}{
	// 			"action": "join-chat",
	// 			"message": map[string]string{
	// 				"text": "chat1",
	// 			},
	// 			"target": fmt.Sprintf("chat%v", chatId/10),
	// 		}

	// 		jsonBytes, err := json.Marshal(message)
	// 		if err != nil {
	// 			fmt.Println("Ошибка при кодировании в JSON:", err)
	// 			return
	// 		}

	// 		err = k.WriteMessage(websocket.TextMessage, jsonBytes)
	// 		if err != nil {
	// 			log.Println("Error writing message:", err)
	// 			return
	// 		}

	// 		for i := 1; i <= 1; i++ {
	// 			message := map[string]interface{}{
	// 				"action": "send-message",
	// 				"message": map[string]string{
	// 					"text": fmt.Sprintf("helloooooooooooooooooo chat%v !!!", chatId/10),
	// 				},
	// 				"target": fmt.Sprintf("chat%v", chatId/10),
	// 			}
	// 			jsonBytes, err := json.Marshal(message)
	// 			if err != nil {
	// 				fmt.Println("Ошибка при кодировании в JSON:", err)
	// 				return
	// 			}
	// 			startTime := time.Now()
	// 			err = k.WriteMessage(websocket.TextMessage, jsonBytes)
	// 			if err != nil {
	// 				log.Println("Error writing message:", err)
	// 				break
	// 			}
	// 			endTime := time.Now()
	// 			sendDuration := endTime.Sub(startTime)
	// 			log.Printf("Sent message: %s, Time: %v\n", message, sendDuration)
	// 		}
	// 	}()

	// 	// Горутина для чтения сообщений
	// 	go func() {
	// 		defer wg.Done()
	// 		for {
	// 			_, msg, err := k.ReadMessage()
	// 			if err != nil {
	// 				log.Println("Error reading message:", err)
	// 				break
	// 			}
	// 			receiveTime := time.Now()
	// 			log.Printf("Received message: %s, Time: %v\n", string(msg), receiveTime)
	// 		}
	// 	}()
	// }

}

func generateWebSocketKey() string {
	key := make([]byte, 16)
	_, err := cry.Read(key)
	if err != nil {
		return ""
	}

	keyBase64 := base64.StdEncoding.EncodeToString(key)
	return keyBase64
}

func establishWebSocketConnection(url, accessToken string) (*websocket.Conn, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Add("Cookie", fmt.Sprintf("access_token=%s", accessToken))
	req.Header.Add("Connection", "Upgrade")
	req.Header.Add("Upgrade", "websocket")
	req.Header.Add("Sec-WebSocket-Version", "13")
	req.Header.Add("Sec-WebSocket-Key", generateWebSocketKey())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("failed to upgrade to WebSocket: %v", resp.Status)
	}

	url = "ws://localhost:3000/api/user/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, req.Header)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
