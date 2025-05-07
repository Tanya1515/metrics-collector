// Server is used to process http-requests and gather data from agent.
// Processed data can be saved to in-memory storage or to PostgreSQL database.
// Processed data consists of data of runtime package
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"

	psql "github.com/Tanya1515/metrics-collector.git/cmd/storage/postgresql"
	str "github.com/Tanya1515/metrics-collector.git/cmd/storage/structure"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

// Application - data type to describe the server work
type Application struct {
	// Storage - object interface for saving data.
	Storage storage.RepositoryInterface
	// Logger - logger for saving info about all events in the application.
	Logger zap.SugaredLogger
	// SecretKey - key for chacking hash of incoming data
	SecretKey string
	// CryptoKey - key for encrypting incoming data (asymmetrical encryption)
	CryptoKey string
}

func init() {
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
	postgreSQLFlag = flag.String("d", "", "credentials for database")
	storeIntervalFlag = flag.Int("i", 1, "time duration for saving metrics")
	fileStorePathFlag = flag.String("f", "/tmp/metrics-db.json", "filename for storing metrics")
	restoreFlag = flag.Bool("r", true, "store all info")
	secretKeyFlag = flag.String("k", "", "secret key for hash")
	cryptoKeyPathFlag = flag.String("crypto-key", "", "path to key for asymmetrical encryption")
	configFilePathFlag = flag.String("config", "", "path to config file for the application")
}

var (
	serverAddressFlag  *string
	storeIntervalFlag  *int
	fileStorePathFlag  *string
	restoreFlag        *bool
	postgreSQLFlag     *string
	secretKeyFlag      *string
	cryptoKeyPathFlag  *string
	configFilePathFlag *string
	buildVersion       string = "N/A"
	buildDate          string = "N/A"
	buildCommit        string = "N/A"
)

