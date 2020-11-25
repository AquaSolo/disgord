// Объявите этот файл частью основного пакета, чтобы его можно было собрать в
// исполняемый файл.
package main

// Добавьте все необходимые пакеты Go.
import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Версия - это константа, в которой хранится информация о версии Disgord.
const Version = "v0.0.0-alpha"

// Session объявлен в глобальном пространстве, поэтому его можно легко использовать
// в любом месте программы.
// В этом случае ошибки не будет. Поэтому мы не будем её обрабатывать.
var Session, _ = discordgo.New()

// Считываем все параметры конфигурации из переменных
// среды и аргументов командной строки.
func init() {

	// Токен аутентификации Discord.
	Session.Token = os.Getenv("DG_TOKEN")
	if Session.Token == "" {
		flag.StringVar(&Session.Token, "t", "", "Discord Authentication Token")
	}
}

func main() {

	// Объявите все необходимые здесь.
	var err error

	// Распечатайте красивый логотип.
	fmt.Printf(` 
	________  .__                               .___
	\______ \ |__| ______ ____   ___________  __| _/
	||    |  \|  |/  ___// ___\ /  _ \_  __ \/ __ | 
	||    '   \  |\___ \/ /_/  >  <_> )  | \/ /_/ | 
	||______  /__/____  >___  / \____/|__|  \____ | 
	\_______\/        \/_____/   %-16s\/`+"\n\n", Version)

	// Анализируйте аргументы командной строки.
	flag.Parse()

	// Убедитесь, что токен был предоставлен.
	if Session.Token == "" {
		log.Println("You must provide a Discord authentication token.")
		return
	}

	// Откройте подключение к Discord через веб-соединение.
	err = Session.Open()
	if err != nil {
		log.Printf("error opening connection to Discord, %s\n", err)
		os.Exit(1)
	}

	// Ждём CTRL-C.
	log.Printf(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Заканчиваем.
	Session.Close()

	// Выход.
}
