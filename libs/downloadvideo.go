package libs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"time"
)

func DownloadVideo(url string, callback func(string)) error {
	pathLoad := os.Getenv("PATH_TO_LOAD_VIDEO")
	proxyUrl := os.Getenv("PROXY_URL")

	if pathLoad == "" {
		return fmt.Errorf("PATH_TO_LOAD_VIDEO is not set")
	}
	if proxyUrl == "" {
		return fmt.Errorf("PROXY_URL is not set")
	}

	// Аргументы команды
	args := []string{
		"yt-dlp",
		"--no-cache-dir",
		"--no-mtime",
		"--proxy", proxyUrl,
		"-o", fmt.Sprintf("%s/%%(id)s.%%(ext)s", pathLoad),
		"--merge-output-format=mp4/mkv",
		"-f", "w",
		url,
	}

	cmd := exec.Command(args[0], args[1:]...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// === Буферизация прогресса с таймером ===
	var buffer []string
	var mu sync.Mutex
	var lastSent string

	// Флаг, чтобы знать, когда остановиться
	done := make(chan bool)
	ticker := time.NewTicker(3 * time.Second) // интервал обновления

	// Функция отправки последнего сообщения из буфера
	flush := func() {
		mu.Lock()
		defer mu.Unlock()
		if len(buffer) > 0 {
			latest := buffer[len(buffer)-1]
			if latest != lastSent && callback != nil {
				callback(latest)
				lastSent = latest
			}
			buffer = buffer[:0] // очищаем
		}
	}

	// Горутина для периодической отправки
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				flush()
			case <-done:
				flush() // финальная отправка
				return
			}
		}
	}()

	// Читаем stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stderr:", line) // лог в консоль

			mu.Lock()
			buffer = append(buffer, line)
			mu.Unlock()
		}
	}()

	// Читаем stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stdout:", line)

			mu.Lock()
			buffer = append(buffer, line)
			mu.Unlock()
		}
	}()

	// Ждём завершения команды
	if err := cmd.Wait(); err != nil {
		close(done)
		return fmt.Errorf("command failed: %w", err)
	}

	close(done)                        // останавливаем тикер и отправляем последнее сообщение
	time.Sleep(100 * time.Millisecond) // даём тикеру шанс обработать финальный flush

	return nil
}

func GetCode(url string) string {
	// Поддерживает: ?v=..., /v/..., /embed/..., /shorts/..., youtu.be/...
	re := regexp.MustCompile(`(?:v=|\/v\/|\/embed\/|\/shorts\/|youtu\.be\/)([A-Za-z0-9_-]{11})`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1] // возвращаем первую захваченную группу
	}
	return ""
}
