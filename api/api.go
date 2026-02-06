package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
// @Summary Configuração de rotas
// @Description Configura todas as rotas disponíveis na API
func (api *API) setupRoutes() {
	// Rota de health check
	// @Summary Health check
	// @Description Verifica se a API está funcionando
	// @Tags System
	// @Produce json
	// @Success 200 {object} map[string]interface{} "API funcionando"
	// @Router /health [get]
	api.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "AgroChat API está rodando!",
		})
	})

	// Documentação Swagger
	api.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Grupo de rotas para WhatsApp
	whatsapp := api.Router.Group("/whatsapp")
	{
		// Enviar mensagem
		whatsapp.POST("/send", api.sendMessage)

		// Verificar status da conexão
		whatsapp.GET("/status", api.getStatus)
	}

	// Endpoint principal para integração com AgroServer
	api.Router.POST("/enviar-verificacao", api.enviarVerificacao)

	// Rota simples para enviar mensagem
	api.Router.POST("/enviar-mensagem", api.enviarMensagemSimples)
}

// sendMessage envia uma mensagem via WhatsApp
// @Summary Envia mensagem via WhatsApp (endpoint antigo)
// @Description Envia uma mensagem para um número de WhatsApp com verificações detalhadas
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param body body object{phone=string,message=string} true "Dados da mensagem"
// @Success 200 {object} map[string]interface{} "Mensagem enviada com sucesso"
// @Failure 400 {object} map[string]interface{} "Dados inválidos"
// @Failure 500 {object} map[string]interface{} "Erro no servidor"
// @Failure 503 {object} map[string]interface{} "WhatsApp não conectado"
// @Router /whatsapp/send [post]
func (api *API) sendMessage(c *gin.Context) {
	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se está conectado
	if !api.Client.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WhatsApp não está conectado",
		})
		return
	}

	// Formatar número no formato internacional
	jid, err := api.parseJID(req.Phone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número inválido: " + err.Error()})
		return
	}

	fmt.Printf("[SEND] Enviando mensagem para: %s (JID: %s)\n", req.Phone, jid.String())

	// Verificar se o número existe no WhatsApp (apenas informativo)
	isOnWhatsApp, err := api.Client.IsOnWhatsApp(c.Request.Context(), []string{req.Phone})
	if err == nil && len(isOnWhatsApp) > 0 {
		if isOnWhatsApp[0].IsIn {
			fmt.Printf("[SUCCESS] Número confirmado no WhatsApp!\n")
		} else {
			fmt.Printf("[WARNING] Número NÃO está no WhatsApp! Enviando mesmo assim...\n")
		}
	}

	fmt.Printf("[MESSAGE] Mensagem: %s\n", req.Message)

	// Enviar mensagem
	resp, err := api.Client.SendMessage(c.Request.Context(), jid, &waProto.Message{
		Conversation: &req.Message,
	})

	if err != nil {
		fmt.Printf("[ERROR] Erro ao enviar: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao enviar mensagem: " + err.Error(),
		})
		return
	}

	fmt.Printf("[SUCCESS] Mensagem enviada! Timestamp: %v, ID: %s\n", resp.Timestamp, resp.ID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Mensagem enviada com sucesso!",
		"data": gin.H{
			"phone":     req.Phone,
			"jid":       jid.String(),
			"text":      req.Message,
			"messageId": resp.ID,
			"timestamp": resp.Timestamp,
		},
	})
}

// getStatus retorna o status da conexão do WhatsApp
// @Summary Status da conexão WhatsApp
// @Description Retorna informações sobre o status da conexão com o WhatsApp
// @Tags WhatsApp
// @Produce json
// @Success 200 {object} map[string]interface{} "Status da conexão"
// @Router /whatsapp/status [get]
func (api *API) getStatus(c *gin.Context) {
	isConnected := api.Client.IsConnected()

	c.JSON(http.StatusOK, gin.H{
		"connected": isConnected,
		"device": gin.H{
			"id": api.Client.Store.ID,
		},
	})
}

