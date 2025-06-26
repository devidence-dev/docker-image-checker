package docker

import (
	"context"

	"github.com/pablopin/docker-image-checker/internal/model"
)

// CheckStrategy define la interfaz para diferentes estrategias de verificación
type CheckStrategy interface {
	Check(ctx context.Context, container model.Container) (*model.UpdateInfo, error)
	CanHandle(container model.Container) bool
}

// Client define la interfaz del cliente Docker
type Client interface {
	ListContainers(ctx context.Context) ([]model.Container, error)
	Close() error
}

// Checker implementa la lógica principal de verificación usando Strategy pattern
type Checker struct {
	client     Client
	strategies []CheckStrategy
}

// NewChecker crea un nuevo verificador
func NewChecker(client Client) *Checker {
	checker := &Checker{
		client:     client,
		strategies: make([]CheckStrategy, 0),
	}

	// Registrar estrategias por defecto
	checker.RegisterStrategy(NewRegistryStrategy())
	
	return checker
}

// RegisterStrategy registra una nueva estrategia
func (c *Checker) RegisterStrategy(strategy CheckStrategy) {
	c.strategies = append(c.strategies, strategy)
}

// CheckAll verifica todos los contenedores
func (c *Checker) CheckAll(ctx context.Context) (*model.CheckReport, error) {
	containers, err := c.client.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	report := &model.CheckReport{
		Total:     len(containers),
		Available: make([]model.UpdateInfo, 0),
		Failed:    make([]model.UpdateInfo, 0),
		UpToDate:  make([]model.UpdateInfo, 0),
	}

	for _, container := range containers {
		updateInfo, err := c.checkContainer(ctx, container)
		if err != nil {
			updateInfo = &model.UpdateInfo{
				Container: container,
				Error:     err,
			}
			report.Failed = append(report.Failed, *updateInfo)
			continue
		}

		if updateInfo.Error != nil {
			report.Failed = append(report.Failed, *updateInfo)
		} else if !updateInfo.IsUpToDate {
			report.Available = append(report.Available, *updateInfo)
		} else {
			report.UpToDate = append(report.UpToDate, *updateInfo)
		}
	}

	return report, nil
}

// checkContainer verifica un contenedor específico usando las estrategias disponibles
func (c *Checker) checkContainer(ctx context.Context, container model.Container) (*model.UpdateInfo, error) {
	for _, strategy := range c.strategies {
		if strategy.CanHandle(container) {
			return strategy.Check(ctx, container)
		}
	}

	// Si no hay estrategia disponible, marcar como no verificable
	return &model.UpdateInfo{
		Container:  container,
		IsUpToDate: true, // Asumimos que está actualizado si no podemos verificar
		Error:      nil,
	}, nil
}

// Close cierra el cliente
func (c *Checker) Close() error {
	return c.client.Close()
}
