# AgroChat

Bot do WhatsApp desenvolvido em Go usando whatsmeow com API REST para integração.

## Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 12 ou superior
- GCC (MinGW no Windows) - opcional, apenas se usar SQLite

## Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` com as seguintes variáveis:

```env
DB_HOST=flutterbox
DB_USER=flutter
DB_PASSWORD=4002
DB_NAME=AgroChatDB
DB_PORT=5432

API_PORT=8080
```

### Banco de Dados

Configure o PostgreSQL e crie o database:

```sql
CREATE DATABASE AgroChatDB;
```

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

## Documentação da API

Após iniciar o servidor, acesse a documentação interativa do Swagger:

🔗 **http://localhost:8080/swagger/index.html**

### Endpoints Disponíveis

#### **GET** `/health`
Verifica se a API está funcionando

#### **GET** `/whatsapp/status`
Retorna o status da conexão com o WhatsApp

#### **POST** `/whatsapp/send`
Envia mensagem via WhatsApp (endpoint antigo com verificações detalhadas)

**Request Body:**
```json
{
  "phone": "5588992422814",
  "message": "Sua mensagem aqui"
}
```

#### **POST** `/enviar-mensagem`
Envia mensagem simples via WhatsApp

**Request Body:**
```json
{
  "numero": "88992422814",
  "mensagem": "Sua mensagem aqui"
}
```

**Response:**
```json
{
  "sucesso": true,
  "mensagem": "Mensagem enviada com sucesso",
  "timestamp": "2026-02-06T13:30:00Z"
}
```

#### **POST** `/enviar-verificacao`
Envia código de verificação formatado para clientes do AgroServer

**Request Body:**
```json
{
  "nomeCliente": "João Silva",
  "nomeVendedor": "Maria Santos",
  "documento": "12345678900",
  "telefone": "88992422814",
  "endereco": "Rua das Flores, 123",
  "codigoVerificacao": "123456",
  "metodo": "whatsapp"
}
```

**Response:**
```json
{
  "sucesso": true,
  "mensagem": "Mensagem enviada com sucesso",
  "idMensagem": "msg_abc12345",
  "dataEnvio": "2026-02-06T13:30:00Z"
}
```

## Estrutura do Projeto

```
AgroChat/
├── api/
│   └── api.go          # Rotas e handlers da API REST
├── docs/               # Documentação Swagger (gerada automaticamente)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── main.go             # Ponto de entrada da aplicação
├── go.mod              # Gerenciamento de dependências
├── go.sum              # Checksums das dependências
├── agrochat.db         # Banco de dados SQLite (se usado)
├── .env                # Variáveis de ambiente
├── .gitignore          # Arquivos ignorados pelo Git
├── start.ps1           # Script de inicialização (PowerShell)
└── README.md           # Documentação do projeto
```

## Tecnologias

- **whatsmeow**: Cliente WhatsApp Web em Go
- **PostgreSQL**: Armazenamento de sessões e dados
- **Gin**: Framework web para API REST
- **Swagger/OpenAPI**: Documentação interativa da API
- **qrterminal**: Exibição de QR Code no terminal

## Formatação de Números

O sistema detecta e formata automaticamente números brasileiros:

- **Entrada:** `88992422814` ou `(88) 99242-2814`
- **Processamento:** Adiciona código do país `55` se necessário
- **Correção:** Remove o 9º dígito de celulares (formato antigo)
- **Saída:** `558892422814@s.whatsapp.net`

## Deploy

### Desenvolvimento
```bash
# AgroChat rodando em:
localhost:8080

# AgroServer rodando em:
localhost:3000
```

### Produção
```bash
# AgroChat (interno):
http://agrochat-service:8080

# AgroServer (público):
https://api.agrosystemapp.com
```

## Atualizar Documentação Swagger

Após modificar as anotações nos comentários do código, execute:

```bash
swag init
```

Isso regerará os arquivos em `docs/`.

## Logs

Os logs são exibidos no formato:

```
[CHECK] Verificando porta 8080...
[OK] Porta 8080 livre
[VERIFICATION] ===== ENVIO DE VERIFICAÇÃO =====
[CLIENT] Cliente: João Silva
[VENDOR] Vendedor: Maria Santos
[PHONE] Telefone: 88992422814 → 558892422814@s.whatsapp.net
[CODE] Código: 123456
===============================================
[SUCCESS] Verificação enviada com sucesso!
```

## Contato

Para dúvidas sobre a integração:
- **Documentação Swagger:** http://localhost:8080/swagger/index.html
- **Endpoint AgroServer:** http://localhost:3000/api-docs

---

**Versão:** 1.0.0  
**Data:** 06/02/2026  
**Autor:** AgroChat Team
