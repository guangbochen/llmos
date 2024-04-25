package utils

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/elemental"
	"github.com/llmos-ai/llmos/pkg/utils/log"
)

const (
	elementalConfigDir  = "/tmp/elemental"
	elementalConfigFile = "config.yaml"
)

func SaveTemp(obj interface{}, prefix string, logger log.Logger, print bool) (string, error) {
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
	if logger.IsDebug() || print {
		pterm.Info.Print(string(bytes))
	}

	return tempFile.Name(), nil
}

func SaveElementalConfig(elemental *elemental.ElementalConfig, logger log.Logger) (string, string, error) {
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
		pterm.Info.Print(string(bytes))
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

func ReadLLMOSConfigFile(path string) (*config.LLMOSConfig, error) {
	var err error
	var data []byte

	// check if is valid path
	_, err = os.Stat(path)
	if err == nil {
		data, err = GetLocalLLMOSConfig(path)
		if err != nil {
			return nil, err
		}
	}

	// check if source is a valid url
	if strings.Contains(path, "http") {
		url, err := url.Parse(path)
		if err != nil {
			slog.Debug("invalid source url", "path", path)
			return nil, err
		}
		data, err = GetURLLLMOSConfig(url.String())
		if err != nil {
			return nil, fmt.Errorf("error reading LLMOS config file from url: %s", err.Error())
		}

	}

	return config.LoadLLMOSConfig(data)
}

func GetLocalLLMOSConfig(path string) ([]byte, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading local LLMOS config file: %s", err.Error())
	}
	return bytes, nil
}

func GetURLLLMOSConfig(url string) ([]byte, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	resp, err := retryClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the HTTP response is a success (2xx) or success-like code (3xx)
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, fmt.Errorf("url response status code is invalid: %d", resp.StatusCode)
}
