package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pquerna/otp/totp"
)

// ValidateCode 验证给定的 TOTP 动态验证码是否有效。
func ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// padBase32Secret 为密钥添加 Base32 的填充字符。
func padBase32Secret(secret string) string {
	// Base32 编码长度需是 8 的倍数
	missingPadding := len(secret) % 8
	if missingPadding != 0 {
		secret += strings.Repeat("=", 8-missingPadding)
	}
	return secret
}

// GenerateDynamicCode 使用提供的密钥生成当前时间的动态验证码。
func GenerateDynamicCode(secret string) (string, error) {
	// 为密钥添加填充
	secret = padBase32Secret(secret)

	// 生成基于当前时间的动态验证码
	code, err := totp.GenerateCode(secret, time.Now()) // 直接使用密钥字符串
	if err != nil {
		return "", fmt.Errorf("无法生成动态验证码: %w", err)
	}
	return code, nil
}

// saveToDatabase 保存名称和密钥到数据库，如果已存在则提示是否替换，并返回密钥和错误
func saveToDatabase(db *sql.DB, name, secret string) (string, error) {
	// 检查名称是否已存在
	var existingSecret string
	err := db.QueryRow("SELECT secret FROM mfa_secret WHERE name = ?", name).Scan(&existingSecret)

	switch {
	case err == nil:
		// 不替换，返回数据库中的secret
		return existingSecret, nil

		// // 名称已存在，提示是否替换
		// fmt.Printf("名称 '%s' 已存在，当前密钥为：%s\n", name, existingSecret)
		// var choice string
		// fmt.Print("是否替换现有的密钥？(y/n): ")
		// fmt.Scanln(&choice)

		// if choice == "y" || choice == "Y" {
		// 	// 替换密钥
		// 	_, err = db.Exec("UPDATE mfa_secret SET secret = ? WHERE name = ?", secret, name)
		// 	if err != nil {
		// 		return "", fmt.Errorf("更新记录时出错: %w", err)
		// 	}
		// 	fmt.Println("密钥已更新！")
		// 	return secret, nil
		// } else {
		// 	fmt.Println("密钥未更新。")
		// 	return existingSecret, nil
		// }

	case err == sql.ErrNoRows:
		// 名称不存在，插入新记录
		_, err = db.Exec("INSERT INTO mfa_secret (name, secret) VALUES (?, ?)", name, secret)
		if err != nil {
			return "", fmt.Errorf("插入记录时出错: %w", err)
		}
		fmt.Println("新密钥已保存！")
		return secret, nil

	default:
		// 处理其他类型的错误
		return "", fmt.Errorf("查询数据库时出错: %w", err)
	}
}

func main() {
	// MySQL 连接
	dsn := "root:root@tcp(127.0.0.1:3306)/go_mfa" // 替换为实际的 MySQL 用户名、密码和数据库名
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	defer db.Close()

	// 配置数据库连接池
	// 最大连接池生命周期（最大生存时间）
	db.SetConnMaxLifetime(time.Minute * 3) // 可以根据实际需要调整（比如 3 分钟）
	// 最大打开连接数
	db.SetMaxOpenConns(10) // 根据应用的负载来调整最大连接数
	// 最大空闲连接数
	db.SetMaxIdleConns(10) // 根据数据库和应用负载来调整空闲连接数

	// 检查数据库连接是否成功
	if err = db.Ping(); err != nil {
		log.Fatalf("无法连接到 MySQL 数据库: %v", err)
	}
	fmt.Println("成功连接到数据库！")

	// 名称和密钥
	var name, secret string

	// 用户输入名称和密钥
	fmt.Print("请输入 TOTP 名称: ")
	fmt.Scanln(&name)
	fmt.Print("请输入 TOTP 密钥: ")
	fmt.Scanln(&secret)

	// 保存或更新到数据库
	resultSecret, err := saveToDatabase(db, name, secret)
	if err != nil {
		log.Fatalf("保存数据时出错: %v", err)
	}

	// 第一步：生成当前时间的动态验证码
	currentCode, err := GenerateDynamicCode(resultSecret)
	if err != nil {
		log.Fatalf("生成动态验证码时出错: %v", err)
	}
	fmt.Printf("当前时间的动态验证码为: %s\n", currentCode)
}
