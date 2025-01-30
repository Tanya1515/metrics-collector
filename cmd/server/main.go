package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"
	"fmt"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"
	str "github.com/Tanya1515/metrics-collector.git/cmd/storage/structure"
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
)

// при синхронной записи сбрасывается значение PollCount

func init() {
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
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

	Storage = &str.MemStorage{}

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
			App.Logger.Errorln("Error when converting string to bool: %s", err)
		}
	}

	err = Storage.Init(restore, fileStore, time.Duration(storeInterval))
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", App.HTMLMetrics())
		r.Get("/value/{metricType}/{metricName}", App.GetMetricPath())
		r.Post("/update/{metricType}/{metricName}/{metricValue}", App.UpdateValuePath())
		r.Post("/value/", App.GetMetric())
		r.Post("/update/", App.UpdateValue())
	})

	err = http.ListenAndServe(serverAddress, App.WithLoggerZipper(r))
	if err != nil {
		fmt.Println(err)
		App.Logger.Fatalw(err.Error(), "event", "start server")
	}
}
