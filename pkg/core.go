package pkg

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

// MinioHelper 封装了连接 Minio 和 PostgreSQL 的功能
type MinioHelper struct {
	minioClient *minio.Client
	db          *sql.DB
}

type MinioConfig struct {
	endpoint  string
	accessKey string
	secretKey string
}

type PostgreConfig struct {
	dbName     string
	dbUser     string
	dbPassword string
	dbHost     string
	dbPort     string
}

var MinioHelperIns *MinioHelper

// NewMinioHelper 创建一个 MinioHelper 实例
func NewMinioHelper(minioConfig *MinioConfig, postgreConfig *PostgreConfig) (*MinioHelper, error) {
	// 初始化 Minio 客户端
	minioClient, err := minio.New(minioConfig.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioConfig.accessKey, minioConfig.secretKey, ""),
		Secure: false, // Set this value to true for secure (HTTPS) access.
	})
	if err != nil {
		return nil, err
	}

	// 初始化 PostgreSQL 数据库连接
	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			postgreConfig.dbHost,
			postgreConfig.dbPort,
			postgreConfig.dbUser,
			postgreConfig.dbPassword,
			postgreConfig.dbName))
	if err != nil {
		return nil, err
	}

	// 确保数据库连接正常
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &MinioHelper{
		minioClient: minioClient,
		db:          db,
	}, nil
}

// Close 关闭 MinioHelper 实例中的资源（Minio 客户端和数据库连接）
func (m *MinioHelper) Close() {

	if m.db != nil {
		m.db.Close()
	}
}

func init() {
	// 设置 viper 配置文件的路径和文件名
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	// 将配置文件映射到 MinioConfig 结构
	minioConfig := &MinioConfig{
		endpoint:  viper.GetString("minio.endpoint"),
		accessKey: viper.GetString("minio.access_key"),
		secretKey: viper.GetString("minio.secret_key"),
	}

	// 将配置文件映射到 PostgresConfig 结构
	postgreConfig := &PostgreConfig{
		dbName:     viper.GetString("postgre.db_name"),
		dbUser:     viper.GetString("postgre.db_user"),
		dbPassword: viper.GetString("postgre.db_password"),
		dbHost:     viper.GetString("postgre.db_host"),
		dbPort:     viper.GetString("postgre.db_port"),
	}

	MinioHelperIns, err = NewMinioHelper(minioConfig, postgreConfig)
	if err != nil {
		log.Fatalf("init minio helper error %s", err)
	}

}
