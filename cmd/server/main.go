package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"
	psql "github.com/Tanya1515/metrics-collector.git/cmd/storage/postgresql"
	str "github.com/Tanya1515/metrics-collector.git/cmd/storage/structure"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

type Application struct {
	Storage storage.RepositoryInterface
	Logger  zap.SugaredLogger
}

var (
	serverAddressFlag *string
	storeIntervalFlag *int
	fileStorePathFlag *string
	restoreFlag       *bool
	postgreSQLFlag    *string
)

// при синхронной записи сбрасывается значение PollCount

func init() {
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
	postgreSQLFlag = flag.String("d", "", "credentials for database")
	storeIntervalFlag = flag.Int("i", 300, "time duration for saving metrics")
	fileStorePathFlag = flag.String("f", "/tmp/metrics-db.json", "filename for storing metrics")
	restoreFlag = flag.Bool("r", true, "store all info")
}

func main() {
	var Storage storage.RepositoryInterface

	flag.Parse()

	serverAddress, envExists := os.LookupEnv("ADDRESS")
	if !(envExists) {
		serverAddress = *serverAddressFlag
	}

	postgreSQLAddress, envExists := os.LookupEnv("DATABASE_DSN")
	if !(envExists) {
		postgreSQLAddress = *postgreSQLFlag
	}

	if postgreSQLAddress != "" {
		postgreSQLAddrPortDatabase := strings.Split((strings.Split((strings.Split(postgreSQLAddress, "@"))[1], "?"))[0], ":")
		postgreSQLDatabase := "postgres"
		postgreSQLPort := "5432"
		fmt.Println(postgreSQLAddrPortDatabase)
		if len(postgreSQLAddrPortDatabase) == 2 {
			fmt.Println(postgreSQLAddrPortDatabase[0])
			fmt.Println(postgreSQLAddrPortDatabase[1])
			postgreSQLPortDatabase := strings.Split(postgreSQLAddrPortDatabase[1], "/")
			if len(postgreSQLPortDatabase) == 2 {
				fmt.Println(postgreSQLPortDatabase[0])
				fmt.Println(postgreSQLPortDatabase[1])
				postgreSQLDatabase = postgreSQLPortDatabase[1]
			}
			postgreSQLPort = postgreSQLPortDatabase[0]
		}
		postgreSQLAddr := postgreSQLAddrPortDatabase[0]
		Storage = &psql.PostgreSQLConnection{Address: postgreSQLAddr, Port: postgreSQLPort, UserName: "postgres", Password: "postgres", DBName: postgreSQLDatabase}
	} else {
		Storage = &str.MemStorage{}
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	App := Application{Storage: Storage, Logger: *logger.Sugar()}

	App.Logger.Infow(
		"Starting server",
		"addr", serverAddress,
	)

	storeInterval := 300
	restore := true
	storeIntervalEnv, envExists := os.LookupEnv("STORE_INTERVAL")
	if !(envExists) {
		storeInterval = *storeIntervalFlag
	} else {
		storeInterval, err = strconv.Atoi(storeIntervalEnv)
		if err != nil {
			App.Logger.Errorln("Error when converting string to int: ", err)
		}
	}

	fileStore, envExists := os.LookupEnv("FILE_STORAGE_PATH")
	if !(envExists) {
		fileStore = *fileStorePathFlag
	}

	restoreEnv, envExists := os.LookupEnv("RESTORE")
	if !(envExists) {
		restore = *restoreFlag
	} else {
		restore, err = strconv.ParseBool(restoreEnv)
		if err != nil {
			App.Logger.Errorln("Error when converting string to bool: ", err)
		}
	}

	err = Storage.Init(restore, fileStore, storeInterval)
	if err != nil {
		fmt.Println(err)
	}

	commonMiddlewares := []data.Middleware{
		App.MiddlewareZipper,
		App.MiddlewareLogger,
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", App.MiddlewareChain(App.HTMLMetrics(), commonMiddlewares...))
		r.Get("/value/{metricType}/{metricName}", App.MiddlewareChain(App.GetMetricPath(), commonMiddlewares...))
		r.Post("/update/{metricType}/{metricName}/{metricValue}", App.MiddlewareChain(App.UpdateValuePath(), commonMiddlewares...))
		r.Post("/value/", App.MiddlewareChain(App.GetMetric(), commonMiddlewares...))
		r.Post("/update/", App.MiddlewareChain(App.UpdateValue(), commonMiddlewares...))
		r.Post("/updates/", App.MiddlewareChain(App.UpdateAllValues(), commonMiddlewares...))
		r.Get("/ping", App.MiddlewareChain(App.CheckStorageConnection(), commonMiddlewares...))
	})

	err = http.ListenAndServe(serverAddress, r)
	if err != nil {
		fmt.Println(err)
		App.Logger.Fatalw(err.Error(), "event", "start server")
	}
}
