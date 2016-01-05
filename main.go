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

	router.Run(":8080")
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
			Text: "Какой вариант выберете?",
			Answers: []answer{
				answer{Text: "Этот"},
				answer{Text: "Другой"},
			},
		},
	}
}
