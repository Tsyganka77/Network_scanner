package batch

import(
	"sync"
	"network-scanner/ping"
)

//Результат сканирования одного хоста
type ScanResult struct {
	Host string
	Result ping.PingResult
}

//ScanHosts сканирует несколько хостов параллельно
//hosts - список хостов
//maxConcurrent - максимум одновременных горутин
func ScanHosts(hosts []string, maxConcurrent int) []ScanResult {
	results := make([]ScanResult, 0, len(hosts))

	//Канал для получения результатов
	//Буферизированный, чтобы горутины не блокировались при отправке
	resultChan := make(chan ScanResult, len(hosts))

	//WaitGroup для ожидания завершения всех горутин
	var wg sync.WaitGroup

	//Семафор для ограничения колва горутин
	semaphore := make(chan struct{}, maxConcurrent)

	//Запускаем горрутины для каждого хоста
	for _, host := range hosts {
		wg.Add(1)
		go func(h string){
			defer wg.Done()

			//Захватываем слот в Семафоре
			semaphore <-struct{}{}

			//освобождаем слот после завершения
			defer func(){<-semaphore}()

			//Выполняем пинг
			result := ping.PingHost(h)

			//Отправляем результат в канал
			resultChan <- ScanResult{
				Host: h,
				Result: result,
			}

		}(host)
	}

	//Отдельная горутина закроет канал, когда все завершится
	go func(){
		wg.Wait()
		close(resultChan)
	}()

	//Собираем все результаты из канала
	for res := range resultChan{
		results = append(results, res)
	}

	return results
}
