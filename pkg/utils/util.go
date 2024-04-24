package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/log"
)

const (
	elementalConfigDir  = "/tmp/elemental"
	elementalConfigFile = "config.yaml"
)

func SaveTemp(obj interface{}, prefix string, logger log.Logger) (string, error) {
	tempFile, err := os.CreateTemp("/tmp", fmt.Sprintf("%s.", prefix))
	if err != nil {
		return "", err
	}

	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	if _, err = tempFile.Write(bytes); err != nil {
		return "", err
	}
	if err = tempFile.Close(); err != nil {
		return "", err
	}

	logger.Info("Saved file successfully", "fileName", tempFile.Name())
	if logger.IsDebug() {
		fmt.Printf("config file:\n%v\n", string(bytes))
	}

	return tempFile.Name(), nil
}

func SaveElementalConfig(elemental *config.ElementalConfig, logger log.Logger) (string, string, error) {
	err := os.MkdirAll(elementalConfigDir, os.ModePerm)
	if err != nil {
		return "", "", err
	}

	bytes, err := yaml.Marshal(elemental)
	if err != nil {
		return "", "", err
	}

	file := filepath.Join(elementalConfigDir, elementalConfigFile)
	err = os.WriteFile(file, bytes, os.ModePerm)
	if err != nil {
		return "", "", err
	}

	logger.Info("Saved elemental config file successfully", "fileName", file)
	if logger.IsDebug() {
		fmt.Printf("config file:\n%v\n", string(bytes))
	}

	return elementalConfigDir, file, nil
}

func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

func SetEnv(env []string) {
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) >= 2 {
			os.Setenv(pair[0], pair[1])
		}
	}
}

func IsRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	if env := os.Getenv("KUBERNETES_SERVICE_HOST"); env != "" {
		return true
	}
	return false
}

func IsValidPathOrURL(path string) error {
	var err error

	// check if is valid path
	if _, err = os.Stat(path); err == nil {
		return nil
	}

	// check if source is a valid url
	if strings.Contains(path, "http") {
		url, err := url.Parse(path)
		if err != nil {
			slog.Debug("invalid source url: %s", path)
			return err
		}

		client := http.Client{
			Timeout: 1 * time.Second,
		}
		resp, err := client.Get(url.String())
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check if the HTTP response is a success (2xx) or success-like code (3xx)
		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
			return nil
		}
		return fmt.Errorf("url response status code is invalid: %d", resp.StatusCode)
	}

	return err
}
