package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Name  string `json:"name"`
	Email string `gorm:"unique" json:"email"`
	Age   uint   `json:"age"`
}

var redisClient *redis.Client
var db *gorm.DB
var err error

func initRedis(ctx context.Context) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	fmt.Println("Подключение к Redis установлено")
}

func initPostgres() {
	dsn := "host=127.0.0.1 user=admin password=di1mon11421 dbname=test port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	fmt.Println("Подключение к базе данных установлено")
}

func GetUserFromPgrs(userID uint) (*Users, error) {
	var user Users
	result := db.First(&user, userID)
	if result.Error != nil {
		return nil, fmt.Errorf("ошибка получения пользователя из постгрес: %v", result.Error)
	}
	return &user, nil
}
func GetUserFromRedis(ctx context.Context, userID uint) (*Users, error) {
	var user Users
	key := fmt.Sprintf("user:%d", userID)
	userJSON, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		fmt.Println("Пользователь не найден. Запрос отправлен в Redis")
		return nil, nil
	}
	if err != nil {
		fmt.Println("Пользователь не найден. Запрос отправлен в Postgres")
		return nil, fmt.Errorf("ошибка при получении данных из Redis")
	}
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}
	fmt.Printf("Пользователь %d получен из Redis\n", userID)
	return &user, nil
}

func SetUserToRedis(ctx context.Context, user *Users) error {
	key := fmt.Sprintf("user:%d", user.ID)
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}
	err = redisClient.Set(ctx, key, userJSON, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("ошибка сохранения в Redis: %v", err)
	}

	log.Printf("Пользователь %s добавлен в редис", user.Name)
	return nil
}

func GetUSerFromCache(ctx context.Context, userID uint) (*Users, error) {
	var user *Users
	user, err = GetUserFromRedis(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка: %f", err)
	}
	if user != nil {
		return user, nil
	}
	user, err = GetUserFromPgrs(userID)
	if err != nil {
		return nil, err
	}
	err := SetUserToRedis(ctx, user)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось сохранить в Redis: %v\n", err)
	}
	return user, nil

}

func main() {
	ctx := context.Background()
	initRedis(ctx)
	initPostgres()
	err = db.AutoMigrate(&Users{})
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}
	user1, err := GetUSerFromCache(ctx, 1)
	if err != nil {
		log.Fatalf("Ошибка получения пользователя 1: %v", err)
	}
	fmt.Println(user1)
	user1, err = GetUSerFromCache(ctx, 1)
	if err != nil {
		log.Fatalf("Ошибка получения пользователя 1: %v", err)
	}
	fmt.Println(user1)
	user2, err := GetUSerFromCache(ctx, 2)
	if err != nil {
		log.Fatalf("Ошибка получения пользователя 1: %v", err)
	}
	fmt.Println(user2)
	user5, err := GetUSerFromCache(ctx, 5)
	if err != nil {
		log.Fatalf("Ошибка получения пользователя 1: %v", err)
	}
	fmt.Println(user5)
}
