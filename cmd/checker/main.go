package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pablopin/docker-image-checker/internal/config"
	"github.com/pablopin/docker-image-checker/internal/docker"
	"github.com/pablopin/docker-image-checker/internal/model"
	"github.com/pablopin/docker-image-checker/internal/notification"
	"github.com/robfig/cron/v3"
)

const (
	ColorGreen  = "\033[92m"
	ColorYellow = "\033[93m"
	ColorRed    = "\033[91m"
	ColorBlue   = "\033[94m"
	ColorReset  = "\033[0m"
)

func main() {
	var (
		configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
		daemon     = flag.Bool("daemon", false, "Run as daemon")
		once       = flag.Bool("once", false, "Run check once and exit")
	)
	flag.Parse()

	// Cargar configuración
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("%sError loading configuration: %v%s", ColorRed, err, ColorReset)
	}

	// Crear cliente Docker
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatalf("%sError creating Docker client: %v%s", ColorRed, err, ColorReset)
	}
	defer dockerClient.Close()

	// Crear checker
	checker := docker.NewChecker(dockerClient)

	// Configurar sistema de notificaciones
	notificationManager := notification.NewNotificationManager()

	if cfg.Notifications.Telegram.Enabled {
		telegramNotifier, err := notification.NewTelegramNotifier(
			cfg.TelegramBotToken,
			cfg.TelegramChatID,
			cfg.Notifications.Telegram.TemplateFile,
		)
		if err != nil {
			log.Fatalf("%sError creating Telegram notifier: %v%s", ColorRed, err, ColorReset)
		}
		notificationManager.Subscribe(telegramNotifier)
	}

	// Crear aplicación
	app := &App{
		checker:  checker,
		notifier: notificationManager,
		config:   cfg,
	}

	if *once {
		// Ejecutar una sola vez
		if err := app.runOnce(); err != nil {
			log.Fatalf("%sError running check: %v%s", ColorRed, err, ColorReset)
		}
		return
	}

	if *daemon {
		// Ejecutar como daemon
		app.runDaemon()
	} else {
		// Ejecutar una sola vez por defecto
		if err := app.runOnce(); err != nil {
			log.Fatalf("%sError running check: %v%s", ColorRed, err, ColorReset)
		}
	}
}

// App encapsula la lógica de la aplicación
type App struct {
	checker  *docker.Checker
	notifier *notification.NotificationManager
	config   *config.Config
}

// runOnce ejecuta la verificación una sola vez
func (a *App) runOnce() error {
	fmt.Printf("%s--- Iniciando verificación de imágenes Docker ---%s\n", ColorBlue, ColorReset)

	ctx := context.Background()
	report, err := a.checker.CheckAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to check containers: %w", err)
	}

	// Añadir información del hostname
	hostname, _ := os.Hostname()
	report.Hostname = hostname
	report.Timestamp = time.Now()

	// Mostrar resultados en consola
	a.printReport(report)

	// Enviar notificaciones si hay actualizaciones o errores
	if len(report.Available) > 0 || len(report.Failed) > 0 {
		fmt.Printf("📢 Enviando notificaciones (Actualizaciones: %d, Errores: %d)...\n", len(report.Available), len(report.Failed))

		notificationData := &model.NotificationData{
			Report:   report,
			Hostname: hostname,
		}

		if err := a.notifier.NotifyAll(notificationData); err != nil {
			fmt.Printf("%sWarning: Failed to send notifications: %v%s\n", ColorYellow, err, ColorReset)
		} else {
			fmt.Printf("✅ Notificaciones enviadas exitosamente\n")
		}
	} else {
		fmt.Printf("ℹ️  No hay actualizaciones ni errores, no se envían notificaciones\n")
	}

	fmt.Printf("%s--- Verificación completada ---%s\n", ColorBlue, ColorReset)
	return nil
}

// runDaemon ejecuta la verificación de forma continua
func (a *App) runDaemon() {
	c := cron.New()

	_, err := c.AddFunc(a.config.Checker.Schedule, func() {
		if err := a.runOnce(); err != nil {
			log.Printf("%sError running check: %v%s", ColorRed, err, ColorReset)
		}
	})
	if err != nil {
		log.Fatalf("%sInvalid cron schedule configuration: %v%s", ColorRed, err, ColorReset)
	}

	fmt.Printf("%s--- Iniciando modo daemon (schedule: %s) ---%s\n", ColorBlue, a.config.Checker.Schedule, ColorReset)

	// Canal para manejar señales de interrupción
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	c.Start()
	defer c.Stop()

	// Ejecutar primera verificación inmediatamente
	if err := a.runOnce(); err != nil {
		log.Printf("%sError in initial check: %v%s", ColorRed, err, ColorReset)
	}

	// Esperar señal de interrupción
	<-sigChan
	fmt.Printf("%s--- Deteniendo daemon ---%s\n", ColorBlue, ColorReset)
}

// printReport muestra el reporte en consola con colores
func (a *App) printReport(report *model.CheckReport) {
	fmt.Printf("\n📊 Resumen:\n")
	fmt.Printf("   - 🖥️  Host: %s\n", report.Hostname)
	fmt.Printf("   - 🛟  Contenedores con actualizaciones disponibles: %s%d%s\n", ColorYellow, len(report.Available), ColorReset)
	fmt.Printf("   - ✅  Contenedores verificados: %d\n", report.Total)
	fmt.Printf("   - ❌  Fallidos: %s%d%s\n", ColorRed, len(report.Failed), ColorReset)

	if len(report.Available) > 0 {
		fmt.Printf("\n📦 Actualizaciones disponibles:\n")
		for _, update := range report.Available {
			fmt.Printf("   - 🔄 %s%s%s (%s)\n", ColorYellow, update.Container.Name, ColorReset, update.Container.ImageName)
			fmt.Printf("     • Versión actual: %s\n", update.CurrentVersion)
			fmt.Printf("     • Nueva versión: %s\n", update.LatestVersion)
		}
	}

	if len(report.Failed) > 0 {
		fmt.Printf("\n🚫 Fallos en:\n")
		for _, failed := range report.Failed {
			fmt.Printf("   - %s%s%s (%s) ❌\n", ColorRed, failed.Container.Name, ColorReset, failed.Container.ImageName)
			if failed.Error != nil {
				fmt.Printf("     Error: %v\n", failed.Error)
			}
		}
	}

	if len(report.UpToDate) > 0 {
		fmt.Printf("\n✅ Actualizados (%d):\n", len(report.UpToDate))
		for _, upToDate := range report.UpToDate {
			fmt.Printf("   - %s%s%s (%s)\n", ColorGreen, upToDate.Container.Name, ColorReset, upToDate.Container.ImageName)
		}
	}
}
