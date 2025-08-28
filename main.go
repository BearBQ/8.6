package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Name  string `json: "name"`
	Email string `gorm:"unique" json:email`
	Age   uint   `json: "email"`
}

func main() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	dsn := "host=127.0.0.1 user=admin password=di1mon11421 dbname=test port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	_ = db
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}

	db.AutoMigrate(&Users{})
	// db.Create(&Users{
	// 	Name:  "Admin",
	// 	Email: "admin@test.com",
	// 	Age:   333,
	// })
	// db.Create(&Users{
	// 	Name:  "Egor",
	// 	Email: "egor@test.com",
	// 	Age:   333,
	// })
	// db.Create(&Users{
	// 	Name:  "Vasya",
	// 	Email: "vasya@test.com",
	// 	Age:   333,
	// })

}
