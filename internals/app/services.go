package app

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/connector"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/connectorconfig"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/externalconfig"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/history"

	// "github.com/myrteametrics/myrtea-engine-api/v5/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/modeler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier/notification"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/search"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/roles"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tasker"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// InitializeRepositories initialize all myrtea Postgresql repositories
func initRepositories() {
	dbClient := postgres.DB()
	users.ReplaceGlobals(users.NewPostgresRepository(dbClient))
	roles.ReplaceGlobals(roles.NewPostgresRepository(dbClient))
	permissions.ReplaceGlobals(permissions.NewPostgresRepository(dbClient))
	// groups.ReplaceGlobals(groups.NewPostgresRepository(dbClient))
	fact.ReplaceGlobals(fact.NewPostgresRepository(dbClient))
	situation.ReplaceGlobals(situation.NewPostgresRepository(dbClient))
	scheduler.ReplaceGlobalRepository(scheduler.NewPostgresRepository(dbClient))
	notification.ReplaceGlobals(notification.NewPostgresRepository(dbClient))
	issues.ReplaceGlobals(issues.NewPostgresRepository(dbClient))
	rootcause.ReplaceGlobals(rootcause.NewPostgresRepository(dbClient))
	action.ReplaceGlobals(action.NewPostgresRepository(dbClient))
	draft.ReplaceGlobals(draft.NewPostgresRepository(dbClient))
	search.ReplaceGlobals(search.NewPostgresRepository(dbClient))
	calendar.ReplaceGlobals(calendar.NewPostgresRepository(dbClient))
	connector.ReplaceGlobals(connector.NewPostgresRepository(dbClient))
	rule.ReplaceGlobals(rule.NewPostgresRepository(dbClient))
	modeler.ReplaceGlobals(modeler.NewPostgresRepository(dbClient))
	externalconfig.ReplaceGlobals(externalconfig.NewPostgresRepository(dbClient))
	connectorconfig.ReplaceGlobals(connectorconfig.NewPostgresRepository(dbClient))
	history.ReplaceGlobals(history.New(dbClient))
}

func initServices() {
	initCoordinator()
	initNotifier()
	initScheduler()
	initTasker()
	initCalendars()
}

func stopServices() {
	tasker.T().StopBatchProcessor()
	scheduler.S().C.Stop()
}

func initNotifier() {
	notifier.ReplaceGlobals(notifier.NewNotifier())
}

func initScheduler() {
	scheduler.ReplaceGlobals(scheduler.NewScheduler())
	err := scheduler.S().Init()
	if err != nil {
		zap.L().Error("Couldn't init fact scheduler", zap.Error(err))
	} else {
		if viper.GetBool("ENABLE_CRONS_ON_START") {
			scheduler.S().C.Start()
		}

	}
}
func initTasker() {
	tasker.ReplaceGlobals(tasker.NewTasker())
	tasker.T().StartBatchProcessor()
}

func initCalendars() {
	calendar.Init()

}

func initCoordinator() {
	zap.L().Info("Initialize Coordinator...")

	models, err := modeler.R().GetAll()
	if err != nil {
		zap.L().Error("Fetching models", zap.Error(err))
		return
	}

	instanceName := viper.GetString("INSTANCE_NAME")
	if err = coordinator.InitInstance(instanceName, models); err != nil {
		zap.L().Fatal("Intialisation of coordinator master", zap.Error(err))
	}
	if viper.GetBool("ENABLE_CRONS_ON_START") {
		for _, li := range coordinator.GetInstance().LogicalIndices {
			cron := li.GetCron()
			if cron != nil {
				cron.Start()
			}
		}
	}
}
