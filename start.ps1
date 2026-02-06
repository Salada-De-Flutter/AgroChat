# Script para iniciar AgroChat
# Para qualquer processo na porta 8080 e inicia o servidor

Write-Host "[INIT] Iniciando AgroChat..." -ForegroundColor Cyan

# Verificar e parar processo na porta 8080
Write-Host "[CHECK] Verificando porta 8080..." -ForegroundColor Yellow
$connection = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue

if ($connection) {
    $processId = $connection.OwningProcess
    Write-Host "[KILL] Parando processo $processId na porta 8080..." -ForegroundColor Red
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
    Write-Host "[OK] Porta 8080 liberada!" -ForegroundColor Green
} else {
    Write-Host "[OK] Porta 8080 já está livre" -ForegroundColor Green
}

# Iniciar servidor
Write-Host "[START] Iniciando servidor Go..." -ForegroundColor Cyan
go run main.go
