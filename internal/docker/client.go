package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/pablopin/docker-image-checker/internal/model"
)

// DockerClient implementa la interfaz Client para Docker
type DockerClient struct {
	cli *client.Client
}

// NewDockerClient crea un nuevo cliente Docker
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerClient{cli: cli}, nil
}

// ListContainers lista todos los contenedores en ejecución
func (dc *DockerClient) ListContainers(ctx context.Context) ([]model.Container, error) {
	containers, err := dc.cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]model.Container, 0, len(containers))
	for _, c := range containers {
		name := strings.TrimPrefix(c.Names[0], "/")

		container := model.Container{
			ID:        c.ID,
			Name:      name,
			ImageName: c.Image,
			ImageID:   c.ImageID,
			Status:    c.Status,
		}

		result = append(result, container)
	}

	return result, nil
}

// GetImageInfo obtiene información detallada de una imagen
func (dc *DockerClient) GetImageInfo(ctx context.Context, imageID string) (*types.ImageInspect, error) {
	inspect, _, err := dc.cli.ImageInspectWithRaw(ctx, imageID)
	return &inspect, err
}

// DistributionInspect inspecciona una imagen en el registro remoto
func (dc *DockerClient) DistributionInspect(ctx context.Context, image string) (registry.DistributionInspect, []byte, error) {
	inspect, err := dc.cli.DistributionInspect(ctx, image, "")
	return inspect, nil, err
}

// Close cierra la conexión del cliente
func (dc *DockerClient) Close() error {
	return dc.cli.Close()
}
