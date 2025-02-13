package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pquerna/otp/totp"
	"github.com/rengchi/ji"
)

// Config config.json配置
type Config struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Passwd   string `json:"passwd"`
	Database string `json:"database"`
	Prefix   string `json:"prefix"`
}

// ValidateCode 验证给定的 TOTP 动态验证码是否有效。
func ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// padBase32Secret 为密钥添加 Base32 的填充字符。
func padBase32Secret(secret string) string {
	missingPadding := len(secret) % 8
	if missingPadding != 0 {
		secret += strings.Repeat("=", 8-missingPadding)
	}
	return secret
}

// GenerateDynamicCode 使用提供的密钥生成当前时间的动态验证码。
func GenerateDynamicCode(secret string) (string, error) {
	// 为密钥添加 Base32 填充字符
	secret = padBase32Secret(secret)

	// 尝试生成动态验证码
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		// 检查是否是 Base32 解码失败的错误
		if strings.Contains(err.Error(), "Decoding of secret as base32 failed") {
			return "", fmt.Errorf("无法解码密钥，密钥格式可能不正确或已损坏，请检查密钥的正确性: %w", err)
		}
		// 其他错误处理
		return "", fmt.Errorf("无法生成动态验证码: %w", err)
	}
	return code, nil
}

// saveToDatabase 保存名称和密钥到数据库，如果已存在则返回错误而不保存。
func saveToDatabase(db *sql.DB, name, secret string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务时出错: %w", err)
	}

	// 检查名称是否已存在
	var existingSecret string
	err = tx.QueryRow("SELECT secret FROM mfa_secret WHERE name = ?", name).Scan(&existingSecret)

	switch err {
	case nil:
		// 如果已存在，直接返回
		tx.Rollback() // 没有更改数据库，回滚事务
		return nil
	case sql.ErrNoRows:
		// 如果不存在，插入新记录
		_, err = tx.Exec("INSERT INTO mfa_secret (name, secret) VALUES (?, ?)", name, secret)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("插入记录时出错: %w", err)
		}
		// 提交事务
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("提交事务时出错: %w", err)
		}
		return nil
	default:
		tx.Rollback()
		return fmt.Errorf("查询数据库时出错: %w", err)
	}
}

// createTable 如果 mfa_secret 表不存在，则创建表
func createTableIfNotExists(db *sql.DB) error {
	// 创建表语句
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS mfa_secret (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		secret VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name)  -- 确保名称唯一
	);`

	// 执行创建表的查询
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("创建表时出错: %w", err)
	}
	return nil
}

func main() {
	// 读取配置文件
	var conf = &Config{}
	err := ji.ReadJSONFromFile("./config.json", conf)
	if err != nil {
		log.Fatalf("请检查：%v文件是否存在，格式参考：%v\n", "./config.json", "./example.config.json")
	}
	// MySQL 连接
	dsn := ji.Assemble(conf.User, ":", conf.Passwd, "@tcp(", conf.Host, ":", conf.Port, ")/", conf.Database) // 替换为实际的 MySQL 用户名、密码和数据库名
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		// 数据库连接配置错误，输出具体的错误信息
		log.Fatalf("无法连接到数据库配置: %v\n\n请检查以下几项:\n1. 确保数据库连接字符串格式正确，包括用户名、密码、数据库名、主机地址和端口\n2. 确保使用的数据库驱动程序和版本与你的 MySQL 数据库兼容\n3. 检查数据库连接字符串中的特殊字符是否需要转义", err)
	}
	defer db.Close()

	// 配置数据库连接池
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// 检查数据库连接是否成功
	if err = db.Ping(); err != nil {
		// 数据库连接失败，输出更加详细的错误信息
		log.Fatalf("无法连接到 MySQL 数据库: %v\n\n请检查以下几项:\n1. 确保 MySQL 数据库已启动并在 %v:%v 端口监听\n2. 确保防火墙没有阻止该端口\n3. 检查数据库用户名、密码是否正确\n4. 如果数据库在远程服务器上，确保网络连接正常", err, conf.Host, conf.Port)
	}
	fmt.Println("成功连接到数据库！")

	// 确保表存在
	err = createTableIfNotExists(db)
	if err != nil {
		log.Fatalf("创建数据库表时出错: %v", err)
	}

	// 用户输入名称和密钥
	var name, secret string
	var code, existingSecret string // 声明code和existingSecret变量

	// 输入名称
	fmt.Print("请输入 TOTP 名称: ")
	fmt.Scanln(&name)

	// 查询数据库，检查名称是否存在
	err = db.QueryRow("SELECT secret FROM mfa_secret WHERE name = ?", name).Scan(&existingSecret)
	if err == nil {
		// 如果名称已存在，直接生成动态验证码。并测试生成验证码
		code, err = GenerateDynamicCode(existingSecret) // 使用 `err =` 来更新错误变量
		if err != nil {
			log.Fatalf("生成动态验证码时出错！%v", err)
		}

		fmt.Printf("当前时间的动态验证码为: %s\n", code)
		return
	} else if err != sql.ErrNoRows {
		log.Fatalf("查询数据库时出错: %v", err)
	}

	// 如果名称不存在，提示输入密钥
	fmt.Print("请输入 TOTP 密钥: ")
	fmt.Scanln(&secret)

	// 先生成验证码以确保密钥有效
	code, err = GenerateDynamicCode(secret) // 使用 `err =` 来更新错误变量
	if err != nil {
		log.Fatalf("生成动态验证码时出错！%v", err)
	}

	// 保存或更新密钥到数据库
	err = saveToDatabase(db, name, secret)
	if err != nil {
		log.Fatalf("保存数据时出错: %v", err)
	}

	// 输出生成的动态验证码
	fmt.Printf("当前时间的动态验证码为: %s\n", code)
}
