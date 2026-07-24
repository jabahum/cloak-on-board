package server

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/approvals"
	"github.com/jabahum/keycloak-onboarder/backend/internal/config"
	"github.com/jabahum/keycloak-onboarder/backend/internal/middleware"
	"github.com/jabahum/keycloak-onboarder/backend/internal/notifications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/provisioning"
	"github.com/jabahum/keycloak-onboarder/backend/internal/settings"
	"github.com/jabahum/keycloak-onboarder/backend/internal/templates"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config config.Config
	db     *pgxpool.Pool
	router *gin.Engine
}

type workflowExecutor struct {
	provisioner *provisioning.Provisioner
	manager     *provisioning.ClientManager
}

func (e workflowExecutor) Provision(ctx context.Context, id string) (string, error) {
	job, err := e.provisioner.ProvisionApplication(ctx, id)
	return job.ID, err
}
func (e workflowExecutor) Update(ctx context.Context, id string, req applications.UpdateApplicationRequest) error {
	_, err := e.manager.UpdateApplication(ctx, id, req)
	return err
}
func (e workflowExecutor) Delete(ctx context.Context, id string, deleteKeycloak bool) error {
	return e.manager.DeleteApplication(ctx, id, deleteKeycloak)
}

func (s *Server) approvalService(n *notifications.Service) *approvals.Service {
	appRepo := applications.NewRepository(s.db)
	appService := applications.NewService(appRepo)
	settingsRepo := settings.NewRepository(s.db)
	settingsService := settings.NewService(settingsRepo, s.config)
	jobService := provisioning.NewService(provisioning.NewRepository(s.db))
	provisioner := provisioning.NewProvisioner(appService, settingsService, jobService)
	manager := provisioning.NewClientManager(appService, settingsService, jobService)
	return approvals.NewService(s.db, appService, n, workflowExecutor{provisioner: provisioner, manager: manager})
}

func New(cfg config.Config, db *pgxpool.Pool) *Server {
	s := &Server{
		config: cfg,
		db:     db,
		router: gin.New(),
	}

	s.router.Use(middleware.RequestID())
	s.router.Use(middleware.Recovery())
	s.router.Use(middleware.CORS(cfg.AllowedOrigins))
	s.router.Use(middleware.Logger())

	s.registerRoutes()

	return s
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.AppPort)
	return s.router.Run(addr)
}

func (s *Server) applicationHandler() *applications.Handler {
	appRepo := applications.NewRepository(s.db)
	appService := applications.NewService(appRepo)

	settingsRepo := settings.NewRepository(s.db)
	settingsService := settings.NewService(settingsRepo, s.config)

	jobRepo := provisioning.NewRepository(s.db)
	jobService := provisioning.NewService(jobRepo)

	provisioner := provisioning.NewProvisioner(
		appService,
		settingsService,
		jobService,
	)

	handler := applications.NewHandler(appService)
	handler.SetProvisioner(func(ctx context.Context, applicationID string) (any, error) {
		return provisioner.ProvisionApplication(ctx, applicationID)
	})
	handler.SetClientManager(provisioning.NewClientManager(appService, settingsService, jobService))

	return handler
}

func (s *Server) templateHandler() *templates.Handler {
	repo := templates.NewRepository(s.db)
	service := templates.NewService(repo)
	return templates.NewHandler(service)
}

func (s *Server) settingsHandler() *settings.Handler {
	repo := settings.NewRepository(s.db)
	service := settings.NewService(repo, s.config)
	return settings.NewHandler(service)
}

func (s *Server) provisioningHandler() *provisioning.Handler {
	repo := provisioning.NewRepository(s.db)
	service := provisioning.NewService(repo)
	return provisioning.NewHandler(service)
}
