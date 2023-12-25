package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// LogMessageType представляет тип сообщения в логе.
type LogMessageType string

const (
	ERROR   LogMessageType = "ERROR"
	WARNING LogMessageType = "WARNING"
	INFO    LogMessageType = "INFO"
)

// LogEntry представляет запись в логе.
type LogEntry struct {
	Type    LogMessageType
	Message string
}

// LogAnalyzer представляет анализатор логов.
type LogAnalyzer struct {
	LogFilePath   string
	DetailLevel   LogMessageType
	OutputFile    string
	Stats         map[LogMessageType]int
	TotalMessages int
	mutex         sync.Mutex
}

// NewLogAnalyzer создает новый экземпляр LogAnalyzer.
func NewLogAnalyzer(logFilePath, detailLevel, outputFile string) *LogAnalyzer {
	return &LogAnalyzer{
		LogFilePath: logFilePath,
		DetailLevel: LogMessageType(strings.ToUpper(detailLevel)),
		OutputFile:  outputFile,
		Stats:       make(map[LogMessageType]int),
	}
}

// Analyze анализирует лог-файл и собирает статистику.
func (la *LogAnalyzer) Analyze() error {
	file, err := os.Open(la.LogFilePath)
	if err != nil {
		return fmt.Errorf("ошибка при открытии файла лога: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("ошибка при чтении файла лога: %v", err)
		}

		entry := parseLogEntry(line)
		if entry.Type >= la.DetailLevel {
			la.updateStats(entry.Type)
		}
		la.TotalMessages++
	}

	return nil
}

// updateStats обновляет статистику на основе типа сообщения.
func (la *LogAnalyzer) updateStats(logType LogMessageType) {
	la.mutex.Lock()
	defer la.mutex.Unlock()
	la.Stats[logType]++
}

// PrintReport выводит отчет в консоль или файл.
func (la *LogAnalyzer) PrintReport() error {
	var output io.Writer

	if la.OutputFile != "" {
		file, err := os.Create(la.OutputFile)
		if err != nil {
			return fmt.Errorf("ошибка при создании файла отчета: %v", err)
		}
		defer file.Close()
		output = file
	} else {
		output = os.Stdout
	}

	fmt.Fprintln(output, "Статистика по сообщениям:")
	for logType, count := range la.Stats {
		fmt.Fprintf(output, "%s: %d\n", logType, count)
	}

	fmt.Fprintf(output, "Всего сообщений: %d\n", la.TotalMessages)

	return nil
}

// parseLogEntry парсит запись из лога.
func parseLogEntry(line string) LogEntry {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 2 {
		return LogEntry{
			Type:    LogMessageType(parts[0]),
			Message: parts[1],
		}
	}
	// Если запись не соответствует ожидаемому формату, считаем ее INFO.
	return LogEntry{
		Type:    INFO,
		Message: line,
	}
}

func main() {
	var logFilePath, detailLevel, outputFile string

	flag.StringVar(&logFilePath, "log", "", "Путь к лог-файлу")
	flag.StringVar(&detailLevel, "level", "INFO", "Уровень детализации анализа (ERROR, WARNING, INFO)")
	flag.StringVar(&outputFile, "output", "", "Путь к файлу отчета")
	flag.Parse()

	// Если не указаны флаги, используем переменные окружения
	if logFilePath == "" {
		logFilePath = os.Getenv("LOG_FILE_PATH")
	}
	if detailLevel == "" {
		detailLevel = os.Getenv("DETAIL_LEVEL")
	}
	if outputFile == "" {
		outputFile = os.Getenv("OUTPUT_FILE")
	}

	// Создаем экземпляр анализатора логов
	logAnalyzer := NewLogAnalyzer(logFilePath, detailLevel, outputFile)

	// Анализируем логи
	err := logAnalyzer.Analyze()
	if err != nil {
		fmt.Printf("Ошибка при анализе логов: %v\n", err)
		os.Exit(1)
	}

	// Выводим отчет
	err = logAnalyzer.PrintReport()
	if err != nil {
		fmt.Printf("Ошибка при выводе отчета: %v\n", err)
		os.Exit(1)
	}
}
