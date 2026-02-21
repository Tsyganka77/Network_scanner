package main

import(
	"fmt"    //ввод вывод
	"os"    //работа с os
	"strings"    //работа со строками
	"time"    //работа со временем
	"network-scanner/ping"    //подключаем пакет ping
)

//Константы

const(
	//Сообщение об успешном пинге
	SuccessMsg = "Хост доступен"
	//Сообщение об ошибке
	ErrorMsg = "Хост недоступен"
)

//main - точка входа
func main(){
	arguments:= os.Args	//os.Args в Go — это срез строк ([]string), содержащий аргументы
				//командной строки, переданные при запуске программы
	if len(arguments)<2{
		fmt.Println("Network Scanner v0.2")
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
//выполняем пинг из другого пакета
result := ping.PingHost(host)
//считаем общее время
totalTime := time.Since(startTime)
//выводим результат
PrintResult(result, totalTime)
}

//printResult - вывод результатов

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
