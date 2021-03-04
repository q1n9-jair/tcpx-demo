package dao

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7" //这里使用的是版本5，最新的是6，有改动
	"go.uber.org/zap"
	"im_socket_server/config"
	"im_socket_server/logs"
	"time"
)

var Client *elastic.Client

//初始化
func inits() {
	config := config.GetConfig()
	esHost := config.GetString("es.addr")

	var err error
	Client, err = elastic.NewClient(elastic.SetURL(esHost),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetGzip(true))
	if err != nil {
		logs.Loggers.Error("initEs:", zap.Error(err))
	}
	info, code, errPing := Client.Ping(esHost).Do(context.Background())
	if errPing != nil {
		logs.Loggers.Error("initEs-errPing:", zap.Error(errPing))
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, errElasticsearchVersion := Client.ElasticsearchVersion(esHost)
	if errElasticsearchVersion != nil {
		logs.Loggers.Error("errElasticsearchVersion:", zap.Error(errElasticsearchVersion))
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)
}

