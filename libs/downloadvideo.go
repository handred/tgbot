package libs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

	// Лучше передавать аргументы отдельно, чтобы избежать инъекций
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

	// Перенаправляем stderr (основной источник вывода yt-dlp)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Опционально: можно также читать stdout, если нужно
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Запускаем команду
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Читаем stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines) // читаем по строкам
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stderr:", line) // для отладки
			callback(line)
		}
	}()

	// Читаем stdout (может быть пустым, но на всякий случай)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stdout:", line)
			callback(line)
		}
	}()

	// Ждём завершения команды
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
