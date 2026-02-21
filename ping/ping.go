package ping

import (
	"os/exec"    //Для запуска внешних команд
	"strconv"    //Преобразование строки в число
        "strings"    //Работа со строками
	"time"    //Работа со временем
	"network-scanner/utils"    //Импорт локального пакета
)

type PingResult struct {
	Host string	//IP или домен
	Available bool	//Доступен ли
	Latency time.Duration	//Время ответа
	Error error	//Ошибка если есть
}

//pingHost - проверка доступности хоста
func PingHost(host string) PingResult{
	result := PingResult{
		Host: host,
		Available: false,
		Latency: 0,
		Error: nil,
	}

	//Засекаем время пинг
	startTime := time.Now()
	//Выполняем команду ping
	// -c 1 - один пакет(Linux/Mac)
	// -n 1 - один пакет(Winda)
	var cmd *exec.Cmd
	//Определяем ОС
	if utils.IsWindows(){
		cmd = exec.Command("ping","-n","1",host)
	}else{
		cmd = exec.Command("ping","-c","1",host)
	}

	//Выполняем команду и получаем вывод
	output, err := cmd.CombinedOutput()
	//Считаем время ответа
	result.Latency = time.Since(startTime)
	//Проверяем ошибку
	if err != nil{
		result.Error = err
		result.Available = false
		return result
	}

	//Хост доступен
	result.Available = true
	//Пытаемся извлечь время ответа из вывода ping
	result.Latency = ParsePingLatency(string(output))

	return result
}

//parsePingLatency - парсинг времени ответа из вывода ping
func ParsePingLatency(output string) time.Duration{
	//Ищем строку с time= (Linux/Mac)
	if strings.Contains(output, "time="){
		//Находим позицию time=
		idx := strings.Index(output, "time=")
		if idx == -1{
			return 0
		}
		//Берем часть после time=
		afterTime := output[idx+5:]
		//Находим пробел или конец
		endIdx := strings.IndexAny(afterTime, "\n")
		if endIdx == -1{
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

	if strings.Contains(output, "time<"){
		return 1 * time.Millisecond
	}
	
	return 0
}
