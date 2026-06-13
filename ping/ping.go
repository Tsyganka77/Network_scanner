package ping

import (
	"context"
	"net"                                  //Для проверки типа DNSError
	pkgerrors "network-scanner/pkg/errors" //Алиас для пакета ошибок
	"network-scanner/utils"                //Импорт локального пакета
	"os/exec"                              //Для запуска внешних команд
	"strconv"                              //Преобразование строки в число
	"strings"                              //Работа со строками
	"time"                                 //Работа со временем
)

type PingResult struct {
	Host      string        //IP или домен
	Available bool          //Доступен ли
	Latency   time.Duration //Время ответа
	Error     error         //Ошибка если есть
}

// pingHost - проверка доступности хоста
func PingHost(ctx context.Context, host string, timeout time.Duration) PingResult {
	result := PingResult{
		Host:      host,
		Available: false,
		Latency:   0,
		Error:     nil,
	}

	select {
	case <-ctx.Done():
		//Возвращаем кастомную ошибку
		result.Error = pkgerrors.NewCanceledError(host)
		return result
	default:
	}

	//Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() //Освобождаем ресурсыт

	//Засекаем время пинг
	startTime := time.Now()
	//Выполняем команду ping
	// -c 1 - один пакет(Linux/Mac)
	// -n 1 - один пакет(Winda)
	var cmd *exec.Cmd
	//Определяем ОС
	if utils.IsWindows() {
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", host)
	} else {
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", host)
	}

	//Выполняем команду и получаем вывод
	output, err := cmd.CombinedOutput()
	//Считаем время ответа
	result.Latency = time.Since(startTime)

	//Проверка таймаута
	if ctx.Err() == context.DeadlineExceeded {
		//Кастомная ошибка таймаута
		result.Error = pkgerrors.NewTimeoutError(host, timeout)
		result.Available = false
		return result
	}

	//проверка отмены пользователем
	if ctx.Err() == context.Canceled {
		//Кастомная ошибка отмены
		result.Error = pkgerrors.NewCanceledError(host)
		result.Available = false
		return result
	}

	//Проверяем ошибку
	if err != nil {
		//Проверяем тип ошибки и создаем кастомную
		if _, ok := err.(*net.DNSError); ok {
			result.Error = pkgerrors.NewDNSError(host, err)
		} else {
			result.Error = pkgerrors.NewHostUnreachableError(host, err)
		}
		result.Available = false
		return result
	}

	//Хост доступен
	result.Available = true
	//Пытаемся извлечь время ответа из вывода ping
	result.Latency = ParsePingLatency(string(output))

	return result
}

// ParsePingLatency - парсинг времени ответа из вывода ping
func ParsePingLatency(output string) time.Duration {
	//Ищем строку с time= (Linux/Mac)
	if strings.Contains(output, "time=") {
		//Находим позицию time=
		idx := strings.Index(output, "time=")
		if idx == -1 {
			return 0
		}
		//Берем часть после time=
		afterTime := output[idx+5:]
		//Находим пробел или конец
		endIdx := strings.IndexAny(afterTime, "\n")
		if endIdx == -1 {
			endIdx = len(afterTime)
		}
		//Извлекаем число
		timeStr := afterTime[:endIdx]
		//Удаляем "ms" если есть
		timeStr = strings.TrimSuffix(timeStr, "ms")
		timeStr = strings.TrimSpace(timeStr)

		val, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return 0
		}
		return time.Duration(val * float64(time.Millisecond))
	}

	if strings.Contains(output, "time<") {
		return 1 * time.Millisecond
	}

	return 0
}
