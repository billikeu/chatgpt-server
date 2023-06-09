package backend

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/billikeu/Go-ChatBot/bot"
	"github.com/billikeu/Go-ChatBot/bot/params"
	"github.com/billikeu/chatgpt-server/backend/conf"
	"github.com/billikeu/chatgpt-server/backend/middlewares"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
)

type Server struct {
	bot        *bot.Bot
	serverConf *conf.Config
}

func NewServer(serverConf *conf.Config) *Server {
	s := &Server{
		serverConf: serverConf,
		bot: bot.NewBot(&bot.Config{
			Proxy: serverConf.Proxy, // socks5://10.0.0.13:3126 , http://10.0.0.13:3127
			ChatGPT: bot.ChatGPTConf{
				SecretKey: serverConf.SecretKey, // your secret key
			},
		}),
	}
	return s
}

func (s *Server) Start() {
	r := gin.Default()
	r.Use(middlewares.Cors())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/chat-process", s.chatProcess)
	r.POST("/config", s.config)
	r.POST("/session", s.session)
	r.POST("/verify", s.verify)

	r.Run(fmt.Sprintf("%s:%d", s.serverConf.Host, s.serverConf.Port)) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// /chat-process
/*
req: {"prompt":"2","options":{"conversationId":"2d1fe5e1-fedc-4e2b-bf75-065b1da01abc","parentMessageId":"chatcmpl-70LeQVv1CupSiy6aeio4rt5safFYA"},"systemMessage":"You are ChatGPT, "}
err: {"message":"","data":null,"status":"Fail"}
*/
func (s *Server) chatProcess(c *gin.Context) {
	c.Header("Content-type", "application/octet-stream")

	// c.Request.b
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// log.Println(string(bodyBytes))
	defer c.Request.Body.Close()
	j := gjson.ParseBytes(bodyBytes)
	prompt := j.Get("prompt").String()
	conversationId := j.Get("options.conversationId").String()
	if conversationId == "" {
		conversationId = uuid.NewV4().String()
	}
	// log.Println(j)
	var chunkIndex int
	err = s.bot.Ask(context.Background(), &params.AskParams{
		ConversationId:    conversationId,
		Prompt:            prompt,
		BotType:           params.BotTypeChatGPT,
		SystemRoleMessage: j.Get("options.systemMessage").String(),
		Callback: func(_params *params.CallParams, err error) {
			if err != nil {
				return
			}
			chunkIndex = _params.ChunkIndex
			m := ChatMessage{
				Role:            "assistant",
				ID:              _params.MsgId,
				ParentMessageID: _params.ParentId,
				ConversationID:  conversationId,
				Text:            _params.Text,
				Detail: DetailInfo{
					ID:      _params.MsgId,
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   params.BotTypeChatGPT,
					Choices: []ChoiceInfo{
						{
							Delta: DeltaInfo{
								Content: _params.Chunk,
							},
							Index:        0,
							FinishReason: _params.Done,
						},
					},
				},
			}
			msg := m.String()
			if _params.ChunkIndex > 1 {
				msg = "\n" + msg
			}
			c.Writer.Write([]byte(msg))
			c.Writer.Flush()
		},
	})
	if err != nil {
		m := FailMessage{
			Message: err.Error(),
			Status:  "fail",
			Data:    nil,
		}
		msg := m.String()
		if chunkIndex > 1 {
			msg = "\n" + msg
		}
		c.Writer.Write([]byte(msg))
		c.Writer.Flush()
	}
}

// /config
func (s *Server) config(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": nil,
		"data": gin.H{
			"apiModel":     "ChatGPTAPI",
			"reverseProxy": "-",
			"timeoutMs":    60000,
			"socksProxy":   "-",
			"httpsProxy":   "-",
			"balance":      "-",
		},
		"status": "Success",
	})
}

// /session
func (s *Server) session(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "",
		"data": gin.H{
			"auth":  false,
			"model": "ChatGPTAPI",
		}})
}

// /verify
func (s *Server) verify(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
