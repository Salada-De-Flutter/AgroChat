package main

import (
	"agrochat/api"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// killPortProcess mata o processo que está usando a porta especificada
func killPortProcess(port string) {
	fmt.Printf("[CHECK] Verificando porta %s...\n", port)

	if runtime.GOOS == "windows" {
		// Windows: usar PowerShell para encontrar e matar o processo
		psCmd := fmt.Sprintf("$p = Get-NetTCPConnection -LocalPort %s -ErrorAction SilentlyContinue; if($p) { Stop-Process -Id $p.OwningProcess -Force; Write-Host '[KILL] Processo encerrado' } else { Write-Host '[OK] Porta livre' }", port)
		cmd := exec.Command("powershell", "-Command", psCmd)
		output, _ := cmd.CombinedOutput()
		fmt.Print(string(output))
	} else {
		// Linux/Mac: usar lsof e kill
		cmd := exec.Command("sh", "-c", fmt.Sprintf("kill -9 $(lsof -ti:%s) 2>/dev/null && echo '[KILL] Processo encerrado' || echo '[OK] Porta livre'", port))
		output, _ := cmd.Output()
		fmt.Print(string(output))
	}
}

func main() {
	fmt.Println("=== AgroChat - WhatsApp Bot ===")

	// Liberar porta 8080 antes de iniciar
	killPortProcess("8080")
	time.Sleep(1 * time.Second) // Aguardar liberação da porta

	// Configurar logging
	dbLog := waLog.Stdout("Database", "INFO", true)

	// Criar contexto
	ctx := context.Background()

	// String de conexão PostgreSQL
	dbConnStr := "host=flutterbox port=5432 user=flutter password=4002 dbname=AgroChatDB sslmode=disable"

	// Criar container de armazenamento PostgreSQL
	container, err := sqlstore.New(ctx, "postgres", dbConnStr, dbLog)
	if err != nil {
		panic(err)
	}

	// Obter primeiro dispositivo do banco ou criar novo
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Registrar handler de eventos
	client.AddEventHandler(eventHandler)

	// Se não estiver conectado, mostrar QR Code
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		fmt.Println("\nEscaneie o QR Code abaixo com seu WhatsApp:")
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Já está autenticado, apenas conectar
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("\n✅ Bot conectado com sucesso!")
	fmt.Println("Aguardando mensagens...")

	// Iniciar API REST
	apiServer := api.NewAPI(client)
	go func() {
		fmt.Println("\n🚀 API iniciando na porta 8080...")
		if err := apiServer.Start("8080"); err != nil {
			fmt.Printf("Erro ao iniciar API: %v\n", err)
		}
	}()

	// Manter o programa rodando
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nDesconectando...")
	client.Disconnect()
}

func eventHandler(evt interface{}) {
	// Aqui você pode processar os eventos recebidos
	// Por exemplo: mensagens, status, etc.
	switch v := evt.(type) {
	case *events.Message:
		fmt.Printf("📨 Mensagem recebida de %s: %s\n", v.Info.Sender, v.Message.GetConversation())
	case *events.Receipt:
		fmt.Printf("📬 Recibo: Tipo=%s, Chat=%s, IDs=%v\n", v.Type, v.Chat, v.MessageIDs)
	default:
		fmt.Printf("Evento recebido: %T\n", evt)
	}
}
