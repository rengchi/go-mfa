package main

import (
	"fmt"
	"log"
	"strings"
	"time"

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

func main() {
	// 名称、密钥
	var secret string

	fmt.Print("请输入 TOTP 密钥: ")
	fmt.Scanln(&secret)

	// 第一步：生成当前时间的动态验证码
	currentCode, err := GenerateDynamicCode(secret)
	if err != nil {
		log.Fatalf("生成动态验证码时出错: %v", err)
	}
	fmt.Printf("当前时间的动态验证码为: %s\n", currentCode)

	// // 第二步：验证输入的动态验证码
	// fmt.Print("请输入要验证的 TOTP 动态验证码: ")
	// var code string
	// fmt.Scanln(&code)

	// if ValidateCode(secret, code) {
	// 	fmt.Println("✅ 动态验证码有效！")
	// } else {
	// 	fmt.Println("❌ 动态验证码无效！")
	// }
}
