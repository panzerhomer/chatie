package models

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const emailRgxString = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

type ChatUser struct {
	// BaseModel
	ID         int    `json:"id"`
	Role       string `json:"role"`
	Name       string `json:"name"`
	Lastname   string `json:"lastname"`
	Patronymic string `json:"patronymic"`
	Tag        string `json:"tag"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	IsOnline   bool   `json:"isOnline"`
}

type User struct {
	ID         int    `json:"id"`
	Role       string `json:"role"`
	Name       string `json:"name"`
	Lastname   string `json:"lastname"`
	Patronymic string `json:"patronymic"`
	Tag        string `json:"tag"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"-"`
	// IsOnline   bool      `json:"isOnline"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (u *User) Sanitaze() {
	u.Password = ""
}

type UserRegister struct {
	Name       string `json:"name"`
	Lastname   string `json:"lastname"`
	Patronymic string `json:"patronymic"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *User) GetID() int {
	return u.ID
}

func (u *User) GetName() string {
	return u.Username
}

func (u *UserRegister) Validate() error {
	if len(u.Password) < 4 || len(u.Password) > 16 {
		return fmt.Errorf("password must be 4-16 symbols")
	}

	emailRegex := regexp.MustCompile(emailRgxString)
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("email is wrong")
	}

	if u.Name == "" {
		return fmt.Errorf("name is empty")
	}
	if u.Lastname == "" {
		return fmt.Errorf("lastname is empty")
	}
	if u.Patronymic == "" {
		return fmt.Errorf("patronymic is empty")
	}

	return nil
}

func (u *User) GenerateUsername() {
	replaceMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e",
		'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k",
		'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r",
		'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts",
		'ч': "ch", 'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
		'э': "e", 'ю': "yu", 'я': "ya",
	}

	russianText := strings.ToLower(u.Name) + strings.ToLower(u.Lastname)

	latinText := strings.Map(func(r rune) rune {
		if replacement, ok := replaceMap[r]; ok {
			return []rune(replacement)[0]
		}
		return r
	}, russianText)

	// r := rand.New(rand.NewSource(99))
	latinText += fmt.Sprintf("%v%v", time.Now().Year()/100, rand.Intn(9000)+1000)

	u.Username = latinText
}
