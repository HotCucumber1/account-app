// Package testenv starts MySQL, RabbitMQ and module containers for end-to-end tests.
package testenv

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	CommandService        = "service"
	CommandMessageHandler = "message-handler"

	mysqlPassword  = "integration:password@local"
	brokerUser     = "integration"
	brokerPassword = "integration-password"
	grpcPort       = "8081/tcp"
	httpPort       = "8082/tcp"
)

// Module describes a service module to start inside the environment.
type Module struct {
	Name string
	// Commands are started in order; the default is message-handler and service.
	Commands []string
	// Env contains additional variables without the module prefix.
	Env map[string]string
}

// App exposes addresses and logs of a started module.
type App struct {
	GRPCAddress string
	HTTPAddress string
	Logs        *LogBuffer
	DB          mysql.TransactionalClient
}

// Environment owns all containers of an end-to-end test run.
type Environment struct {
	network *testcontainers.DockerNetwork
	rootDSN *mysqldriver.Config
	amqpURL string
	apps    map[string]*App
	closers []func() error
}

// Start launches infrastructure and modules. Message handlers of every module start before
// any service, so no event published by a service is lost before its consumer queue exists.
func Start(ctx context.Context, modules ...Module) (*Environment, error) {
	e := &Environment{apps: make(map[string]*App, len(modules))}
	err := e.start(ctx, modules)
	if err != nil {
		return e, errors.Join(err, e.Close())
	}
	return e, nil
}

func (e *Environment) start(ctx context.Context, modules []Module) error {
	var err error
	e.network, err = network.New(ctx)
	if err != nil {
		return err
	}
	e.closers = append(e.closers, func() error {
		closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return e.network.Remove(closeCtx)
	})
	err = e.startMySQL(ctx)
	if err != nil {
		return err
	}
	err = e.startBroker(ctx)
	if err != nil {
		return err
	}
	for _, module := range modules {
		err = e.prepareModule(ctx, module.Name)
		if err != nil {
			return err
		}
	}
	return e.startModules(ctx, modules)
}

// App returns a started module.
func (e *Environment) App(name string) *App {
	app, ok := e.apps[name]
	if !ok {
		panic(fmt.Sprintf("testenv: module %q is not started", name))
	}
	return app
}

// AMQPURL is the broker URL reachable from the test process.
func (e *Environment) AMQPURL() string {
	return e.amqpURL
}

// Logs returns logs of every module, used to diagnose failures.
func (e *Environment) Logs() string {
	var builder strings.Builder
	for name, app := range e.apps {
		builder.WriteString("==== " + name + "\n")
		builder.WriteString(app.Logs.String())
	}
	return builder.String()
}

// Close stops containers in reverse order.
func (e *Environment) Close() error {
	var errs []error
	for i := len(e.closers) - 1; i >= 0; i-- {
		errs = append(errs, e.closers[i]())
	}
	e.closers = nil
	return errors.Join(errs...)
}

func (e *Environment) startMySQL(ctx context.Context) error {
	container, err := tcmysql.Run(ctx, "mysql:8.4",
		tcmysql.WithDatabase("integration"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword(mysqlPassword),
		network.WithNetwork([]string{"mysql"}, e.network),
	)
	if container != nil {
		e.closers = append(e.closers, func() error {
			return testcontainers.TerminateContainer(container)
		})
	}
	if err != nil {
		return err
	}
	dsn, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		return err
	}
	e.rootDSN, err = mysqldriver.ParseDSN(dsn)
	return err
}

func (e *Environment) startBroker(ctx context.Context) error {
	broker, err := e.startContainer(ctx, testcontainers.ContainerRequest{
		Image:          "rabbitmq:4.1-alpine",
		ExposedPorts:   []string{"5672/tcp"},
		Networks:       []string{e.network.Name},
		NetworkAliases: map[string][]string{e.network.Name: {"rabbitmq"}},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": brokerUser,
			"RABBITMQ_DEFAULT_PASS": brokerPassword,
		},
		WaitingFor: wait.ForLog("Server startup complete").WithStartupTimeout(time.Minute),
	})
	if err != nil {
		return err
	}
	address, err := containerAddress(ctx, broker, "5672/tcp")
	if err != nil {
		return err
	}
	e.amqpURL = (&url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(brokerUser, brokerPassword),
		Host:   address,
		Path:   "/",
	}).String()
	return nil
}

