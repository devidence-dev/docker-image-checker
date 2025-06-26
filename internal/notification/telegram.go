package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pablopin/docker-image-checker/internal/model"
)

// TelegramNotifier implementa Observer para notificaciones de Telegram
type TelegramNotifier struct {
	botToken     string
	chatID       string
	templatePath string
	template     *template.Template
}

// NewTelegramNotifier crea un nuevo notificador de Telegram
func NewTelegramNotifier(botToken, chatID, templatePath string) (*TelegramNotifier, error) {
	notifier := &TelegramNotifier{
		botToken:     botToken,
		chatID:       chatID,
		templatePath: templatePath,
	}

	// Cargar plantilla
	if err := notifier.loadTemplate(); err != nil {
		return nil, fmt.Errorf("failed to load telegram template: %w", err)
	}

	return notifier, nil
}

// loadTemplate carga la plantilla de mensaje
func (tn *TelegramNotifier) loadTemplate() error {
	absPath, err := filepath.Abs(tn.templatePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	templateContent, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("telegram").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	tn.template = tmpl
	return nil
}

// Notify implementa la interfaz Observer
func (tn *TelegramNotifier) Notify(data *model.NotificationData) error {
	fmt.Printf("🔔 Preparando notificación de Telegram...\n")

	// Generar mensaje usando la plantilla
	message, err := tn.generateMessage(data)
	if err != nil {
		return fmt.Errorf("failed to generate message: %w", err)
	}

	fmt.Printf("📝 Mensaje generado (primeros 100 chars): %.100s...\n", message)

	// Enviar mensaje via API de Telegram
	err = tn.sendMessage(message)
	if err != nil {
		fmt.Printf("❌ Error enviando mensaje de Telegram: %v\n", err)
		return err
	}

	fmt.Printf("✅ Mensaje de Telegram enviado exitosamente\n")
	return nil
}

// generateMessage genera el mensaje usando la plantilla
func (tn *TelegramNotifier) generateMessage(data *model.NotificationData) (string, error) {
	var buf bytes.Buffer
	if err := tn.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

// sendMessage envía el mensaje via API de Telegram
func (tn *TelegramNotifier) sendMessage(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tn.botToken)

	payload := map[string]interface{}{
		"chat_id":    tn.chatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status code: %d", resp.StatusCode)
	}

	return nil
}
