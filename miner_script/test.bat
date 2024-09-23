:: Set UTF-8 character encoding to avoid garbled characters
::chcp 65001

:: 删除当前目录下的 data 目录及其所有内容
rmdir /s /q "%~dp0data"

:: Read private key and miner address from the file
set /p PRIVATE_KEY=<"%~dp0miner_private_key.txt"
for /f "skip=1 tokens=*" %%a in (%~dp0miner_private_key.txt) do set MINER_ADDRESS=%%a

:: 输出矿工地址到控制台
echo Miner Address: %MINER_ADDRESS%

:: Initialize blockchain data (只输出错误信息)
"%~dp0geth.exe" --datadir "%~dp0data" init "%~dp0zkscam.json"


:: Start Geth node with all configurations passed via command line, enabling HTTP API
start "" cmd /c "%~dp0geth.exe --datadir "%~dp0data" --port 30303 --ipcpath "%~dp0geth.ipc" --http --http.addr 0.0.0.0 --http.port 8545 --allow-insecure-unlock --http.api personal,eth,net,web3,txpool,miner --http.corsdomain '*' --http.vhosts localhost,127.0.0.1 --networkid 63658 --bootnodes enode://17dbe25597a68936a4ed837ca9a132b740e953ddab4d03d28b51284cdb83a7146a57c7c732ab1b7ca51b024baf3f548b9068051f9c6564c010760906089b5995@103.97.58.18:30303 --miner.etherbase %MINER_ADDRESS%"




:: Import private key into Geth account (显示导入结果)
echo Importing private key...
curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_importRawKey\",\"params\":[\"%PRIVATE_KEY%\", \"your-password\"],\"id\":1}" -H "Content-Type: application/json" http://localhost:8545

:: Unlock account (显示解锁结果)
echo Unlocking account...
curl --unix-socket "%~dp0geth.ipc" -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"%MINER_ADDRESS%\", \"your-password\", 0],\"id\":1}" -H "Content-Type: application/json" http://localhost

:: Start mining via RPC (显示挖矿启动结果)
echo Starting mining...
curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"miner_start\",\"params\":[],\"id\":1}" -H "Content-Type: application/json" http://localhost:8545

:: 通知挖矿已开始
echo Geth is mining with account %MINER_ADDRESS%.

:: End script, wait for user to press any key to finish
pause
