# AgroChat

Bot do WhatsApp desenvolvido em Go usando whatsmeow.

## Pré-requisitos

- Go 1.21 ou superior
- GCC (MinGW no Windows) para compilar o SQLite

## Como executar

```bash
go run main.go
```

Na primeira execução, um QR Code será exibido no terminal. Escaneie com seu WhatsApp:
1. Abra o WhatsApp no celular
2. Vá em Configurações > Aparelhos conectados
3. Toque em "Conectar um aparelho"
4. Escaneie o QR Code exibido no terminal

## Como compilar

```bash
go build -o agrochat.exe
```

## Estrutura do Projeto

```
AgroChat/
├── main.go         # Ponto de entrada da aplicação
├── go.mod          # Gerenciamento de dependências
├── go.sum          # Checksums das dependências
├── agrochat.db     # Banco de dados SQLite (gerado automaticamente)
├── .gitignore      # Arquivos ignorados pelo Git
└── README.md       # Documentação do projeto
```

## Tecnologias

- **whatsmeow**: Cliente WhatsApp Web em Go
- **SQLite**: Armazenamento de sessões
- **qrterminal**: Exibição de QR Code no terminal
