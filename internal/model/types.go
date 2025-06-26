package model

import "time"

// Container representa un contenedor Docker
type Container struct {
	ID        string
	Name      string
	ImageName string
	ImageID   string
	Status    string
}

// ImageInfo contiene información sobre una imagen
type ImageInfo struct {
	Name          string
	LocalDigest   string
	RemoteDigest  string
	CurrentTag    string
	LatestTag     string
	IsUpToDate    bool
	Error         error
}

// UpdateInfo representa información de actualización para un contenedor
type UpdateInfo struct {
	Container      Container
	CurrentVersion string
	LatestVersion  string
	IsUpToDate     bool
	Error          error
}

// CheckReport representa el reporte completo de verificación
type CheckReport struct {
	Hostname  string
	Timestamp time.Time
	Total     int
	Available []UpdateInfo
	Failed    []UpdateInfo
	UpToDate  []UpdateInfo
}

// NotificationData representa los datos para las notificaciones
type NotificationData struct {
	Report   *CheckReport
	Hostname string
}
