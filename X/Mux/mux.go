// Пакет mux предоставляет простой мультиплексор маршрутов сообщений Discord, который
// анализирует сообщения, а затем выполняет соответствующий зарегистрированный обработчик, если он найден.
// Mux может использоваться как с Disgord, так и с библиотекой DiscordGo.
package mux

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Маршрут содержит информацию об определенном обработчике маршрута сообщения.
type Route struct {
	Pattern     string      // шаблон соответствия, который должен запускать этот обработчик маршрута
	Description string      // краткое описание этого маршрута
	Help        string      // подробная справочная строка для этого маршрута
	Run         HandlerFunc // функция обработчика маршрута для вызова
}

// Context содержит немного дополнительных данных,
// которые мы передаем обработчикам маршрутов
// Таким образом, обработка некоторых из них должна происходить только один раз.
type Context struct {
	Fields          []string
	Content         string
	IsDirected      bool
	IsPrivate       bool
	HasPrefix       bool
	HasMention      bool
	HasMentionFirst bool
}

// HandlerFunc - подпись функции, необходимая для обработчика маршрута сообщения.
type HandlerFunc func(*discordgo.Session, *discordgo.Message, *Context)

// Mux - основная структура для всех методов мультиплексирования.
type Mux struct {
	Routes  []*Route
	Default *Route
	Prefix  string
}

// New возвращает новый мультиплексор маршрута сообщений Discord.
func New() *Mux {
	m := &Mux{}
	m.Prefix = "-dg "
	return m
}

// Route позволяет зарегистрировать маршрут.
func (m *Mux) Route(pattern, desc string, cb HandlerFunc) (*Route, error) {

	r := Route{}
	r.Pattern = pattern
	r.Description = desc
	r.Run = cb
	m.Routes = append(m.Routes, &r)

	return &r, nil
}

// FuzzyMatch пытается найти лучший маршрут для данного сообщения.
func (m *Mux) FuzzyMatch(msg string) (*Route, []string) {

	// Преобразование строки сообщения в фрагмент слов.
	fields := strings.Fields(msg)

	// нет смысла продолжать, если нет полей.
	if len(fields) == 0 {
		return nil, nil
	}

	// Найдите совпадение в списке команд.
	var r *Route
	var rank int

	var fk int
	for fk, fv := range fields {

		for _, rv := range m.Routes {

			// Если мы найдем точное совпадение, немедленно верните его.
			if rv.Pattern == fv {
				return rv, fields[fk:]
			}

			// Какой-то "Fuzzy" поиск...
			if strings.HasPrefix(rv.Pattern, fv) {
				if len(fv) > rank {
					r = rv
					rank = len(fv)
				}
			}
		}
	}
	return r, fields[fk:]
}

// OnMessageCreate это функция обработчика событий DiscordGo. Это должно быть
// зарегистрировано с помощью функции DiscordGo.Session.AddHandler. Эта функция
// получит все сообщения Discord и проанализирует их на соответствие зарегистрированным
// маршрутам.
func (m *Mux) OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {

	var err error

	// Игнорировать все сообщения, созданные самой учетной записью бота.
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	// Создаёт Context struct, в которую мы можем помещать различную информацию.
	ctx := &Context{
		Content: strings.TrimSpace(mc.Content),
	}

	// Fetch the channel for this Message.
	var c *discordgo.Channel
	c, err = ds.State.Channel(mc.ChannelID)
	if err != nil {
		// Попробуйте получить через REST API.
		c, err = ds.Channel(mc.ChannelID)
		if err != nil {
			log.Printf("невозможно получить канал для сообщения, %s", err)
		} else {
			// Попытка добавить этот канал в наш State.
			err = ds.State.ChannelAdd(c)
			if err != nil {
				log.Printf("ошибка обновления состояния с помощью канала, %s", err)
			}
		}
	}
	// Добавляет Channel информацию в Context (если мы успешно получили канал).
	if c != nil {
		if c.Type == discordgo.ChannelTypeDM {
			ctx.IsPrivate, ctx.IsDirected = true, true
		}
	}

	// Обнаруживает @name или @nick упоминания.
	if !ctx.IsDirected {

		// Определить, был ли Bot @mentioned
		for _, v := range mc.Mentions {

			if v.ID == ds.State.User.ID {

				ctx.IsDirected, ctx.HasMention = true, true

				reg := regexp.MustCompile(fmt.Sprintf("<@!?(%s)>", ds.State.User.ID))

				// Было ли @mention первой частью строки?
				if reg.FindStringIndex(ctx.Content)[0] == 0 {
					ctx.HasMentionFirst = true
				}

				// удалить теги упоминания бота из строки содержимого.
				ctx.Content = reg.ReplaceAllString(ctx.Content, "")

				break
			}
		}
	}

	// Обнаружить упоминание префикса
	if !ctx.IsDirected && len(m.Prefix) > 0 {

		// TODO : Необходимо изменить, чтобы поддерживать префикс, определяемый пользователем для каждой гильдии.
		if strings.HasPrefix(ctx.Content, m.Prefix) {
			ctx.IsDirected, ctx.HasPrefix, ctx.HasMentionFirst = true, true, true
			ctx.Content = strings.TrimPrefix(ctx.Content, m.Prefix)
		}
	}

	// На данный момент, если мы специально не упомянули, мы ничего не делаем.
	// позже я могу добавить опцию для глобальных не упомянутых командных слов
	if !ctx.IsDirected {
		return
	}

	// // Пытаемся найти в сообщении команду "best match".
	r, fl := m.FuzzyMatch(ctx.Content)
	if r != nil {
		ctx.Fields = fl
		r.Run(ds, mc.Message, ctx)
		return
	}

	// Если совпадений команд не найдено, вызовите значение по умолчанию.
	// Игнорировать, если только @mentioned в середине сообщения
	if m.Default != nil && (ctx.HasMentionFirst) {
		// TODO: здесь можно использовать ratelimit
		// или ratelimit должен быть внутри обработчика cmd? ..
		// В случае "разговора" с другим ботом это может создать бесконечный
		// цикл. Наверное, чаще всего встречается в личных сообщениях.
		m.Default.Run(ds, mc.Message, ctx)
	}

}