func (e *Environment) prepareModule(ctx context.Context, name string) error {
	root := mysql.NewConnector()
	err := root.Open(e.rootDSN.FormatDSN(), mysql.Config{MaxConnections: 1})
	if err != nil {
		return err
	}
	_, err = root.TransactionalClient().ExecContext(ctx, "testenv.create-database",
		"CREATE DATABASE IF NOT EXISTS `"+name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	err = errors.Join(err, root.Close())
	if err != nil {
		return err
	}

	dsn := e.rootDSN.Clone()
	dsn.DBName = name
	connector := mysql.NewConnector()
	err = connector.Open(dsn.FormatDSN(), mysql.Config{MaxConnections: 5})
	if err != nil {
		return err
	}
	e.closers = append(e.closers, connector.Close)
	e.apps[name] = &App{
		Logs: &LogBuffer{},
		DB:   connector.TransactionalClient(),
	}
	return nil
}

func (e *Environment) startModules(ctx context.Context, modules []Module) error {
	for _, phase := range []string{CommandMessageHandler, CommandService} {
		for _, module := range modules {
			if !module.has(phase) {
				continue
			}
			container, err := e.startApplication(ctx, module, phase)
			if err != nil {
				return fmt.Errorf("start %s %s: %w", module.Name, phase, err)
			}
			if phase != CommandService {
				continue
			}
			app := e.apps[module.Name]
			app.GRPCAddress, err = containerAddress(ctx, container, grpcPort)
			if err != nil {
				return err
			}
			app.HTTPAddress, err = containerAddress(ctx, container, httpPort)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m Module) has(command string) bool {
	commands := m.Commands
	if len(commands) == 0 {
		commands = []string{CommandMessageHandler, CommandService}
	}
	for _, c := range commands {
		if c == command {
			return true
		}
	}
	return false
}

func (e *Environment) startApplication(ctx context.Context, module Module, command string) (testcontainers.Container, error) {
	prefix := strings.ToUpper(module.Name) + "_"
	env := map[string]string{
		"MYSQL_ADDRESS":         "mysql:3306",
		"MYSQL_DATABASE":        module.Name,
		"MYSQL_USER":            e.rootDSN.User,
		"MYSQL_PASSWORD":        e.rootDSN.Passwd,
		"AMQP_URL":              "amqp://" + brokerUser + ":" + brokerPassword + "@rabbitmq:5672/",
		"HTTP_ADDRESS":          ":8082",
		"GRPC_ADDRESS":          ":8081",
		"APP_GRACE_TIMEOUT":     "1s",
		"RELAY_INTERVAL":        "50ms",
		"RELAY_PUBLISH_TIMEOUT": "1s",
	}
	for key, value := range module.Env {
		env[key] = value
	}
	prefixed := make(map[string]string, len(env))
	for key, value := range env {
		prefixed[prefix+key] = value
	}

	return e.startContainer(ctx, testcontainers.ContainerRequest{ // TODO упростить вообще все это
		Image:        "distributed-programming-" + module.Name + ":integration",
		Cmd:          []string{command},
		ExposedPorts: []string{grpcPort, httpPort},
		Networks:     []string{e.network.Name},
		NetworkAliases: map[string][]string{
			e.network.Name: {module.Name + "-" + command},
		},
		Env: prefixed,
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{e.apps[module.Name].Logs},
		},
		WaitingFor: wait.ForHTTP("/healthz").WithPort(httpPort).WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusNoContent
		}).WithStartupTimeout(time.Minute),
	})
}

func (e *Environment) startContainer(ctx context.Context, request testcontainers.ContainerRequest) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if container != nil {
		e.closers = append(e.closers, func() error {
			return testcontainers.TerminateContainer(container)
		})
	}
	return container, err
}

func containerAddress(ctx context.Context, container testcontainers.Container, port string) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}
	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(host, mappedPort.Port()), nil
}
