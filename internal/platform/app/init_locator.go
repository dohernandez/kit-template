package app

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/bool64/zapctxd"
	"github.com/dohernandez/kit-template/internal/platform/config"
	grpcZapLogger "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcCtxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcOpentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	_ "github.com/jackc/pgx/v4/stdlib" // nolint: gci // Postgres driver
	"github.com/jmoiron/sqlx"
	clock "github.com/nhatthm/go-clock"
	clockSvc "github.com/nhatthm/go-clock/service"
	"github.com/opencensus-integrations/ocsql"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const driver = "pgx"

// Locator defines application resources.
type Locator struct {
	Config  *config.Config
	DBx     *sqlx.DB
	Storage *sqluct.Storage

	logger *zapctxd.Logger
	ctxd.LoggerProvider

	clockSvc.ClockProvider

	GRPCUnitaryInterceptors []grpc.UnaryServerInterceptor

	// use cases
}

// Option sets up service locator.
type Option func(l *Locator)

// NewServiceLocator creates application locator.
func NewServiceLocator(cfg *config.Config, opts ...Option) (*Locator, error) {
	l := Locator{
		Config:        cfg,
		ClockProvider: clock.New(),
	}

	for _, o := range opts {
		o(&l)
	}

	var err error

	// logger stuff
	if l.LoggerProvider == nil {
		l.logger = zapctxd.New(zapctxd.Config{
			Level:   cfg.Log.Level,
			DevMode: cfg.IsDev(),
			FieldNames: ctxd.FieldNames{
				Timestamp: "timestamp",
				Message:   "message",
			},
			StripTime: cfg.Log.LockTime,
			Output:    cfg.Log.Output,
		})

		l.LoggerProvider = l.logger
	}

	// Database stuff.
	l.Config.PostgresDB.DriverName = driver

	l.DBx, err = makeDBx(cfg.PostgresDB)
	if err != nil {
		return nil, err
	}

	l.Storage = makeStorage(l.DBx, l.CtxdLogger())

	l.GRPCUnitaryInterceptors = append(l.GRPCUnitaryInterceptors, []grpc.UnaryServerInterceptor{
		// recovering from panic
		grpcRecovery.UnaryServerInterceptor(),
		// adding tracing
		grpcOpentracing.UnaryServerInterceptor(),
		// adding logger
		grpcCtxtags.UnaryServerInterceptor(grpcCtxtags.WithFieldExtractor(grpcCtxtags.CodeGenRequestFieldExtractor)),
		grpcZapLogger.UnaryServerInterceptor(l.ZapLogger()),
	}...)

	// setting up use cases dependencies
	l.setupUsecaseDependencies()

	return &l, nil
}

// makeDBx initializes database.
func makeDBx(cfg config.DBConfig) (*sqlx.DB, error) {
	db, err := makeDB(cfg)
	if err != nil {
		return nil, err
	}

	return sqlx.NewDb(db, cfg.DriverName), nil
}

// makeDB initializes database.
func makeDB(cfg config.DBConfig) (*sql.DB, error) {
	driverName, err := ocsql.Register(cfg.DriverName,
		ocsql.WithQuery(true),
		ocsql.WithRowsClose(true),
		ocsql.WithRowsAffected(true),
		ocsql.WithAllowRoot(true),
	)
	if err != nil {
		return nil, err
	}

	ocsql.RegisterAllViews()

	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		return nil, err
	}

	ocsql.RecordStats(db, time.Second)

	db.SetConnMaxLifetime(cfg.MaxLifetime)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// makeStorage initializes database storage.
func makeStorage(
	db *sqlx.DB,
	logger ctxd.Logger,
) *sqluct.Storage {
	st := sqluct.NewStorage(db)

	st.Format = squirrel.Dollar
	st.OnError = func(ctx context.Context, err error) {
		logger.Error(ctx, "storage failure", "error", err)
	}

	return st
}

func (l *Locator) setupUsecaseDependencies() {
}

// ZapLogger returns *zap.Logger that used in Logger.
func (l *Locator) ZapLogger() *zap.Logger {
	return l.logger.ZapLogger()
}
