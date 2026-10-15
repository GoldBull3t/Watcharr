@echo off
title Watcharr Launcher
echo =========================================
echo       Iniciando Watcharr sem Docker
echo =========================================
echo.

echo 1/2. Iniciando Backend em Go (Porta 3080)...
start "Watcharr Backend" powershell -NoExit -Command "$env:Path += ';C:\Program Files\Go\bin'; $env:MODE = 'DEV'; cd server; go run ."

echo 2/2. Iniciando Frontend Vite (Porta 5173)...
start "Watcharr Frontend" cmd /c "npm run dev"

echo.
echo Aguardando inicializacao dos servicos...
timeout /t 3 >nul

echo Abrindo o navegador em http://localhost:5173 ...
start http://localhost:5173

echo.
echo =========================================
echo Pronto! Os dois terminais estao rodando.
echo Para fechar, basta fechar as duas janelas.
echo =========================================

