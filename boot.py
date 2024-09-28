import os
import subprocess
import time
import json
import requests
import shutil
import sys

def get_script_dir():
    """获取脚本所在目录。如果是打包后的环境，则返回临时解压目录的路径。"""
    if getattr(sys, 'frozen', False):
        # 如果程序是打包后运行的，使用这个路径
        return os.path.dirname(sys.executable)
    else:
        # 如果不是打包后运行的，使用当前文件的目录
        return os.path.dirname(os.path.abspath(__file__))

def main():
    # 获取脚本所在目录 / Get the script directory
    script_dir = get_script_dir()

    # 设置日志文件路径 / Set log file path
    log_file = os.path.join(script_dir, 'log.txt')

    # 定义其他文件路径 / Define other file paths
    data_dir = os.path.join(script_dir, 'data')
    geth_path = os.path.join(script_dir, 'geth.exe')
    zkscam_json = os.path.join(script_dir, 'zkscam.json')
    private_key_file = os.path.join(script_dir, 'miner_private_key.txt')
    importing_status = os.path.join(script_dir, 'Importing_status.json')
    unlocking_status = os.path.join(script_dir, 'Unlocking_status.json')
    miner_status = os.path.join(script_dir, 'miner_status.json')

    # 定义日志记录函数 / Define logging function
    def log(message_cn, message_en):
        combined_message = f"{message_cn} | {message_en}"
        with open(log_file, 'a', encoding='utf-8') as f:
            f.write(combined_message + '\n')
        print(combined_message)

    # 清理之前的日志文件 / Clear previous log file
    with open(log_file, 'w', encoding='utf-8') as f:
        f.write('初始化日志文件... | Initializing log file...\n')

    # 确保剩下的代码逻辑不变
    # 删除 data 目录及其所有内容 / Remove the data directory and all its contents
    try:
        if os.path.exists(data_dir):
            shutil.rmtree(data_dir)
            log('成功删除 data 目录。', 'Successfully deleted the data directory.')
        else:
            log('data 目录不存在，跳过删除。', 'Data directory does not exist, skipping deletion.')
    except Exception as e:
        log(f'删除 data 目录失败: {e}', f'Failed to delete data directory: {e}')
        sys.exit(1)

    # 读取私钥和矿工地址 / Read private key and miner address
    try:
        with open(private_key_file, 'r', encoding='utf-8') as f:
            lines = f.readlines()
            if len(lines) < 2:
                log('私钥文件格式错误，至少需要两行：私钥和矿工地址。',
                    'Private key file format error, at least two lines required: private key and miner address.')
                sys.exit(1)
            private_key = lines[0].strip()
            miner_address = lines[1].strip()
    except Exception as e:
        log(f'读取私钥文件失败: {e}', f'Failed to read private key file: {e}')
        sys.exit(1)

    # 输出矿工地址到控制台 / Output miner address to console
    log(f'Miner Address: {miner_address}', f'Miner Address: {miner_address}')

    # 初始化区块链数据 / Initialize blockchain data
    log('初始化区块链数据...', 'Initializing blockchain data...')
    try:
        init_cmd = [geth_path, '--datadir', data_dir, 'init', zkscam_json]
        subprocess.run(init_cmd, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        log('区块链数据初始化完成。', 'Blockchain data initialization completed.')
    except subprocess.CalledProcessError as e:
        log(f'初始化区块链数据失败: {e.stderr.decode("utf-8")}', f'Failed to initialize blockchain data: {e.stderr.decode("utf-8")}')
        sys.exit(1)
    time.sleep(3)
    geth_process = None  # 保存 Geth 进程句柄 / Save Geth process handle
    # 启动 Geth 节点并在新的控制台窗口中显示 / Start Geth node and display in a new console window
    log('启动 Geth 节点...', 'Starting Geth node...')
    geth_command = [
        geth_path,
        '--datadir', data_dir,
        '--port', '30303',
        '--ipcpath', os.path.join(script_dir, 'geth.ipc'),
        '--http',
        '--syncmode', 'full',
        '--http.addr', '0.0.0.0',
        '--http.port', '8545',
        '--allow-insecure-unlock',
        '--http.api', 'personal,eth,net,web3,txpool,miner,admin',
        '--http.corsdomain', '*',
        '--networkid', '63658',
        '--bootnodes',
        'enode://8d8fcc2f81bb0f6a653b3e71f8ce31c1227ab39fb8a1a3fe6008521767273e29054019bcf933e3a4954131c56790aaef0aff8251fe4c389dae3380483e2576df@103.97.58.18:30303',
        '--miner.etherbase', miner_address,
        'console'
    ]

    try:
        geth_process = subprocess.Popen(
            geth_command,
            cwd=script_dir,
            creationflags=subprocess.CREATE_NEW_CONSOLE
        )
        log('Geth 节点已启动，并在新的控制台窗口中显示。', 'Geth node has been started and is displayed in a new console window.')
    except Exception as e:
        log(f'启动 Geth 节点失败: {e}', f'Failed to start Geth node: {e}')
        sys.exit(1)

    # 等待 Geth 启动 / Wait for Geth to start
    log('等待 Geth 启动和打开 HTTP API...', 'Waiting for Geth to start and open HTTP API...')
    time.sleep(10)

    # 导入私钥 / Import private key
    log('导入私钥...', 'Importing private key...')
    try:
        import_payload = {
            "jsonrpc": "2.0",
            "method": "personal_importRawKey",
            "params": [private_key, private_key],
            "id": 1
        }
        response = requests.post('http://localhost:8545', json=import_payload)
        with open(importing_status, 'w', encoding='utf-8') as f:
            json.dump(response.json(), f, indent=4)
        log(f'导入私钥响应: {response.json()}', f'Import raw key response: {response.json()}')
    except Exception as e:
        log(f'导入私钥失败，请检查你的私钥与地址是否正确，文件是否保存:: {e}', f'Failed to import private key. Please check if your private key and address are correct and the file is saved: {e}')
        if geth_process:
            geth_process.terminate()  # 终止 Geth 进程 / Terminate Geth process
            log('Geth 进程已终止。', 'Geth process terminated.')
        time.sleep(60)
        sys.exit(1)

    # 解锁账户 / Unlock account
    log('解锁账户...', 'Unlocking account...')
    try:
        unlock_payload = {
            "jsonrpc": "2.0",
            "method": "personal_unlockAccount",
            "params": [miner_address, private_key, 0],
            "id": 1
        }
        response = requests.post('http://localhost:8545', json=unlock_payload)
        response_json = response.json()

        # 解析解锁账户的响应 / Parse the unlock account response
        if 'result' in response_json and response_json['result'] is True:
            with open(unlocking_status, 'w', encoding='utf-8') as f:
                json.dump(response_json, f, indent=4)
            log(f'解锁账户成功: {response_json}', f'Account successfully unlocked: {response_json}')
        else:
            log(f'解锁账户失败: {response_json}', f'Failed to unlock account: {response_json}')
            log(f'解锁账户失败, 请检查你的私钥与地址是否正确: ',
                f'Failed to unlock account, please check if your private key and address are correct: ')
            if geth_process:
                geth_process.terminate()  # 终止 Geth 进程 / Terminate Geth process
                log('Geth 进程已终止。', 'Geth process terminated.')
            time.sleep(60)
            sys.exit(1)  # 终止程序 / Terminate the script if unlocking fails
    except Exception as e:
        log(f'解锁账户失败, 请检查你的私钥与地址是否正确: {e}',
            f'Failed to unlock account, please check if your private key and address are correct: {e}')
        if geth_process:
            geth_process.terminate()  # 终止 Geth 进程 / Terminate Geth process
            log('Geth 进程已终止。', 'Geth process terminated.')
        time.sleep(60)

        sys.exit(1)

    # 添加同步检查循环 / Add synchronization check loop
    log('检查同步状态...', 'Checking synchronization status...')
    while True:
        try:
            # 检查同步状态 / Check synchronization status
            sync_payload = {
                "jsonrpc": "2.0",
                "method": "eth_syncing",
                "params": [],
                "id": 1
            }
            sync_response = requests.post('http://localhost:8545', json=sync_payload)
            sync_result = sync_response.json()

            # 获取对等节点数 / Get peer count
            peer_payload = {
                "jsonrpc": "2.0",
                "method": "net_peerCount",
                "params": [],
                "id": 1
            }
            peer_response = requests.post('http://localhost:8545', json=peer_payload)
            peer_result = peer_response.json()

            if 'result' in sync_result:
                if isinstance(sync_result['result'], bool) and not sync_result['result']:
                    log('同步已完成。 | Synchronization complete.', 'Synchronization complete.')
                    break
                else:
                    # 从16进制转换为10进制 / Convert from hex to decimal
                    current_block_hex = sync_result['result'].get('currentBlock', '0x0')
                    highest_block_hex = sync_result['result'].get('highestBlock', '0x0')
                    current_block = int(current_block_hex, 16)
                    highest_block = int(highest_block_hex, 16)

                    # 获取对等节点数并转换为10进制 / Get peer count and convert to decimal
                    peer_count_hex = peer_result.get('result', '0x0')
                    peer_count = int(peer_count_hex, 16)

                    log(
                        f'同步中... 当前块: {current_block}, 最高块: {highest_block}, 对等节点数: {peer_count}...',
                        f'Synchronizing... Current block: {current_block}, Highest block: {highest_block}, Peer count: {peer_count}...'
                    )
            else:
                log('无法获取同步状态。 | Unable to retrieve synchronization status.', 'Unable to retrieve synchronization status.')
        except Exception as e:
            log(f'检查同步状态失败: {e}', f'Failed to check synchronization status: {e}')
        time.sleep(10)

    # 启动挖矿 / Start mining
    log('启动挖矿...', 'Starting mining...')
    try:
        mining_payload = {
            "jsonrpc": "2.0",
            "method": "miner_start",
            "params": [],
            "id": 1
        }
        response = requests.post('http://localhost:8545', json=mining_payload)
        with open(miner_status, 'w', encoding='utf-8') as f:
            json.dump(response.json(), f, indent=4)
        log(f'启动挖矿响应: {response.json()}', f'Mining start response: {response.json()}')
    except Exception as e:
        log(f'启动挖矿失败,请截图保存报告David: {e}',
            f'Failed to start mining. Please take a screenshot and report to David: {e}')
        time.sleep(60)
        if geth_process:
            geth_process.terminate()  # 终止 Geth 进程 / Terminate Geth process
            log('Geth 进程已终止。', 'Geth process terminated.')
        sys.exit(1)

    # 挖矿成功统计 / Mining success counter
    blocks_mined = 0

    # 定期检查区块信息 / Periodically check block information
    log(f'Geth 正在使用账户 {miner_address} 挖矿。', f'Geth is mining using account {miner_address}.')
    log('定期检查区块信息...', 'Periodically checking block information...')

    while True:
        try:
            # 请求最新区块信息 / Request latest block information
            block_number_payload = {
                "jsonrpc": "2.0",
                "method": "eth_blockNumber",
                "params": [],
                "id": 1
            }
            block_response = requests.post('http://localhost:8545', json=block_number_payload)
            block_number = int(block_response.json().get('result', '0x0'), 16)

            # 请求最新区块详情 / Request latest block details
            block_details_payload = {
                "jsonrpc": "2.0",
                "method": "eth_getBlockByNumber",
                "params": [hex(block_number), True],
                "id": 1
            }
            block_details_response = requests.post('http://localhost:8545', json=block_details_payload)
            block_details = block_details_response.json().get('result', {})

            # 检查矿工地址 / Check miner addresses
            miner_addresses = block_details.get('minerAddresses', [])
            if miner_address.lower() in miner_addresses:
                blocks_mined += 1
                log(
                    f'挖矿成功！账户 {miner_address} 已经成功挖出 {blocks_mined} 个区块。',
                    f'Mining successful! Account {miner_address} has successfully mined {blocks_mined} blocks.'
                )
            else:
                log('还没开始挖矿。', 'Mining has not started yet.')

        except Exception as e:
            log(f'检查区块信息失败: {e}', f'Failed to check block information: {e}')

        # 每隔 30 秒检查一次 / Check every 30 seconds
        time.sleep(30)

if __name__ == '__main__':
    main()
