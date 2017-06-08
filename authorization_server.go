package gautu

import (
	"fmt"

	"git.g7n3.com/hamster/gautu/model"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"strings"

	"github.com/garyburd/redigo/redisx"
	"gopkg.in/configo.v2"
)

type GautuServer struct {
	store  *sessions.RedisStore
	router *gin.Engine
	db     *gorm.DB
	rds    *redisx.ConnMux
}

var (
	//oaserver *OAuthServer
	gs *GautuServer
)

func init() {
	gs = new(GautuServer)

	gs.store = NewRedisStore()

	gs.router = gin.Default()

	gs.router.Use(sessions.Sessions("gautu", *gs.store))
	gs.router.Static("/static", "static")
	gs.db = model.Gorm()
	gs.rds = NewRedis()
}

type defaultRedis struct {
	Address  string
	Port     string
	Password string
	DB       string
	User     string
}

func NewRedis() *redisx.ConnMux {

	dr := defaultRedis{
		Address:  "localhost",
		Port:     "6379",
		Password: "",
		DB:       "1",
		User:     "x",
	}

	if pro, err := configo.Get("redis"); err == nil {
		fmt.Println(pro)
		dr.Address = pro.MustGet("addr", "localhost")
		dr.Port = pro.MustGet("port", "6379")
		dr.Password = pro.MustGet("password", "")
		dr.DB = pro.MustGet("db", "1")
		dr.User = pro.MustGet("user", "x")
	}

	//addr := strings.Join([]string{dr.Address, dr.Port}, ":")

	//addr := fmt.Sprintf("redis://%s:%s@%s:%s/%s", dr.User, dr.Password, dr.Address, dr.Port, dr.DB)
	//op := redis.DialPassword(dr.Password)
	//redis.DialNetDial()
	//c, err := redis.Dial("tcp", ":6379", op)
	//log.Panicln(addr)
	addr := strings.Join([]string{dr.Address, dr.Port}, ":")
	//c, err := redis.DialURL(addr)
	//if err != nil {
	//	panic(err)
	//}
	c, err := redis.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	if dr.Password != "" {
		if _, err := c.Do("AUTH", dr.Password); err != nil {
			c.Close()
			panic(err)
		}
	}

	_, err = c.Do("SELECT", dr.DB)
	if err != nil {
		c.Close()
		panic(err)
	}
	cmux := redisx.NewConnMux(c)

	return cmux
}

func GetRedis() redis.Conn {
	return gs.rds.Get()
}

func DefaultDB() *gorm.DB {
	return gs.db
}

func (gs *GautuServer) GetDB() *gorm.DB {
	return gs.db
}

func DefaultRedisStrore() *sessions.RedisStore {
	return gs.store
}
func (gs *GautuServer) GetRedisStrore() *sessions.RedisStore {
	return gs.store
}

func DefaultRouter() *gin.Engine {
	return gs.router
}
func (gs *GautuServer) GetRouter() *gin.Engine {
	return gs.router
}

func DefaultGautuServer() *GautuServer {
	return gs
}

func NewRedisStore() *sessions.RedisStore {
	dr := defaultRedis{
		Address:  "localhost",
		Port:     "6379",
		Password: "",
		DB:       "1",
		User:     "x",
	}

	if pro, err := configo.Get("redis"); err == nil {
		fmt.Println(pro)
		dr.Address = pro.MustGet("addr", "localhost")
		dr.Port = pro.MustGet("port", "6379")
		dr.Password = pro.MustGet("password", "")
		dr.DB = pro.MustGet("db", "1")
		dr.User = pro.MustGet("user", "x")
	}

	addr := strings.Join([]string{dr.Address, dr.Port}, ":")
	store, _ := sessions.NewRedisStoreWithDB(10, "tcp", addr, dr.Password, dr.DB, []byte("secret"))
	return &store
}

func (gs *GautuServer) Run(addr string) {
	gs.router.Run(addr)
}

func Start() {
	gs.Router()

	gs.Run(serverAddr())
}

func serverAddr() (r string) {
	server, err := configo.Get("server")
	if err != nil {
		r = ":7777"
		return
	}
	r = strings.Join([]string{
		server.MustGet("addr", ""),
		server.MustGet("port", "7777"),
	},
		":")
	return
}
