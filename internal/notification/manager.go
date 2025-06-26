package notification

import "github.com/pablopin/docker-image-checker/internal/model"

// Observer define la interfaz para observadores de notificaciones
type Observer interface {
	Notify(data *model.NotificationData) error
}

// Subject define la interfaz para el sujeto observado
type Subject interface {
	Subscribe(observer Observer)
	Unsubscribe(observer Observer)
	NotifyAll(data *model.NotificationData) error
}

// NotificationManager implementa el patrón Observer para notificaciones
type NotificationManager struct {
	observers []Observer
}

// NewNotificationManager crea un nuevo manager de notificaciones
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		observers: make([]Observer, 0),
	}
}

// Subscribe añade un observer
func (nm *NotificationManager) Subscribe(observer Observer) {
	nm.observers = append(nm.observers, observer)
}

// Unsubscribe remueve un observer
func (nm *NotificationManager) Unsubscribe(observer Observer) {
	for i, obs := range nm.observers {
		if obs == observer {
			nm.observers = append(nm.observers[:i], nm.observers[i+1:]...)
			break
		}
	}
}

// NotifyAll notifica a todos los observers
func (nm *NotificationManager) NotifyAll(data *model.NotificationData) error {
	var lastError error
	
	for _, observer := range nm.observers {
		if err := observer.Notify(data); err != nil {
			lastError = err
			// Log error but continue with other observers
		}
	}
	
	return lastError
}
