:: Set UTF-8 character encoding to avoid garbled characters
chcp 65001

:: 删除当前目录下的 data 目录及其所有内容
rmdir /s /q "%~dp0data"

:: Define the log file path
set LOG_FILE=%~dp0script_output.log

:: Clear previous log file
echo Initializing log file... > "%LOG_FILE%"

:: Read private key and miner address from the file
echo Reading private key and miner address... >> "%LOG_FILE%"
set /p PRIVATE_KEY=<"%~dp0miner_private_key.txt"
for /f "skip=1 tokens=*" %%a in (%~dp0miner_private_key.txt) do set MINER_ADDRESS=%%a

echo Private Key: %PRIVATE_KEY% >> "%LOG_FILE%"
echo Miner Address: %MINER_ADDRESS% >> "%LOG_FILE%"

:: Initialize blockchain data
echo Initializing blockchain data... >> "%LOG_FILE%"
"%~dp0geth.exe" --datadir "%~dp0data" init "%~dp0zkscam.json" >> "%LOG_FILE%" 2>&1

:: Start Geth node with all configurations passed via command line, enabling HTTP API
echo Starting Geth node... >> "%LOG_FILE%"
start "" cmd /c "%~dp0geth.exe --datadir %~dp0data --port 30303 --ipcpath "%~dp0geth.ipc" --http --http.addr 0.0.0.0 --http.port 8545 --http.api personal,eth,net,web3,txpool,miner --http.corsdomain '*' --http.vhosts localhost,127.0.0.1 --networkid 63658 --bootnodes enode://17dbe25597a68936a4ed837ca9a132b740e953ddab4d03d28b51284cdb83a7146a57c7c732ab1b7ca51b024baf3f548b9068051f9c6564c010760906089b5995@103.97.58.18:30303 --miner.etherbase %MINER_ADDRESS%" >> "%LOG_FILE%" 2>&1

:: Wait for Geth to start and open the HTTP API
echo Waiting for Geth to start and open the HTTP API... >> "%LOG_FILE%"
timeout /t 10 >nul

:: Unlock account
echo Unlocking account... >> "%LOG_FILE%"
curl --unix-socket "%~dp0geth.ipc" -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"%MINER_ADDRESS%\", \"your-password\", 0],\"id\":1}" -H "Content-Type: application/json" http://localhost > "%~dp0unlock_account_response.json" 2>>"%LOG_FILE%"

:: Check unlock account response
echo Unlock account response: >> "%LOG_FILE%"
type "%~dp0unlock_account_response.json" >> "%LOG_FILE%"

:: Start mining via RPC
echo Starting mining... >> "%LOG_FILE%"
curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"miner_start\",\"params\":[],\"id\":1}" -H "Content-Type: application/json" http://localhost:8545 > "%~dp0start_mining_response.json" 2>>"%LOG_FILE%"

:: Check start mining response
echo Start mining response: >> "%LOG_FILE%"
type "%~dp0start_mining_response.json" >> "%LOG_FILE%"

:: Notify that mining has started
echo Geth is mining with account %MINER_ADDRESS%. >> "%LOG_FILE%"

:: End script, wait for user to press any key to finish
echo Script execution complete. >> "%LOG_FILE%"
pause
