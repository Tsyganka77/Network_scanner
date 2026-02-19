//Network scanner на ЯП Go

package main

import(
	"fmt"    //ввод вывод
	"os"    //работа с os
	"os/exec"    //выполнение внешних команд
	"strings"    //работа со строками
	"time"    //работа со временем
)

//Константы

const(
	//Таймаут для Ping(сколько ждать ответа)
	PingTimeout = 2*time.Second    //2 секунды
	//Сообщение об успешном пинге
	SuccessMsg = "Хост доступен"
	//Сообщение об ошибке
	ErrorMsg = "Хост недоступен"
)

//Объявление структуры(типа данных). Типы в виде структур, которые имеют свойства и поведение.
//Нет наследования, сборка сложых типов через композицию. Объявление типов с помощью ключевого слова type.
//PingResult - результат проверки хоста
type PingResult struct {
	Host string	//IP или домен
	Available bool	//Доступен ли
	Latency time.Duration	//Время ответа
	Error error	//Ошибка если есть
}

//main - точка входа
func main(){
	arguments:= os.Args	//os.Args в Go — это срез строк ([]string), содержащий аргументы
				//командной строки, переданные при запуске программы
	if len(arguments)<2{
		fmt.Println("Network Scanner v0.1")
		fmt.Println(" %s <хост>\n", arguments[0])
		fmt.Println("Пример:")
		fmt.Println(" %s 8.8.8.8\n", arguments[0])
		fmt.Println(" %s google.com\n", arguments[0])
		os.Exit(1)//завершаем с кодом ошибки(передача 1 сообщает об ошибке)
	}

host := arguments[1]	//берем первый аргумент

fmt.Printf("Сканирование хоста: %s\n", host)
fmt.Println(strings.Repeat("-",40))
//засекаем время начала
startTime := time.Now()
//выполняем пинг
result := pingHost(host)
//считаем общее время
totalTime := time.Since(startTime)
//выводим результат
printResult(result, totalTime)
}

//pingHost - проверка доступности хоста
func pingHost(host string) PingResult{
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
	if isWindows(){
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
	result.Latency = parsePingLatency(string(output))

	return result
}

//parsePingLatency - парсинг времени ответа из вывода ping
func parsePingLatency(output string) time.Duration{
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
		//Парсим число и возвращаем примерное значение
		fmt.Printf("Время ответа: %s ms\n", timeStr)
	}
	//Для винды формат другой: time<1ms or time=23ms
	if strings.Contains(output, "time<"){
		return 1 * time.Millisecond
	}
	return 0
}

//printResult - вывод результатов

func printResult(result PingResult, totalTime time.Duration){
	fmt.Println()

	if result.Available{
		//Успех
		fmt.Printf(" %s %s\n", SuccessMsg, result.Host)
		fmt.Printf(" Время ping: %v\n", result.Latency)
		fmt.Printf(" Общее время: %v\n", totalTime)
	}else{
		//Ошибка
		fmt.Printf("%s %s\n", ErrorMsg, result.Host)
		if result.Error != nil {
			fmt.Printf(" Ошибка: %v\n", result.Error)
		}
	}
	fmt.Println(strings.Repeat("-", 40))
}

//isWindows - проверка ОС

func isWindows() bool {	//Простая проверка по переменной окружения
	return os.Getenv("OS") == "Windows_NT"
}
