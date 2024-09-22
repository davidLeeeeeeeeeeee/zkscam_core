
:: Set UTF-8 character encoding to avoid garbled characters
chcp 65001

:: Define the log file path
set LOG_FILE=%~dp0script_output.log

:: Clear previous log file
echo Initializing log file... > "%LOG_FILE%"

:: Initialize blockchain data
echo Initializing blockchain data... >> "%LOG_FILE%"
"%~dp0geth.exe" --datadir "%~dp0data" init "%~dp0zkscam.json" >> "%LOG_FILE%" 2>&1

:: Start Geth node with all configurations passed via command line, enabling HTTP API
echo Starting Geth node... >> "%LOG_FILE%"
start "" cmd /c "%~dp0geth.exe --datadir %~dp0data  --port 30303 --http --http.addr 0.0.0.0 --http.port 8545 --http.api personal,eth,net,web3,txpool,miner --http.corsdomain 'http://localhost' --http.vhosts localhost,127.0.0.1 --networkid 63658 --bootnodes enode://17dbe25597a68936a4ed837ca9a132b740e953ddab4d03d28b51284cdb83a7146a57c7c732ab1b7ca51b024baf3f548b9068051f9c6564c010760906089b5995@103.97.58.18:30303 "

:: Wait for Geth to start and open the HTTP API
echo Waiting for Geth to start and open the HTTP API... >> "%LOG_FILE%"
timeout /t 10 >nul

:: Start mining
:start_mining
:: Read private key file
echo Reading private key file... >> "%LOG_FILE%"
set /p PRIVATE_KEY=<"%~dp0miner_private_key.txt"

:: Use curl to send HTTP request to import the private key
echo Importing private key... >> "%LOG_FILE%"
curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_importRawKey\",\"params\":[\"%PRIVATE_KEY%\", \"your-password\"],\"id\":1}" -H "Content-Type: application/json" http://localhost:8545 > "%~dp0miner_address_response.json" 2>>"%LOG_FILE%"

:: Check private key import response
echo Private key import response: >> "%LOG_FILE%"
type "%~dp0miner_address_response.json" >> "%LOG_FILE%"

:: Extract miner address using findstr and for
for /f "tokens=3 delims=:, \"" %%a in ('findstr /i "\"result\"" "%~dp0miner_address_response.json"') do set "MINER_ADDRESS=%%a"

echo Using extracted miner address: %MINER_ADDRESS% >> "%LOG_FILE%"

:: Unlock account
echo Unlocking account... >> "%LOG_FILE%"
curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"%MINER_ADDRESS%\", \"your-password\", 0],\"id\":1}" -H "Content-Type: application/json" http://localhost:8545 > "%~dp0unlock_account_response.json" 2>>"%LOG_FILE%"

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
