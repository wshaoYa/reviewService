package data

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"reviewService/internal/conf"
	"reviewService/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewReviewRepo, NewESClient, NewRedisClient)

// Data .
type Data struct {
	// TODO wrapped database client
	q   *query.Query
	log *log.Helper
	es  *elasticsearch.TypedClient
	rdb *redis.Client
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, es *elasticsearch.TypedClient, rdb *redis.Client, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}

	query.SetDefault(db)
	return &Data{
		q:   query.Q,
		log: log.NewHelper(logger),
		es:  es,
		rdb: rdb,
	}, cleanup, nil
}

// NewDB 新建DB
func NewDB(c *conf.Data) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(c.Database.GetSource()), &gorm.Config{})
}

// NewESClient esClient构造函数
func NewESClient(c *conf.ES) (*elasticsearch.TypedClient, error) {
	cfg := elasticsearch.Config{
		Addresses: c.Addresses,
	}
	return elasticsearch.NewTypedClient(cfg)
}

// NewRedisClient redis client 构造函数
func NewRedisClient(c *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})
}
