package batch

import(
	"fmt"
	"time"
	"context"
	"sync" //Для работы с горутинами
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
func ScanHosts(ctx context.Context, hosts []string, maxConcurrent int, timeout time.Duration) []ScanResult {
	results := make([]ScanResult, 0, len(hosts))

	//Канал для получения результатов
	//Буферизированный, чтобы горутины не блокировались при отправке
	resultChan := make(chan ScanResult, len(hosts)) //Создает ссылочный тип данных

	//WaitGroup для ожидания завершения всех горутин
	var wg sync.WaitGroup

	//Семафор для ограничения колва горутин
	semaphore := make(chan struct{}, maxConcurrent)

	//Запускаем горрутины для каждого хоста
	for _, host := range hosts {

		select{
		case<-ctx.Done():
			fmt.Println("Сканирование прервано")
			close(resultChan)
			return results
		default:
		}

		wg.Add(1)
		go func(h string){
			defer wg.Done()

			select{
			case <- ctx.Done():
				return
				//Захватываем слот в Семафоре
				case semaphore <-struct{}{}:
			}
			//освобождаем слот после завершения
			defer func(){<-semaphore}()

			//Выполняем пинг
			result := ping.PingHost(ctx, h, timeout)

			select{
			case <- ctx.Done():
				return
			//Отправляем результат в канал
			case resultChan <- ScanResult{
				Host: h,
				Result: result,
			}:
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
		select{
		case <- ctx.Done():
			return results
		default:
			results = append(results, res)
		}
	}

	return results
}
