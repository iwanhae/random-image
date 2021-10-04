package config

import (
	"log"
	"os"
	"reflect"

	"github.com/spf13/viper"
)

type RandomImageConfig struct {
	S3Endpoint        string  `mapstructure:"s3_endpoint"`
	S3AccessKey       string  `mapstructure:"s3_access_key"`
	S3SecrretKey      string  `mapstructure:"s3_secret_key"`
	S3BucketName      string  `mapstructure:"s3_bucket_name"`
	S3CacheBucketName *string `mapstructure:"s3_cache_bucket_name"`
	S3EnableTLS       bool    `mapstructure:"s3_enable_tls"`

	MongoURI      string `mapstructure:"mongo_uri"`
	MongoDatabase string `mapstructure:"mongo_database"`
}

const (
	EnvPrefix = "RI"
)

func Init() RandomImageConfig {
	var res RandomImageConfig
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		log.Println("Load from ENV")
		v.AutomaticEnv()
		rt := reflect.TypeOf(res)
		for i := 0; i < rt.NumField(); i++ {
			t := rt.Field(i)
			tag, ok := t.Tag.Lookup("mapstructure")
			if !ok {
				continue
			}
			v.BindEnv(tag)
		}
	} else {
		log.Println("Load from dotenv")
		v.SetConfigType("dotenv")
		v.SetConfigFile(".env")
		err := v.ReadInConfig()
		if err != nil {
			log.Fatalf("fail to read dotenv file:%s", err.Error())
		}
	}

	err := v.Unmarshal(&res)
	if err != nil {
		panic(err)
	}
	return res
}
