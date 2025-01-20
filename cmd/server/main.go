package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"
	psql "github.com/Tanya1515/metrics-collector.git/cmd/storage/postgresql"
	str "github.com/Tanya1515/metrics-collector.git/cmd/storage/structure"
)

type ResultMetrics struct {
	GaugeMetrics   string
	CounterMetrics string
}

type Application struct {
	Storage   storage.RepositoryInterface
	Logger    zap.SugaredLogger
	Backup    bool
	FileStore string
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

var (
	serverAddressFlag *string
	storeIntervalFlag *int
	fileStorePathFlag *string
	restoreFlag       *bool
	postgreSQLFlag    *string
)

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
		Storage = &psql.PostgreSQLConnection{Address: postgreSQLAddress, UserName: "collector", Password: "password", DBName: "metrics_collector"}

	} else {
		Storage = &str.MemStorage{}
	}

	err := Storage.Init()
	if err != nil {
		panic(err)
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
			App.Logger.Errorln("Error when converting string to int: %v", err)
		}
	}

	App.FileStore, envExists = os.LookupEnv("FILE_STORAGE_PATH")
	if !(envExists) {
		App.FileStore = *fileStorePathFlag
	}

	restoreEnv, envExists := os.LookupEnv("RESTORE")
	if !(envExists) {
		restore = *restoreFlag
	} else {
		restore, err = strconv.ParseBool(restoreEnv)
		if err != nil {
			App.Logger.Errorln("Error when converting string to bool: %s", err)
		}
	}
	if restore && (App.FileStore != "") {
		App.Store()
	}

	if (storeInterval != 0) && (App.FileStore != "") {
		go App.SaveMetrics(time.Duration(storeInterval))
	} else if (storeInterval == 0) && (App.FileStore != "") {
		App.Backup = true
	}
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", App.HTMLMetrics())
		r.Get("/value/{metricType}/{metricName}", App.GetMetricPath())
		r.Post("/update/{metricType}/{metricName}/{metricValue}", App.UpdateValuePath())
		r.Post("/value/", App.GetMetric())
		r.Post("/update/", App.UpdateValue())
		r.Get("/ping", App.CheckStorageConnection())
	})

	err = http.ListenAndServe(serverAddress, App.WithLoggerZipper(r))
	if err != nil {
		App.Logger.Fatalw(err.Error(), "event", "start server")
	}
}