// parseJID converte número de telefone para JID do WhatsApp
func (api *API) parseJID(phone string) (types.JID, error) {
	// Remover caracteres não numéricos
	phone = strings.ReplaceAll(phone, "+", "")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	fmt.Printf("[DEBUG] Número original: %s (tamanho: %d)\n", phone, len(phone))

	// Se não tem código do país, adicionar 55 (Brasil)
	if !strings.HasPrefix(phone, "55") && len(phone) >= 10 && len(phone) <= 11 {
		// Verificar se começa com DDD brasileiro válido (11-99)
		if len(phone) >= 2 {
			fmt.Printf("   [FIX] Adicionando código do país 55\n")
			fmt.Printf("   [OLD] Sem código: %s\n", phone)
			phone = "55" + phone
			fmt.Printf("   [NEW] Com código: %s\n", phone)
		}
	}

	// Tratar números brasileiros - converter automaticamente para formato correto
	if strings.HasPrefix(phone, "55") && len(phone) >= 12 {
		ddd := phone[2:4]
		numero := phone[4:]

		fmt.Printf("   [INFO] Número BR detectado - DDD: %s, Número: %s (tamanho: %d)\n", ddd, numero, len(numero))

		// Se o número tem 9 dígitos e começa com 9, remover o 9 inicial
		if len(numero) == 9 && numero[0] == '9' {
			numeroCorrigido := numero[1:] // Remove o primeiro 9
			phoneCorrigido := "55" + ddd + numeroCorrigido
			fmt.Printf("   [FIX] CORRIGINDO: 9 dígitos → 8 dígitos\n")
			fmt.Printf("   [OLD] Formato antigo: %s\n", phone)
			fmt.Printf("   [NEW] Formato correto: %s\n", phoneCorrigido)
			phone = phoneCorrigido
		} else if len(numero) == 8 {
			fmt.Printf("   [OK] Formato correto: 8 dígitos\n")
		} else {
			fmt.Printf("   [WARNING] Formato não padrão: %d dígitos\n", len(numero))
		}
	}

	// Verificar se é um número válido
	if len(phone) < 10 {
		return types.JID{}, fmt.Errorf("número muito curto (mínimo 10 dígitos)")
	}

	fmt.Printf("   [INFO] Número final para envio: %s\n", phone)
	return types.NewJID(phone, types.DefaultUserServer), nil
}

// VerificacaoRequest representa a requisição de envio de verificação
type VerificacaoRequest struct {
	NomeCliente       string `json:"nomeCliente" binding:"required"`
	NomeVendedor      string `json:"nomeVendedor" binding:"required"`
	Documento         string `json:"documento" binding:"required"`
	Telefone          string `json:"telefone" binding:"required"`
	Endereco          string `json:"endereco"`
	CodigoVerificacao string `json:"codigoVerificacao" binding:"required"`
	Metodo            string `json:"metodo" binding:"required"`
}

// enviarVerificacao envia código de verificação via WhatsApp
// @Summary Enviar código de verificação
// @Description Envia código de verificação formatado para clientes do AgroServer
// @Tags Verificação
// @Accept json
// @Produce json
// @Param body body VerificacaoRequest true "Dados da verificação"
// @Success 200 {object} map[string]interface{} "Verificação enviada com sucesso"
// @Failure 400 {object} map[string]interface{} "Dados inválidos"
// @Failure 500 {object} map[string]interface{} "Erro no servidor"
// @Failure 503 {object} map[string]interface{} "WhatsApp não conectado"
// @Router /enviar-verificacao [post]
func (api *API) enviarVerificacao(c *gin.Context) {
	var req VerificacaoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"sucesso":  false,
			"mensagem": "Dados inválidos",
			"erro":     err.Error(),
		})
		return
	}

	// Validar método
	if req.Metodo != "whatsapp" && req.Metodo != "sms" {
		c.JSON(http.StatusBadRequest, gin.H{
			"sucesso":  false,
			"mensagem": "Método inválido",
			"erro":     "Método deve ser 'whatsapp' ou 'sms'",
		})
		return
	}

	// Por enquanto, apenas WhatsApp está implementado
	if req.Metodo == "sms" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"sucesso":  false,
			"mensagem": "SMS não implementado",
			"erro":     "Atualmente apenas WhatsApp está disponível",
		})
		return
	}

	// Verificar se está conectado
	if !api.Client.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"sucesso":  false,
			"mensagem": "WhatsApp não está conectado",
			"erro":     "Serviço temporariamente indisponível",
		})
		return
	}

	// Formatar número
	jid, err := api.parseJID(req.Telefone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"sucesso":  false,
			"mensagem": "Número de telefone inválido",
			"erro":     err.Error(),
		})
		return
	}

	// Formatar mensagem
	mensagem := api.formatarMensagemVerificacao(req)

	// Log da operação
	fmt.Printf("\n[VERIFICATION] ===== ENVIO DE VERIFICAÇÃO =====\n")
	fmt.Printf("[CLIENT] Cliente: %s\n", req.NomeCliente)
	fmt.Printf("[VENDOR] Vendedor: %s\n", req.NomeVendedor)
	fmt.Printf("[PHONE] Telefone: %s → %s\n", req.Telefone, jid.String())
	fmt.Printf("[CODE] Código: %s\n", req.CodigoVerificacao)
	fmt.Printf("===============================================\n\n")

	// Enviar mensagem
	resp, err := api.Client.SendMessage(c.Request.Context(), jid, &waProto.Message{
		Conversation: &mensagem,
	})

	if err != nil {
		fmt.Printf("[ERROR] Erro ao enviar: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"sucesso":  false,
			"mensagem": "Erro ao enviar mensagem",
			"erro":     err.Error(),
		})
		return
	}

	// Gerar ID único para rastreamento
	idMensagem := "msg_" + uuid.New().String()[:8]

	fmt.Printf("[SUCCESS] Verificação enviada com sucesso!\n")
	fmt.Printf("[ID] %s\n", idMensagem)
	fmt.Printf("[TIME] Timestamp: %v\n\n", resp.Timestamp)

	c.JSON(http.StatusOK, gin.H{
		"sucesso":    true,
		"mensagem":   "Mensagem enviada com sucesso",
		"idMensagem": idMensagem,
		"dataEnvio":  time.Now().Format(time.RFC3339),
	})
}

