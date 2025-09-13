package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ImageDownloader handles image downloading
type ImageDownloader struct {
	imagesDir string
	client    *http.Client
}

// NewImageDownloader creates a new image downloader
func NewImageDownloader(imagesDir string) *ImageDownloader {
	return &ImageDownloader{
		imagesDir: imagesDir,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DownloadImage downloads an image and returns the local path
func (d *ImageDownloader) DownloadImage(url, productID string) (string, error) {
	if url == "" || strings.HasPrefix(url, "data:") {
		return "", fmt.Errorf("invalid URL: %s", url)
	}

	// Нормализуем URL
	normalizedURL := d.normalizeURL(url)

	// Определяем расширение файла
	ext := d.extractExtension(normalizedURL)

	// Создаем безопасное имя файла
	safeProductID := d.sanitizeFilename(productID)
	filename := fmt.Sprintf("%s.%s", safeProductID, ext)
	localPath := filepath.Join(d.imagesDir, filename)

	// Проверяем, существует ли файл
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil // Файл уже существует
	}

	// Создаем HTTP запрос
	req, err := http.NewRequest("GET", normalizedURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Referer", "https://www.pricerunner.com/")

	// Выполняем запрос
	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Создаем файл
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Копируем содержимое
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(localPath) // Удаляем частично загруженный файл
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return localPath, nil
}

// normalizeURL нормализует URL изображения
func (d *ImageDownloader) normalizeURL(url string) string {
	if strings.HasPrefix(url, "//") {
		return "https:" + url
	}
	if strings.HasPrefix(url, "/") {
		return "https://www.pricerunner.com" + url
	}
	return url
}

// extractExtension извлекает расширение файла из URL
func (d *ImageDownloader) extractExtension(url string) string {
	re := regexp.MustCompile(`\.([a-zA-Z0-9]+)(?:\?|$)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		ext := strings.ToLower(matches[1])
		// Проверяем, что это действительно расширение изображения
		validExts := map[string]bool{
			"jpg": true, "jpeg": true, "png": true, "gif": true,
			"webp": true, "avif": true, "svg": true, "bmp": true,
		}
		if validExts[ext] {
			return ext
		}
	}
	return "jpg" // По умолчанию
}

// sanitizeFilename создает безопасное имя файла
func (d *ImageDownloader) sanitizeFilename(filename string) string {
	// Заменяем недопустимые символы на подчеркивания
	re := regexp.MustCompile(`[^\w\-]`)
	return re.ReplaceAllString(filename, "_")
}
