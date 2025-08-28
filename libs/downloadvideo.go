package libs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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
	// args := []string{
	// 	"yt-dlp",
	// 	"--no-cache-dir",
	// 	"--no-mtime",
	// 	"--proxy", proxyUrl,
	// 	"-o", fmt.Sprintf("%s/%%(id)s.%%(ext)s", pathLoad),
	// 	"--merge-output-format=mp4/mkv",
	// 	"-f", "w",
	// 	url,
	// }

	args := []string{
		"yt-dlp",
		"--no-cache-dir",
		"--no-mtime",
		"--proxy", proxyUrl,
		"-o", fmt.Sprintf("%s/%%(id)s.%%(ext)s", pathLoad),
		"--merge-output-format", "mp4",
		"-f", "bestvideo[height>=1080]+bestaudio/bestvideo+bestaudio",
		"--concurrent-fragments", "4",
		"--embed-subs",
		"--recode-video", "mp4", // перекодировать в mp4, если нужно
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

	// === Используем универсальный обработчик с буферизацией ===
	sendProgress, stopProgress := NewThrottledHandler(callback, 3*time.Second)
	defer stopProgress() // гарантированно останавливаем тикер и отправляем остатки

	// Читаем stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stderr:", line)
			sendProgress(line)
		}
	}()

	// Читаем stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stdout:", line)
			sendProgress(line)
		}
	}()

	// Ждём завершения процесса
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

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
