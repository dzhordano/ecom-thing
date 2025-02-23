package boot

import (
	"context"
	"log"

	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type options struct {
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
	ConfigName         string
}

type Option func(o *options)

// WithUnaryInterceptors - add unary interceptors.
func WithUnaryInterceptors(inter ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.UnaryInterceptors = append(o.UnaryInterceptors, inter...)
	}
}

// WithStreamInterceptors - add stream interceptors.
func WithStreamInterceptors(inter ...grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.StreamInterceptors = append(o.StreamInterceptors, inter...)
	}
}

// WithConfigName - override config entry name
func WithConfigName(inter ...grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.StreamInterceptors = append(o.StreamInterceptors, inter...)
	}
}

type Boot struct {
	*rkboot.Boot
	options options
}

// NewBoot - return new Boot
func NewBoot(config []byte, opts ...Option) *Boot {
	// Загрузжаем basic entries из конфигурации (boot.yaml).
	// rkentry.BootstrapBuiltInEntryFromYAML(config)
	// rkentry.BootstrapPluginEntryFromYAML(config)
	// rkentry.BootstrapUserEntryFromYAML(config)
	// rkentry.BootstrapWebFrameEntryFromYAML(config)

	// Загрузжаем entries из конфигурации (boot.yaml).
	boot := rkboot.NewBoot(
		rkboot.WithBootConfigRaw(config),
	)

	options := options{
		ConfigName: "config",
	}
	for _, opt := range opts {
		opt(&options)
	}

	return &Boot{Boot: boot, options: options}
}

type service interface {
	Name() string
	DBName() string
	GrpcRegFunc() func(server *grpc.Server)
	MigrateModels() interface{}
	// GwRegFunc() func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
}

func (b *Boot) Run(ctx context.Context, services ...service) {
	for _, svc := range services {
		b.registerService(svc)
		// pgEntry := rkpostgres.GetPostgresEntry(svc.DBName())

		// pgEntry.Bootstrap(ctx)

		// invDb := pgEntry.GetDB("inventory-db-1")
		// if !invDb.DryRun {
		// 	invDb.AutoMigrate(svc.MigrateModels())
		// }
	}

	// Bootstrap entries as sequence of plugin, user defined and web framework
	b.Bootstrap(ctx)

	log.Println("start serve")

	// Ждем сигнала выключения
	b.WaitForShutdownSig(ctx)
}

func (b *Boot) registerService(s service) {
	// Получение GrpcEntry
	grpcEntry := rkgrpc.GetGrpcEntry(s.Name())
	// Регистрация gRPC сервера
	grpcEntry.AddRegFuncGrpc(s.GrpcRegFunc())
	// Регистрация gRPC-Gateway proxy
	// grpcEntry.AddRegFuncGw(s.GwRegFunc())
	// Регистрация middleware
	grpcEntry.AddUnaryInterceptors(
		b.options.UnaryInterceptors...,
	)
	grpcEntry.AddStreamInterceptors(
		b.options.StreamInterceptors...,
	)
}

// Config - возвращает конфиг приложения
func (b *Boot) Config() *viper.Viper {
	return rkentry.GlobalAppCtx.GetConfigEntry(b.options.ConfigName).Viper
}