func main() {
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
	var Storage storage.RepositoryInterface
	var err error
	flag.Parse()

	configFilePath, envExists := os.LookupEnv("CONFIG")
	if !(envExists) {
		configFilePath = *configFilePathFlag
	}

	configApp := new(data.ConfigApp)
	if configFilePath != "" {
		config, err := os.ReadFile(configFilePath)
		if err != nil {
			fmt.Println("Error while parsing config file: ", err)
		}

		err = json.Unmarshal(config, configApp)
		if err != nil {
			fmt.Println("Error while unmarshaling config data: ", err)
		}
	}

	serverAddress, envExists := os.LookupEnv("ADDRESS")
	if !(envExists) {
		serverAddress = *serverAddressFlag
	}

	if serverAddress == "localhost:8080" && configFilePath != "" {
		serverAddress = configApp.ServerAddress
	}

	postgreSQLAddress, envExists := os.LookupEnv("DATABASE_DSN")
	if !(envExists) {
		postgreSQLAddress = *postgreSQLFlag
	}
	if postgreSQLAddress == "" && configFilePath != "" {
		postgreSQLAddress = configApp.PostgreSQL
	}

	cryptoKeyPath, envExists := os.LookupEnv("CRYPTO_KEY")
	if !(envExists) {
		cryptoKeyPath = *cryptoKeyPathFlag
	}

	if cryptoKeyPath == "" && configFilePath != "" {
		cryptoKeyPath = configApp.CryptoKeyPath
	}

	var storeInterval int

	storeIntervalEnv, envExists := os.LookupEnv("STORE_INTERVAL")
	if !(envExists) {
		storeInterval = *storeIntervalFlag
	} else {
		storeInterval, err = strconv.Atoi(storeIntervalEnv)
		if err != nil {
			fmt.Println("Error when converting string to int:", err)
		}
	}

	if storeInterval == 300 && configFilePath != "" {
		if configApp.StoreInterval != "" {
			storeInterval, err = strconv.Atoi(strings.Split(configApp.StoreInterval, "s")[0])
			if err != nil {
				fmt.Println("Error when converting string to int: ", err)
			}
		}
	}

	fileStore, envExists := os.LookupEnv("FILE_STORAGE_PATH")
	if !(envExists) {
		fileStore = *fileStorePathFlag
	}

	if fileStore == "/tmp/metrics-db.json" && configFilePath != "" {
		fileStore = configApp.FileStorePath
	}

	var restore bool
	restoreEnv, envExists := os.LookupEnv("RESTORE")
	if !(envExists) {
		restore = *restoreFlag
	} else {
		restore, err = strconv.ParseBool(restoreEnv)
		if err != nil {
			fmt.Println("Error when converting string to bool: ", err)
		}
	}

	if restore && configFilePath != "" {
		restore = configApp.Restore
	}

	Gctx, cancelG := context.WithCancel(context.Background())
	shutdown := make(chan struct{})
	if postgreSQLAddress != "" {
		postgreSQLAddrPortDatabase := strings.Split((strings.Split((strings.Split(postgreSQLAddress, "@"))[1], "?"))[0], ":")
		postgreSQLDatabase := "postgres"
		postgreSQLPort := "5432"
		if len(postgreSQLAddrPortDatabase) == 2 {
			postgreSQLPortDatabase := strings.Split(postgreSQLAddrPortDatabase[1], "/")
			if len(postgreSQLPortDatabase) == 2 {
				postgreSQLDatabase = postgreSQLPortDatabase[1]
			}
			postgreSQLPort = postgreSQLPortDatabase[0]
		}
		postgreSQLAddr := postgreSQLAddrPortDatabase[0]
		Storage = &psql.PostgreSQLConnection{StoreType: storage.StoreType{Restore: restore, BackupTimer: storeInterval, FileStore: fileStore}, Address: postgreSQLAddr, Port: postgreSQLPort, UserName: "postgres", Password: "postgres", DBName: postgreSQLDatabase}
	} else {
		Storage = &str.MemStorage{StoreType: storage.StoreType{Restore: restore, BackupTimer: storeInterval, FileStore: fileStore}}
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	secretKeyHash, secretKeyExists := os.LookupEnv("KEY")
	if !(secretKeyExists) {
		secretKeyHash = *secretKeyFlag
	}

	if secretKeyHash == "" && configFilePath != "" {
		secretKeyHash = configApp.SecretKey
	}

	App := Application{Storage: Storage, Logger: *logger.Sugar(), SecretKey: secretKeyHash}

	App.Logger.Infow(
		"Starting server",
		"addr", serverAddress,
	)

	if cryptoKeyPath != "" {
		file, err := os.Open(cryptoKeyPath)
		if err != nil {
			App.Logger.Errorln("Error while openning file with crypto key: ", err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			App.CryptoKey += scanner.Text() + "\n"
		}
	}

	err = Storage.Init(shutdown, Gctx)
	if err != nil {
		fmt.Println(err)
	}

	commonMiddlewares := []data.Middleware{}
	if secretKeyHash != "" {
		commonMiddlewares = append(commonMiddlewares, App.MiddlewareHash, App.MiddlewareZipper, App.MiddlewareLogger)
	} else {
		commonMiddlewares = append(commonMiddlewares, App.MiddlewareZipper, App.MiddlewareLogger)
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

	go func() {
		App.Logger.Infoln("Starting server for pprof")
		err = http.ListenAndServe("localhost:8081", nil)
		if err != nil {
			App.Logger.Fatalw(err.Error(), "event", "start server for pprof")
		}
	}()

	srv := http.Server{Addr: serverAddress, Handler: r}

	gracefulSutdown := make(chan os.Signal, 1)

	signal.Notify(gracefulSutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-gracefulSutdown
		cancelG()
		if (fileStore != "") && (storeInterval != 0) {
			<-shutdown
		} else {
			close(shutdown)
		}
		srv.Shutdown(context.Background())
	}()

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		App.Logger.Fatalw(err.Error(), "event", "start server")
	}

	<-shutdown
	App.Logger.Infoln("Server successfully shutdown!")
}
