// errors.Is() - проверка типа ошибок: "является лм эта ошибка равной указанной?". Данный метод надежный, ему не важен регистр,
// заглядывает внутрь.
// errors.As() - извлечение деталей: ищет ошибку указанного типа в цепочке, есди находит, то записывает в переменную,
// возвращает true если нашла
package main

import (
	"context"
	"errors"                               //Стандартный пакет для ls/As
	"fmt"                                  //ввод вывод
	"network-scanner/batch"                //подключаем пакет batch
	"network-scanner/ping"                 //подключаем пакет ping
	pkgerrors "network-scanner/pkg/errors" //Пакет кастомных ошибок
	"os"                                   //работа с os
	"os/signal"
	"strings" //работа со строками
	"syscall"
	"time" //работа со временем
)

//Константы

const (
	//Сообщение об успешном пинге
	SuccessMsg = "Хост доступен"
	//Сообщение об ошибке
	ErrorMsg = "Хост недоступен"
	//Максимум параллельных пингов
	MaxConcurrent  = 10
	DefaultTimeout = 2 * time.Second
)

// main - точка входа
func main() {
	arguments := os.Args //os.Args в Go — это срез строк ([]string), содержащий аргументы
	//командной строки, переданные при запуске программы
	if len(arguments) < 2 {
		printUsage(arguments[0])
		os.Exit(1)
	}

	//Корневой контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() //Освобождаем ресурсы при выходе

	//создаем канал для сигналов
	sigChan := make(chan os.Signal, 1)

	//подписываемся на SIGINT (Ctrl+C) и SIGTERM
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//горутина для обработки сигналов
	go func() {
		sig := <-sigChan
		fmt.Printf("\n\n Получен сигнал: %v\n", sig)
		fmt.Println("Завершение работы")
		cancel() //отмена контекста
	}()

	//Режим 1: 1 хост
	if len(arguments) == 2 {
		runSingleHost(ctx, arguments[1])
		return
	}

	//Режим 2: несколько хостов
	if len(arguments) > 2 {
		runMultipleHosts(ctx, arguments[1:])
		return
	}
}

// runSingleHost - сканирование одного хоста
func runSingleHost(ctx context.Context, host string) {
	fmt.Printf("Сканирование хоста: %s\n", host)
	fmt.Println(strings.Repeat("-", 40))
	//засекаем время начала
	startTime := time.Now()
	//выполняем пинг из другого пакета, передаем контекст и таймаут
	result := ping.PingHost(ctx, host, DefaultTimeout)
	//считаем общее время
	totalTime := time.Since(startTime)

	//проверка не было ли отмены
	if ctx.Err() != nil {
		fmt.Println("Сканирование прервано пользователем")
		return
	}

	//выводим результат
	PrintResult(result, totalTime)
}

// runMultipleHosts - сканирование нескольких хостов параллельно
func runMultipleHosts(ctx context.Context, hosts []string) {
	fmt.Printf("Сканирование %d хостов (параллельно: %d)\n", len(hosts), MaxConcurrent)
	fmt.Println(strings.Repeat("-", 40))

	startTime := time.Now()
	//Запускаем параллельное сканирование через batch
	results := batch.ScanHosts(ctx, hosts, MaxConcurrent, DefaultTimeout)

	totalTime := time.Since(startTime)

	//проверка не было ли отмены
	if ctx.Err() != nil {
		fmt.Println("Сканирование прервано пользователем")
		fmt.Printf("Частичные результаты: %d хостов обработано\n", len(results))
		return
	}

	//выводим результат
	PrintMultipleResults(results, totalTime)
}

// PrintResult - вывод результатов
func PrintResult(result ping.PingResult, totalTime time.Duration) {
	fmt.Println()

	if result.Available {
		//Успех
		fmt.Printf(" %s %s\n", SuccessMsg, result.Host)
		fmt.Printf(" Время ответа: %v\n", result.Latency)
		fmt.Printf(" Общее время: %v\n", totalTime)
	} else {
		//Ошибка
		fmt.Printf("%s %s\n", ErrorMsg, result.Host)
		if result.Error != nil {
			//errors.Is() - для проверки типа
			if errors.Is(result.Error, pkgerrors.ErrTimeout) {
				fmt.Printf("Таймаут: %v\n", result.Error)
			} else if errors.Is(result.Error, pkgerrors.ErrCanceled) {
				fmt.Printf("Отменено пользователем\n")
			} else if errors.Is(result.Error, pkgerrors.ErrDNSFailed) {
				fmt.Printf("Ошибка DNS: %v\n", result.Error)
			} else {
				fmt.Printf("Ошибка: %v\n", result.Error)
			}
			//errors.As() - для извлечения деталей
			var pingErr *pkgerrors.PingError
			if errors.As(result.Error, &pingErr) {
				fmt.Printf("Хост: %s\n", pingErr.Host)
			}
		}
	}
	fmt.Println(strings.Repeat("-", 40))
}

// PrintMultipleResults - вывод результатов нескольких хостов
func PrintMultipleResults(results []batch.ScanResult, totalTime time.Duration) {
	fmt.Println()

	available := 0
	unavailable := 0
	timeouts := 0  //Счетчик таймаутов
	dnsErrors := 0 //Счетчик DNS ошибок
	canceled := 0  //Счетчик отмен

	for _, res := range results {
		if res.Result.Available {
			fmt.Printf("%s (%v)\n", res.Host, res.Result.Latency)
			available++
		} else {
			if errors.Is(res.Result.Error, pkgerrors.ErrTimeout) {
				fmt.Printf("%s (таймаут)\n", res.Host)
				timeouts++
			} else if errors.Is(res.Result.Error, pkgerrors.ErrDNSFailed) {
				fmt.Printf("%s (DNS ошибка)\n", res.Host)
				dnsErrors++
			} else if errors.Is(res.Result.Error, pkgerrors.ErrCanceled) {
				fmt.Printf("%s (отменено)\n", res.Host)
				canceled++
			} else {
				fmt.Printf("%s\n", res.Host)
			}
			unavailable++
		}
	}

	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Итого: %d доступно, %d недоступно\n", available, unavailable)
	fmt.Printf("Таймаутов: %d\n", timeouts)
	fmt.Printf("DNS ошибок: %d\n", dnsErrors)
	fmt.Printf("Отменено: %d\n", canceled)
	fmt.Printf("Общее время сканирования: %v\n", totalTime)

	if totalTime.Seconds() > 0 {
		fmt.Printf("Средняя скорость: %.2f хостов/сек\n", float64(len(results))/totalTime.Seconds())
	}
}

// printUsage - справка
func printUsage(program string) {
	fmt.Println("Network Scanner v0.3")
	fmt.Println(" %s <хост>\n", program)
	fmt.Println("Пример:")
	fmt.Println(" %s 8.8.8.8\n", program)
	fmt.Println(" %s google.com\n", program)
	fmt.Println("\nТаймаут по умолчанию: %v\n", DefaultTimeout)
	fmt.Println("\n Нажминете Ctrl+C для прерывания")
}
