package mux

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

// Help функция обеспечивает встроенную команду "help" команда, которая отобразит список
// всех зарегистрированных маршрутов (commands).
// Чтобы использовать эту функцию, она должна быть сначала
// зарегистрирована в функции Mux.Route.
func (m *Mux) Help(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {

	// Установите префикс команды для отображения.
	cp := ""
	if ctx.IsPrivate {
		cp = ""
	} else if ctx.HasPrefix {
		cp = m.Prefix
	} else {
		cp = fmt.Sprintf("@%s ", ds.State.User.Username)
	}

	// Сортирует команды.
	maxlen := 0
	keys := make([]string, 0, len(m.Routes))
	cmdmap := make(map[string]*Route)

	for _, v := range m.Routes {

		// Отображать только команды с описанием.
		if v.Description == "" {
			continue
		}

		// Вычислить максимальную длину строки command + args.
		l := len(v.Pattern) // TODO: Добавьте часть +args :)
		if l > maxlen {
			maxlen = l
		}

		cmdmap[v.Pattern] = v

		// help и about добавляются отдельно ниже.
		if v.Pattern == "help" || v.Pattern == "about" {
			continue
		}

		keys = append(keys, v.Pattern)
	}

	sort.Strings(keys)

	// TODO: Ссылка "Узнать больше" должна быть настраиваемой.
	resp := "\n*Команды можно сокращать и смешивать с другим текстом. Узнайте больше на <https://github.com/bwmarrin/disgord>*\n"
	resp += "```autoit\n"

	v, ok := cmdmap["help"]
	if ok {
		keys = append([]string{v.Pattern}, keys...)
	}

	v, ok = cmdmap["about"]
	if ok {
		keys = append([]string{v.Pattern}, keys...)
	}

	// Добавить отсортированный результат в помощь msg
	for _, k := range keys {
		v := cmdmap[k]
		resp += fmt.Sprintf("%s%-"+strconv.Itoa(maxlen)+"s # %s\n", cp, v.Pattern+v.Help, v.Description)
	}

	resp += "```\n"

	ds.ChannelMessageSend(dm.ChannelID, resp)

	return
}
