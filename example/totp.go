package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/cobra"
)

var issuer, accountName, path string

func generate(issuer, accountName string) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,        // 应用名
		AccountName: accountName,   // 用户账号
		Period:      30,            // 30秒
		Digits:      otp.DigitsSix, // 默认六位
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(key.URL())
	fmt.Println("Secret:", key.Secret()) // 存入数据库
	now := time.Now()
	code, err := totp.GenerateCode(key.Secret(), now)

	if err != nil {
		log.Fatalf("Failed to generate TOTP code: %v", err)
	}
	fmt.Printf("Generated TOTP Code: %s\n", code)

	isValid := totp.Validate(code, key.Secret())
	if isValid {
		fmt.Println("TOTP Code is valid!")
	} else {
		fmt.Println("Invalid TOTP Code.")
	}
}

func main() {
	// 创建根命令
	var rootCmd = &cobra.Command{
		Use:   "生成动态口令二维码",
		Short: "TOTP动态码",
		Run: func(cmd *cobra.Command, args []string) {
			// 参数校验
			if issuer == "" && accountName == "" {
				_ = cmd.Usage()
				os.Exit(1)
			}
			generate(issuer, accountName)
		},
	}

	// 添加参数：全称 --name，缩写 -n
	rootCmd.Flags().StringVarP(&issuer, "issuer", "i", "", "请输入你的应用名称（必填）")
	rootCmd.Flags().StringVarP(&accountName, "account_name", "a", "", "请输入你的账号（必填）")
	//rootCmd.Flags().StringVarP(&path, "path", "p", "", "请输入二维码保存地址（必填）")

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("执行错误:", err)
		os.Exit(1)
	}

}
