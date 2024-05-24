package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	users := make([]map[string]string, 0)

	for i := 0; i < 1000; i++ {
		psw := fmt.Sprintf("ivan%v@mail.ru", i)
		data := map[string]string{
			"name":       "Ivan",
			"lastname":   "Ivanov",
			"patronymic": "Ivanovich",
			"email":      psw,
			"password":   "ivanivan",
		}
		users = append(users, data)
	}

	usersJSON, err := json.Marshal(users)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	file, err := os.Create("users1000.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer file.Close()

	_, err = file.Write(usersJSON)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Users data saved to users.json.")
}
