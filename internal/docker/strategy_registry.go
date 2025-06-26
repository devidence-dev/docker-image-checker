package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pablopin/docker-image-checker/internal/model"
)

// RegistryStrategy implementa verificación contra registros remotos
type RegistryStrategy struct {
	dockerClient *DockerClient
}

// NewRegistryStrategy crea una nueva estrategia de registro
func NewRegistryStrategy() *RegistryStrategy {
	// Para esta implementación, necesitamos acceso al cliente Docker
	// En una implementación más robusta, esto se inyectaría via DI
	client, _ := NewDockerClient()
	return &RegistryStrategy{
		dockerClient: client,
	}
}

// CanHandle determina si esta estrategia puede manejar el contenedor
func (rs *RegistryStrategy) CanHandle(container model.Container) bool {
	// Esta estrategia puede manejar cualquier imagen que no sea local/build
	if strings.Contains(container.ImageName, "<none>") {
		return false
	}

	// Evitar imágenes que parezcan builds locales
	if strings.HasPrefix(container.ImageName, "localhost/") ||
		strings.HasPrefix(container.ImageName, "local/") ||
		strings.Contains(container.ImageName, "build") {
		return false
	}

	return true
}

// Check verifica si la imagen está actualizada
func (rs *RegistryStrategy) Check(ctx context.Context, container model.Container) (*model.UpdateInfo, error) {
	updateInfo := &model.UpdateInfo{
		Container:      container,
		CurrentVersion: "unknown",
		LatestVersion:  "unknown",
		IsUpToDate:     true,
	}

	// Obtener información de la imagen local
	localImage, err := rs.dockerClient.GetImageInfo(ctx, container.ImageID)
	if err != nil {
		updateInfo.Error = err
		return updateInfo, nil
	}

	// Extraer tag actual
	if len(localImage.RepoTags) > 0 {
		updateInfo.CurrentVersion = rs.extractTag(localImage.RepoTags[0])
	}

	// Verificar contra el registro remoto
	remoteInfo, _, err := rs.dockerClient.DistributionInspect(ctx, container.ImageName)
	if err != nil {
		// Si no se puede obtener info remota, asumimos que está actualizada
		// (evita falsos positivos para imágenes privadas o locales)
		updateInfo.Error = err
		return updateInfo, nil
	}

	// Comparar digests
	localDigest := localImage.ID
	remoteDigest := remoteInfo.Descriptor.Digest.String()

	if localDigest != remoteDigest {
		updateInfo.IsUpToDate = false
		// Intentar obtener la versión más reciente desde Docker Hub
		latestVersion, err := rs.getLatestVersionFromDockerHub(container.ImageName)
		if err != nil {
			// Si falla, usar el tag extraído de la imagen
			updateInfo.LatestVersion = rs.extractTag(container.ImageName)
		} else {
			updateInfo.LatestVersion = latestVersion
		}
	}

	return updateInfo, nil
}

// extractTag extrae el tag de una imagen completa
func (rs *RegistryStrategy) extractTag(image string) string {
	parts := strings.Split(image, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return "latest"
}

// DockerHubResponse estructura para la respuesta de Docker Hub API
type DockerHubResponse struct {
	Results []DockerHubTag `json:"results"`
}

// DockerHubTag estructura para un tag de Docker Hub
type DockerHubTag struct {
	Name         string    `json:"name"`
	LastUpdated  time.Time `json:"last_updated"`
}

// getLatestVersionFromDockerHub obtiene la última versión desde Docker Hub
func (rs *RegistryStrategy) getLatestVersionFromDockerHub(imageName string) (string, error) {
	// Extraer repositorio de la imagen (ej: "nginx:latest" -> "nginx")
	repo := rs.extractRepository(imageName)
	if repo == "" {
		return "", fmt.Errorf("unable to extract repository from image name")
	}

	// Construir URL de la API de Docker Hub
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/?page_size=100", repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tags from Docker Hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Docker Hub API returned status code: %d", resp.StatusCode)
	}

	var response DockerHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode Docker Hub response: %w", err)
	}

	// Ordenar tags por fecha de actualización (más reciente primero)
	sort.Slice(response.Results, func(i, j int) bool {
		return response.Results[i].LastUpdated.After(response.Results[j].LastUpdated)
	})

	// Buscar el tag más reciente que no sea "latest"
	for _, tag := range response.Results {
		if tag.Name != "latest" && !strings.Contains(tag.Name, "rc") && !strings.Contains(tag.Name, "beta") && !strings.Contains(tag.Name, "alpha") {
			return tag.Name, nil
		}
	}

	// Si no se encuentra una versión específica, devolver "latest"
	return "latest", nil
}

// extractRepository extrae el repositorio de una imagen completa
func (rs *RegistryStrategy) extractRepository(imageName string) string {
	// Remover el tag (ej: "nginx:latest" -> "nginx")
	parts := strings.Split(imageName, ":")
	if len(parts) == 0 {
		return ""
	}

	repo := parts[0]

	// Si no tiene namespace, agregar "library/" para imágenes oficiales
	if !strings.Contains(repo, "/") {
		repo = "library/" + repo
	}

	return repo
}