// formatarMensagemVerificacao cria a mensagem formatada para o cliente
func (api *API) formatarMensagemVerificacao(req VerificacaoRequest) string {
	docFormatado := formatarDocumento(req.Documento)

	mensagem := fmt.Sprintf(`🌾 *AgroSystem - Verificação de Cliente*

Olá *%s*!

Seu vendedor *%s* iniciou um cadastro para você.
`, req.NomeCliente, req.NomeVendedor)

	if req.Endereco != "" {
		mensagem += fmt.Sprintf("\n📍 Endereço: %s", req.Endereco)
	}

	mensagem += fmt.Sprintf(`
📄 Documento: %s

🔐 Código de verificação: *%s*

Por favor, compartilhe este código com seu vendedor para confirmar seus dados.

_Válido por 15 minutos_`, docFormatado, req.CodigoVerificacao)

	return mensagem
}

// formatarDocumento formata CPF ou CNPJ
func formatarDocumento(doc string) string {
	// Remover caracteres não numéricos
	doc = strings.ReplaceAll(doc, ".", "")
	doc = strings.ReplaceAll(doc, "-", "")
	doc = strings.ReplaceAll(doc, "/", "")
	doc = strings.TrimSpace(doc)

	// Formatar CPF (11 dígitos)
	if len(doc) == 11 {
		return fmt.Sprintf("%s.%s.%s-%s", doc[0:3], doc[3:6], doc[6:9], doc[9:11])
	}

	// Formatar CNPJ (14 dígitos)
	if len(doc) == 14 {
		return fmt.Sprintf("%s.%s.%s/%s-%s", doc[0:2], doc[2:5], doc[5:8], doc[8:12], doc[12:14])
	}

	// Retornar original se não for CPF nem CNPJ
	return doc
}

// enviarMensagemSimples envia uma mensagem simples para um número
// @Summary Enviar mensagem simples
// @Description Envia uma mensagem de texto simples para um número de WhatsApp
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param body body object{numero=string,mensagem=string} true "Número e mensagem"
// @Success 200 {object} map[string]interface{} "Mensagem enviada com sucesso"
// @Failure 400 {object} map[string]interface{} "Dados inválidos"
// @Failure 500 {object} map[string]interface{} "Erro no servidor"
// @Failure 503 {object} map[string]interface{} "WhatsApp não conectado"
// @Router /enviar-mensagem [post]
func (api *API) enviarMensagemSimples(c *gin.Context) {
	var req struct {
		Numero   string `json:"numero" binding:"required"`
		Mensagem string `json:"mensagem" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"sucesso":  false,
			"mensagem": "Dados inválidos",
			"erro":     err.Error(),
		})
		return
	}

	// Verificar se está conectado
	if !api.Client.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"sucesso":  false,
			"mensagem": "WhatsApp não está conectado",
			"erro":     "Serviço temporariamente indisponível",
		})
		return
	}

	// Formatar número
	jid, err := api.parseJID(req.Numero)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"sucesso":  false,
			"mensagem": "Número de telefone inválido",
			"erro":     err.Error(),
		})
		return
	}

	// Log
	fmt.Printf("\n[SEND] Enviando mensagem\n")
	fmt.Printf("[TO] %s → %s\n", req.Numero, jid.String())
	fmt.Printf("[MSG] %s\n\n", req.Mensagem)

	// Enviar mensagem
	resp, err := api.Client.SendMessage(c.Request.Context(), jid, &waProto.Message{
		Conversation: &req.Mensagem,
	})

	if err != nil {
		fmt.Printf("[ERROR] Erro ao enviar: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"sucesso":  false,
			"mensagem": "Erro ao enviar mensagem",
			"erro":     err.Error(),
		})
		return
	}

	fmt.Printf("[SUCCESS] Mensagem enviada!\n")
	fmt.Printf("[TIME] Timestamp: %v\n\n", resp.Timestamp)

	c.JSON(http.StatusOK, gin.H{
		"sucesso":   true,
		"mensagem":  "Mensagem enviada com sucesso",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Start inicia o servidor da API
func (api *API) Start(port string) error {
	fmt.Printf("\n[INFO] API iniciada em http://localhost:%s\n", port)
	fmt.Printf("[INFO] Documentação Swagger disponível em: http://localhost:%s/swagger/index.html\n\n", port)
	return api.Router.Run(":" + port)
}
