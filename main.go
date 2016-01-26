package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/gin-gonic/gin"
	"github.com/tommy351/gin-sessions"
)

type (
	question struct {
		Text    string
		Answers []answer
	}

	answer struct {
		Text  string
		Votes int64
	}
)

var (
	questions            []question
	currentQuestionIndex int64
	currentGeneration    int64
)

func main() {
	staticBox := rice.MustFindBox("static")
	templateBox := rice.MustFindBox("templates")
	templateString, err := templateBox.String("question.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmplQuestion, err := template.New("question").Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	store := sessions.NewCookieStore([]byte("secret100to1"))
	router.Use(sessions.Middleware("100to1", store))
	router.SetHTMLTemplate(tmplQuestion)
	router.StaticFS("/static", staticBox.HTTPBox())

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/play/user")
	})

	router.GET("/play/:controls", func(c *gin.Context) {
		session := sessions.Get(c)

		generation := session.Get("generation")
		if generation == nil {
			session.Set("generation", currentGeneration)
			session.Save()
		} else if generation.(int64) != currentGeneration {
			session.Clear()
			session.Save()
		}

		answerKey := fmt.Sprintf("answer%d", currentQuestionIndex)
		currentAnswerIndex := -1
		if session.Get(answerKey) != nil {
			currentAnswerIndex = session.Get(answerKey).(int)
		}
		c.HTML(http.StatusOK, "question", gin.H{
			"currentQuestionIndex": currentQuestionIndex,
			"question":             questions[currentQuestionIndex],
			"controls":             c.Param("controls"),
			"currentAnswerIndex":   currentAnswerIndex,
		})
	})

	router.GET("/play/:controls/shiftquestion/:shift", func(c *gin.Context) {
		if c.Param("controls") == "vih" {
			if c.Param("shift") == "next" && currentQuestionIndex < int64(len(questions)-1) {
				atomic.AddInt64(&currentQuestionIndex, 1)
			} else if c.Param("shift") == "prev" && currentQuestionIndex > 0 {
				atomic.AddInt64(&currentQuestionIndex, -1)
			}
			c.Redirect(http.StatusFound, "/play/vih")
		} else {
			c.Redirect(http.StatusFound, "/play/user")
		}
	})

	router.GET("/play/:controls/setanswer/:questionIndex/:answerIndex", func(c *gin.Context) {
		questionIndex, _ := strconv.Atoi(c.Param("questionIndex"))
		newAnswerIndex, _ := strconv.Atoi(c.Param("answerIndex"))
		atomic.AddInt64(&questions[questionIndex].Answers[newAnswerIndex].Votes, 1)

		session := sessions.Get(c)
		answerKey := "answer" + c.Param("questionIndex")
		currentAnswerIndex := session.Get(answerKey)
		if currentAnswerIndex != nil {
			atomic.AddInt64(&questions[questionIndex].Answers[currentAnswerIndex.(int)].Votes, -1)
		}
		session.Set(answerKey, newAnswerIndex)
		session.Save()
		c.Redirect(http.StatusFound, "/play/user")
	})

	router.Run(":80")
}

