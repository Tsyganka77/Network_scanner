package main

import(
	"fmt"    //ввод вывод
	"os"    //работа с os
	"strings"    //работа со строками
	"time"    //работа со временем
	"network-scanner/ping"    //подключаем пакет ping
	"network-scanner/batch"    //подключаем пакет batch
)

//Константы

const(
	//Сообщение об успешном пинге
	SuccessMsg = "Хост доступен"
	//Сообщение об ошибке
	ErrorMsg = "Хост недоступен"
	//Максимум параллельных пингов
	MaxConcurrent = 10
)

//main - точка входа
func main(){
	arguments:= os.Args	//os.Args в Go — это срез строк ([]string), содержащий аргументы
				//командной строки, переданные при запуске программы
	//Режим 1: 1 хост
	if len(arguments) == 2{
		runSingleHost(arguments[1])
		return
	}

	//Режим 2: несколько хостов
	if len(arguments) > 2{
		runMultipleHosts(arguments[1:])
		return
	}

	//Режим 0: нет аргументов - справка
	printUsage(arguments[0])
	os.Exit(1)
}

//runSingleHost - сканирование одного хоста
func runSingleHost(host string) {
	fmt.Printf("Сканирование хоста: %s\n", host)
	fmt.Println(strings.Repeat("-",40))
	//засекаем время начала
	startTime := time.Now()
	//выполняем пинг из другого пакета
	result := ping.PingHost(host)
	//считаем общее время
	totalTime := time.Since(startTime)
	//выводим результат
	PrintResult(result, totalTime)
}

//runMultipleHosts - сканирование нескольких хостов параллельно
func runMultipleHosts(hosts []string) {
	fmt.Printf("Сканирование %d хостов (параллельно: %d)\n",len(hosts), MaxConcurrent)
        fmt.Println(strings.Repeat("-",40))
        
        startTime := time.Now()
        //Запускаем параллельное сканирование через batch
        results:= batch.ScanHosts(hosts, MaxConcurrent)
        
        totalTime := time.Since(startTime)
        //выводим результат
        PrintMultipleResults(results, totalTime)
}


//PrintResult - вывод результатов
func PrintResult(result ping.PingResult, totalTime time.Duration){
	fmt.Println()

	if result.Available{
		//Успех
		fmt.Printf(" %s %s\n", SuccessMsg, result.Host)
		fmt.Printf(" Время ответа: %v\n", result.Latency)
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

//PrintMultipleResults - вывод результатов нескольких хостов
func PrintMultipleResults(results []batch.ScanResult, totalTime time.Duration){
	fmt.Println()

	available := 0
	unavailable := 0

	for _, res := range results {
		if res.Result.Available {
			fmt.Printf("%s (%v)\n", res.Host, res.Result.Latency)
			available++
		}else{
			fmt.Printf("%s\n", res.Host)
			unavailable++
		}
	}

	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Итого: %d доступно, %d недоступно\n", available, unavailable)
	fmt.Printf("Общее время сканирования: %v\n", totalTime)

	if totalTime.Seconds()>0{
		fmt.Printf("Средняя скорость: %.2f хостов/сек\n", float64(len(results))/totalTime.Seconds())
	}
}

//printUsage - справка
func printUsage(program string){
	fmt.Println("Network Scanner v0.3")
	fmt.Println(" %s <хост>\n", program)
	fmt.Println("Пример:")
	fmt.Println(" %s 8.8.8.8\n", program)
	fmt.Println(" %s google.com\n", program)
}
