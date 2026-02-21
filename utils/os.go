package utils

import "os"
//Проверка работает ли программа на винде
//Названиес большой буквы - функция доступна снаружи пакета
func IsWindows() bool { 
	//в винде переменная окружения OS равна "Windows_NT"
        return os.Getenv("OS") == "Windows_NT"
}