func init() {
	currentGeneration = time.Now().Unix()

	questions = []question{
		question{
			Text: "Что Сизому нравится в девочках?",
			Answers: []answer{
				answer{Text: "Длинные волосы"},
				answer{Text: "Статус мамы"},
				answer{Text: "Лайкает шутки"},
				answer{Text: "Любит ездить в мерседесе"},
				answer{Text: "Баба Зина"},
				answer{Text: "Любит пересдачи"},
			},
		},
		question{
			Text: "Зачем ты приехал на ЗМШ?",
			Answers: []answer{
				answer{Text: "За мальчиком"},
				answer{Text: "За девочкой"},
				answer{Text: "За мечтой"},
				answer{Text: "На пересдачу"},
				answer{Text: "Я молодой ученый"},
				answer{Text: "Я не приехал"},
			},
		},
		question{
			Text: "Что вы взяли с собой на Зимнюю школу?",
			Answers: []answer{
				answer{Text: "Упаковку предохранителей"},
				answer{Text: "Горячительное"},
				answer{Text: "Голову"},
				answer{Text: "Книгу"},
				answer{Text: "Мел"},
				answer{Text: "Самовар"},
			},
		},
		question{
			Text: "Кто является работником Зимней школы?",
			Answers: []answer{
				answer{Text: "Мамы"},
				answer{Text: "Шур"},
				answer{Text: "Лившиц"},
				answer{Text: "Дунаев"},
				answer{Text: "Щеколдин"},
				answer{Text: "Я"},
			},
		},
		question{
			Text: "Если они пойдут в баню, то кому повезет больше?",
			Answers: []answer{
				answer{Text: "Мне"},
				answer{Text: "Лехтику"},
				answer{Text: "Багняку"},
				answer{Text: "Галке"},
				answer{Text: "Звереву"},
				answer{Text: "Ксюше и Маше"},
			},
		},
		question{
			Text: "Что важнее?",
			Answers: []answer{
				answer{Text: "Где"},
				answer{Text: "С кем"},
				answer{Text: "Как"},
				answer{Text: "По чем"},
				answer{Text: "Зачем"},
				answer{Text: "Когда"},
			},
		},
		question{
			Text: "С каким зверем ассоциирует себя Зверев А?",
			Answers: []answer{
				answer{Text: "Лев"},
				answer{Text: "Козел"},
				answer{Text: "Утконос"},
				answer{Text: "Опоссум"},
				answer{Text: "Кролик"},
				answer{Text: "Орел"},
			},
		},
		question{
			Text: "Любимое блюдо Ще в столовой?",
			Answers: []answer{
				answer{Text: "Щи"},
				answer{Text: "Лещи"},
				answer{Text: "Борщи"},
				answer{Text: "Щавель"},
				answer{Text: "Мощи"},
				answer{Text: "Хрящи"},
			},
		},
		question{
			Text: "Что длиннее?",
			Answers: []answer{
				answer{Text: "Кипятильник"},
				answer{Text: "Локоть Дунаева"},
				answer{Text: "Тапок Ще"},
				answer{Text: "Рулон"},
				answer{Text: "Лекция"},
				answer{Text: "Гипотенуза"},
			},
		},
		question{
			Text: "Зачем ты идёшь в мамскую?",
			Answers: []answer{
				answer{Text: "За чаем"},
				answer{Text: "Поговорить"},
				answer{Text: "Кипяток в доширак"},
				answer{Text: "Принести ведро"},
				answer{Text: "Воткнуть кипятильник"},
				answer{Text: "Отдать лимоны"},
			},
		},
		question{
			Text: "Зачем ты идешь в баню?",
			Answers: []answer{
				answer{Text: "Чтобы париться"},
				answer{Text: "Чтобы не париться"},
				answer{Text: "Ты Лехтик"},
				answer{Text: "Помыться"},
				answer{Text: "За мылом"},
				answer{Text: "Потом помыться"},
			},
		},
		question{
			Text: "Продолжите фразу «меня все знают я ведь ...»",
			Answers: []answer{
				answer{Text: "Баба Зина"},
				answer{Text: "Глеб Багняк"},
				answer{Text: "Максимка в костюме-снежинке"},
				answer{Text: "Кепка"},
				answer{Text: "Зампотех"},
				answer{Text: "Пикапер 88 левела"},
			},
		},
		question{
			Text: "Сколько на самом деле раз был женат Сизый?",
			Answers: []answer{
				answer{Text: "1"},
				answer{Text: "2"},
				answer{Text: "е"},
				answer{Text: "3"},
				answer{Text: "пи"},
				answer{Text: "5"},
			},
		},
		question{
			Text: "Какой любимый спорт лыжника Германа?",
			Answers: []answer{
				answer{Text: "Академическая гребля"},
				answer{Text: "Трехколесный велосипед"},
				answer{Text: "Фигурное катание на дальность"},
				answer{Text: "Прыжки с высоты"},
				answer{Text: "Триоблом"},
				answer{Text: "Лапта"},
			},
		},
		question{
			Text: "Кто победит?",
			Answers: []answer{
				answer{Text: "Шур"},
				answer{Text: "Лехтик"},
				answer{Text: "Федоров"},
				answer{Text: "Лившиц"},
				answer{Text: "Зверев"},
				answer{Text: "Фрешер"},
			},
		},
	}
}
