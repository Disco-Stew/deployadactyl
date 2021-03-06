package service_test

import (
	"io"
	"os"

	"github.com/compozed/deployadactyl/artifetcher"
	"github.com/compozed/deployadactyl/artifetcher/extractor"
	"github.com/compozed/deployadactyl/config"
	"github.com/compozed/deployadactyl/controller"
	"github.com/compozed/deployadactyl/controller/deployer"
	"github.com/compozed/deployadactyl/controller/deployer/bluegreen"
	"github.com/compozed/deployadactyl/controller/deployer/bluegreen/pusher"
	"github.com/compozed/deployadactyl/eventmanager"
	I "github.com/compozed/deployadactyl/interfaces"
	"github.com/compozed/deployadactyl/logger"
	"github.com/compozed/deployadactyl/mocks"
	"github.com/compozed/deployadactyl/randomizer"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type Creator struct {
	config       config.Config
	eventManager I.EventManager
	logger       *logging.Logger
	writer       io.Writer
	fileSystem   *afero.Afero
}

func New(level string, configFilename string) (Creator, error) {
	cfg, err := config.Custom(os.Getenv, configFilename)
	if err != nil {
		return Creator{}, err
	}

	l, err := getLevel(level)
	if err != nil {
		return Creator{}, err
	}

	logger := logger.DefaultLogger(GinkgoWriter, l, "creator")

	eventManager := eventmanager.NewEventManager(logger)

	return Creator{
		config:       cfg,
		eventManager: eventManager,
		logger:       logger,
		writer:       GinkgoWriter,
		fileSystem:   &afero.Afero{Fs: afero.NewMemMapFs()},
	}, nil
}

func (c Creator) CreateControllerHandler() *gin.Engine {
	d := c.CreateController()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithWriter(c.CreateWriter()))
	r.Use(gin.ErrorLogger())

	r.POST(ENDPOINT, d.Deploy)

	return r
}

func (c Creator) CreateController() controller.Controller {
	return controller.Controller{
		Deployer: c.CreateDeployer(),
		Log:      c.CreateLogger(),
	}
}

func (c Creator) CreateRandomizer() I.Randomizer {
	return randomizer.Randomizer{}
}

func (c Creator) CreateDeployer() I.Deployer {
	return deployer.Deployer{
		Config:      c.CreateConfig(),
		BlueGreener: c.CreateBlueGreener(),
		Fetcher: &artifetcher.Artifetcher{
			FileSystem: c.CreateFileSystem(),
			Extractor: &extractor.Extractor{
				Log:        c.CreateLogger(),
				FileSystem: c.CreateFileSystem(),
			},
			Log: c.CreateLogger(),
		},
		Prechecker:   c.CreatePrechecker(),
		EventManager: c.CreateEventManager(),
		Randomizer:   c.CreateRandomizer(),
		Log:          c.CreateLogger(),
		FileSystem:   c.CreateFileSystem(),
	}
}

func (c Creator) createFetcher() I.Fetcher {
	return &artifetcher.Artifetcher{
		FileSystem: c.CreateFileSystem(),
		Extractor: &extractor.Extractor{
			Log:        c.CreateLogger(),
			FileSystem: c.CreateFileSystem(),
		},
		Log: c.CreateLogger(),
	}
}

func (c Creator) CreatePusher() (I.Pusher, error) {
	courier := &mocks.Courier{}

	courier.LoginCall.Returns.Output = []byte("logged in\t")
	courier.LoginCall.Returns.Error = nil
	courier.DeleteCall.Returns.Output = []byte("deleted app\t")
	courier.DeleteCall.Returns.Error = nil
	courier.PushCall.Returns.Output = []byte("pushed app\t")
	courier.PushCall.Returns.Error = nil
	courier.RenameCall.Returns.Output = []byte("renamed app\t")
	courier.RenameCall.Returns.Error = nil
	courier.MapRouteCall.Returns.Output = []byte("mapped route\t")
	courier.MapRouteCall.Returns.Error = nil
	courier.ExistsCall.Returns.Bool = false
	courier.CleanUpCall.Returns.Error = nil

	p := &pusher.Pusher{
		Courier: courier,
		Log:     c.CreateLogger(),
	}

	return p, nil
}

func (c Creator) CreateEventManager() I.EventManager {
	return c.eventManager
}

func (c Creator) CreateLogger() *logging.Logger {
	return c.logger
}

func (c Creator) CreateConfig() config.Config {
	return c.config
}

func (c Creator) CreatePrechecker() I.Prechecker {
	prechecker := &mocks.Prechecker{}

	prechecker.AssertAllFoundationsUpCall.Returns.Error = nil

	return prechecker
}

func (c Creator) CreateWriter() io.Writer {
	return c.writer
}

func (c Creator) CreateBlueGreener() I.BlueGreener {
	return bluegreen.BlueGreen{
		PusherCreator: c,
		Log:           c.CreateLogger(),
	}
}

func (c Creator) CreateFileSystem() *afero.Afero {
	return c.fileSystem
}

func getLevel(level string) (logging.Level, error) {
	if level != "" {
		l, err := logging.LogLevel(level)
		if err != nil {
			return 0, errors.Errorf("unable to get log level: %s. error: %s", level, err.Error())
		}
		return l, nil
	}

	return logging.INFO, nil
}
