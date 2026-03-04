// Package hdfs предоставляет клиент для работы с HDFS через WebHDFS API.
package hdfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/zYoma/yandex_kafka7/internal/logger"
)

// WebHDFSWriter реализует запись данных в HDFS через WebHDFS HTTP API.
type WebHDFSWriter struct {
	baseURL   string
	user      string
	basePath  string
	batchSize int
}

// NewWebHDFSWriter создает новый клиент WebHDFS с указанной конфигурацией.
func NewWebHDFSWriter(cfg *HDFSWriterConfig) (*WebHDFSWriter, error) {
	addresses := strings.Split(cfg.Addresses, ",")[0]
	port := cfg.Port
	if port == "" {
		port = "14000"
	}

	var baseURL string
	if strings.HasPrefix(addresses, "http") {
		baseURL = addresses
	} else {
		baseURL = "http://" + addresses + ":" + port
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	return &WebHDFSWriter{
		baseURL:   baseURL,
		user:      "hadoop",
		basePath:  cfg.BasePath,
		batchSize: batchSize,
	}, nil
}

// buildURL строит полный URL для WebHDFS API запроса с указанными параметрами.
func (w *WebHDFSWriter) buildURL(hdfsPath string, params map[string]string) string {
	u, _ := url.Parse(w.baseURL + "/webhdfs/v1" + hdfsPath)
	q := u.Query()
	q.Set("user.name", w.user)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// mkdir создает директорию в HDFS через WebHDFS API.
func (w *WebHDFSWriter) mkdir(hdfsPath string) error {
	params := map[string]string{
		"op":         "MKDIRS",
		"permission": "0777",
	}
	fullURL := w.buildURL(hdfsPath, params)

	req, err := http.NewRequest("PUT", fullURL, nil)
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при создании директории: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка при создании директории: %s", string(body))
	}

	return nil
}

// write записывает данные в файл в HDFS через WebHDFS API с перенаправлением.
func (w *WebHDFSWriter) write(hdfsPath string, content string) (*http.Response, error) {
	params := map[string]string{
		"op":        "CREATE",
		"overwrite": "true",
	}

	redirectURL := w.buildURL(hdfsPath, params)

	req, err := http.NewRequest("PUT", redirectURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании файла: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTemporaryRedirect || resp.StatusCode == http.StatusSeeOther {
		location := resp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("нет Location header в redirect")
		}

		req2, err := http.NewRequest("PUT", location, bytes.NewBufferString(content))
		if err != nil {
			return nil, fmt.Errorf("ошибка при создании запроса записи: %w", err)
		}
		req2.Header.Set("Content-Type", "application/octet-stream")

		return http.DefaultClient.Do(req2)
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ожидался redirect или 201, получен статус %d: %s", resp.StatusCode, string(body))
	}

	req2, err := http.NewRequest("PUT", redirectURL+"&data=true", bytes.NewBufferString(content))
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса записи: %w", err)
	}
	req2.Header.Set("Content-Type", "application/octet-stream")

	return http.DefaultClient.Do(req2)
}

// WriteRequestData записывает данные запроса в файл в HDFS через WebHDFS API.
func (w *WebHDFSWriter) WriteRequestData(ctx context.Context, data []byte, filename string) error {
	logger.Get().Sugar().Infof("WebHDFS: Запись данных запроса: %s", filename)

	requestsDir := path.Join(w.basePath, "requests")
	if err := w.mkdir(requestsDir); err != nil {
		return fmt.Errorf("не удалось создать директорию requests: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405_000")
	fullFilename := path.Join(requestsDir, fmt.Sprintf("%s_%s.json", filename, timestamp))

	resp, err := w.write(fullFilename, string(data))
	if err != nil {
		return fmt.Errorf("ошибка при записи в файл %s: %w", fullFilename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка при записи данных: статус %d, ответ: %s", resp.StatusCode, string(body))
	}

	logger.Get().Sugar().Infof("Успешно записано %d байт через WebHDFS: %s", len(data), fullFilename)
	return nil
}

// Close метод-заглушка для WebHDFSWriter, так как HTTP клиент не требует закрытия.
func (w *WebHDFSWriter) Close() error {
	return nil
}

func (w *WebHDFSWriter) ReadRecommendations(ctx context.Context) ([]byte, error) {
	recPath := path.Join(w.basePath, "recommendations/part-00000")
	params := map[string]string{"op": "OPEN"}
	fullURL := w.buildURL(recPath, params)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to read recommendations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to read recommendations: %d, %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (w *WebHDFSWriter) ListFiles(ctx context.Context, dirPath string) ([]string, error) {
	params := map[string]string{"op": "LISTSTATUS"}
	fullURL := w.buildURL(dirPath, params)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list files: %d, %s", resp.StatusCode, string(body))
	}

	var result struct {
		FileStatuses struct {
			FileStatus []struct {
				PathSuffix string `json:"pathSuffix"`
				Type       string `json:"type"`
			} `json:"FileStatus"`
		} `json:"FileStatuses"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var files []string
	for _, file := range result.FileStatuses.FileStatus {
		if file.Type == "FILE" {
			files = append(files, path.Join(dirPath, file.PathSuffix))
		}
	}

	return files, nil
}

func (w *WebHDFSWriter) ReadFile(ctx context.Context, filePath string) ([]byte, error) {
	params := map[string]string{"op": "OPEN"}
	fullURL := w.buildURL(filePath, params)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to read file: %d, %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
