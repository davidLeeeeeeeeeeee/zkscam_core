:: Set UTF-8 character encoding to avoid garbled characters
::chcp 65001

:: 删除当前目录下的 data 目录及其所有内容
rmdir /s /q "%~dp0data"
:: 设置日志文件路径
set "LOG_FILE=%~dp0log.txt"

:: Temporary file to store output
set "tempfile=%~dp0peers_status.json"
:: Read private key and miner address from the file
set /p PRIVATE_KEY=<"%~dp0miner_private_key.txt"
for /f "skip=1 tokens=*" %%a in (%~dp0miner_private_key.txt) do set MINER_ADDRESS=%%a

:: 输出矿工地址到控制台
echo Miner Address: %MINER_ADDRESS%

:: Initialize blockchain data (只输出错误信息)
"%~dp0geth.exe" --datadir "%~dp0data" init "%~dp0zkscam.json"

:: Start Geth node with all configurations passed via command line, enabling HTTP API
start "" cmd /c "%~dp0geth.exe --datadir "%~dp0data" --port 30303 --ipcpath "%~dp0geth.ipc"   --http --syncmode full --http.addr 0.0.0.0 --http.port 8545 --allow-insecure-unlock --http.api personal,eth,net,web3,txpool,miner,admin --http.corsdomain '*'  --networkid 63658 --bootnodes enode://6083b79f6f1375dcb52663301facec92c23b9a05f426c5b208775b701ed4c8effb49d6855fce9cc60430bd88fc7ec0efa97967f82e33b6b9f95b3ce96a1aae2f@103.97.58.18:30303 --miner.etherbase %MINER_ADDRESS%  console"


::需要管理员权限运行才不会报错
timeout /t 10
:: Import private key into Geth account (显示导入结果)
echo Importing private key...
curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_importRawKey\",\"params\":[\"%PRIVATE_KEY%\", \"123\"],\"id\":1}" -H "Content-Type: application/json" http://localhost:8545> "%~dp0Importing_status.json" 2>>"%LOG_FILE%"

:: Unlock account (显示解锁结果)
echo Unlocking account...
curl --unix-socket "%~dp0geth.ipc" -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"%MINER_ADDRESS%\", \"123\", 0],\"id\":1}" -H "Content-Type: application/json" http://localhost> "%~dp0Unlocking_status.json" 2>>"%LOG_FILE%"



:: Delete status files
::del "%~dp0sync_status.json"
::del "%~dp0peers_status.json"

:: Start mining via RPC (显示挖矿启动结果)
echo Starting mining...
curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"miner_start\",\"params\":[],\"id\":1}" -H "Content-Type: application/json" http://localhost:8545 > "%~dp0miner_status.json" 2>>"%LOG_FILE%"

:: 通知挖矿已开始
echo Geth is mining with account %MINER_ADDRESS%.

:: End script, wait for user to press any key to finish
pause
