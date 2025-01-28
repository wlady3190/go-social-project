package main

import (
	// "time"

	"expvar"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wlady3190/go-social/internal/auth"
	"github.com/wlady3190/go-social/internal/db"
	"github.com/wlady3190/go-social/internal/env"
	"github.com/wlady3190/go-social/internal/ratelimiter"
	"github.com/wlady3190/go-social/internal/store/cache"

	// "github.com/wlady3190/go-social/internal/mailer"
	"github.com/wlady3190/go-social/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.2"

//* COnfiguración para SWAGGER
//	@title			Social API
//	@version		1.0
//	@description	This a new test in go
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1
//
//* Parámetros adicionales
//@securityDefinitions.apikey ApiKeyAuth
//@in			header
//@name			Authorization
//@description
//! De aquí a API

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		//* Para el swagger
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		//! Para la confirmación
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		db: dbConfig{
			addr:              env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns:      env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConnetions: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:       env.GetString("DB_MAX_IDLE_TIME", "15m"),
		}, //! A internal para crear db

		//! Redis
		redisCfg: redisConfig{
			add: env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:  env.GetString("REDIS_PW", ""),
			db:  env.GetInt("REDIS_DB", 0),
			// enabled: false,
			//! se extiende las variables de entorno en env.go par agregar GetBool()
			enabled: env.GetBool("REDIS_ENABLED", false),
		},

		env: env.GetString("ENV", "development"),
		//* Para la expiración del token de UserInvite
		// mail: mailConfig{
		// 	exp:       time.Hour * 24 * 3, //3 dias
		// 	fromEmail: env.GetString("SENDGRID_FROM_EMAIL", ""),
		// 	sendgrid: sendGridConfig{
		// 		apikey: env.GetString("SENDGRID_API_KEY", ""),
		// 	},
		// 	//! Para configurar Mailtrap y más abajo tn se complementar
		// 	// mailTrap: mailTrapConfig{
		// 	// 	apikey: env.GetString("MAILTRAP_API_KEY", ""),
		// 	// },
		// },
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},

			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3,
				iss:    "wladysocial",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUEST_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	//* Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync() //! va a main -> aplication
	//* database
	db, err := db.New(cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConnetions,
		cfg.db.maxIdleTime)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("db connected")

	//! Redis -> Se añade un bool para activar redis en api.go
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.add, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis cache connection established")
		defer rdb.Close()

	}
//* AQuí se puede cambiar el metodo de rate limiter en lugar de NewFixed, como el Token, etc.
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)


	store := store.NewPostgresStorage(db) //! Y se pasa a la API
	cacheStorage := cache.NewRedisStorage(rdb)
	//* Viene de mailer SendGrid
	// mailer := mailer.NewSendgrid(cfg.mail.sendgrid.apikey, cfg.mail.fromEmail)

	//! MailTrap config
	// mailTrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apikey, cfg.mail.fromEmail)

	// if err != nil {
	// 	logger.Fatal(err)
	// }

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	app := &application{
		config: cfg,
		store:  store,
		//!redis, luego de implementar el stogare , las interfaces
		cacheStorage: cacheStorage,
		logger:       logger,
		//mailer: mailer, //* De aqui a auth -> RegisterUserHandler
		// mailer: mailTrap,
		authenticator: jwtAuthenticator,
		//* rate Limiter. se hace al ultimo y se aplica como middleware en API
		rateLimiter: rateLimiter,
	}
	//* Metricas
	expvar.NewString("version").Set(version)
	//* ver info de la bd
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	// ver numro de rutinas de go
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	

	mux := app.mount()

	logger.Fatal(app.run(mux))

}
