package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mau.fi/whatsmeow"
)

type API struct {
	Client *whatsmeow.Client
	Router *gin.Engine
}

// NewAPI cria uma nova instância da API
func NewAPI(client *whatsmeow.Client) *API {
	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := &API{
		Client: client,
		Router: router,
	}

	api.setupRoutes()
	return api
}

// setupRoutes configura todas as rotas da API
func (api *API) setupRoutes() {
	// Rota de health check
	api.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "AgroChat API está rodando!",
		})
	})

	// Grupo de rotas para WhatsApp
	whatsapp := api.Router.Group("/whatsapp")
	{
		// Enviar mensagem
		whatsapp.POST("/send", api.sendMessage)
		
		// Verificar status da conexão
		whatsapp.GET("/status", api.getStatus)
	}
}

// sendMessage envia uma mensagem via WhatsApp
func (api *API) sendMessage(c *gin.Context) {
	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implementar lógica de envio de mensagem
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Mensagem enviada!",
		"data": gin.H{
			"phone": req.Phone,
			"text":  req.Message,
		},
	})
}

// getStatus retorna o status da conexão do WhatsApp
func (api *API) getStatus(c *gin.Context) {
	isConnected := api.Client.IsConnected()
	
	c.JSON(http.StatusOK, gin.H{
		"connected": isConnected,
		"device": gin.H{
			"id": api.Client.Store.ID,
		},
	})
}

// Start inicia o servidor da API
func (api *API) Start(port string) error {
	return api.Router.Run(":" + port)
}
