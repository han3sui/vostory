// cmd/createadmin/main.go
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/config"
	"iot-alert-center/pkg/log"
	"os"
	"time"

	"github.com/duke-git/lancet/v2/cryptor"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	envConf       = flag.String("conf", "config/dev.yml", "config path, eg: -conf ./config/dev.yml")
	adminUsername = flag.String("user", "admin", "admin username, eg: -user superadmin")
)

func main() {
	flag.Parse()

	// 初始化配置
	conf := config.NewConfig(*envConf)
	logger := log.NewLog(conf)

	// 初始化数据库连接
	db := repository.NewDB(conf, logger)

	// 创建管理员账号
	if err := createSuperAdmin(db, logger, *adminUsername); err != nil {
		logger.Error("Failed to create admin user", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Admin user created successfully")
}

// 生成随机密码
func generateRandomPassword(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes), nil
}

func createSuperAdmin(db *gorm.DB, logger *log.Logger, username string) error {
	// 检查用户是否存在
	var count int64
	db.Model(&model.SysUser{}).Where("login_name = ?", username).Count(&count)
	// 生成随机密码
	password, err := generateRandomPassword(12)
	//对password进行SHA256
	sha256Password := cryptor.Sha256(password)

	if count > 0 {
		logger.Info("Admin user already exists", zap.String("username", username))

		// 询问是否重置密码
		fmt.Print("Admin user already exists. Reset password? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			return nil
		}

		if err != nil {
			return err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(sha256Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		now := time.Now()
		if err := db.Model(&model.SysUser{}).Where("login_name = ?", username).
			Updates(map[string]interface{}{
				"password":        string(hashedPassword),
				"pwd_update_date": &now,
			}).Error; err != nil {
			return err
		}

		printCredentials(username, password)
		return nil
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(sha256Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	deptID := uint(0)

	// 创建管理员用户
	now := time.Now()
	admin := model.SysUser{
		LoginName:     username,
		UserName:      "超级管理员",
		Password:      string(hashedPassword),
		Status:        "0",  // 启用状态
		UserType:      "99", // 超级管理员
		PwdUpdateDate: &now,
		Sex:           "0", // 默认性别
		UserDeptID:    &deptID,
	}

	// 保存到数据库
	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	printCredentials(username, password)
	return nil
}

func printCredentials(username, password string) {
	fmt.Printf("\n========== 超级管理员账号 ==========\n")
	fmt.Printf("用户名: %s\n", username)
	fmt.Printf("密码: %s\n", password)
	fmt.Printf("======================================\n\n")
	fmt.Println("请记录此密码，它不会再次显示。")
}
