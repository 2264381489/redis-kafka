package main

import (
	conf "awesomeProject/config"
	"awesomeProject/router"
	"awesomeProject/svc"
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"log"
)

// 可通过命令行获取 配置文件 地址
var configFile = flag.String("f", "etc/redis-redis.yaml", "the config file")

func main() {

	flag.Parse()

	var c conf.Config
	config.WithOptions(func(options *config.Options) {
		options.DecoderConfig.TagName = "json"
	})
	config.AddDriver(yaml.Driver)
	err := config.LoadFiles(*configFile)
	if err != nil {
		log.Fatalf("config load err :%+v", err)
	}

	err = config.Decode(&c)
	if err != nil {
		log.Fatalf("config decode err :%+v", err)
	}
	ctx := context.Background()
	// 生成的 上下文 将注册到每一个 http handler 和 后台任务里面
	serviceContext := svc.NewServiceContext(ctx, c)

	r := gin.Default()
	router.RegisterHandlers(r, serviceContext)
	r.Run()
}
