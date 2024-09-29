const express = require('express');
const axios = require('axios');
const Web3 = require('web3').default;

// 使用本地节点
const web3 = new Web3('http://127.0.0.1:8545');

const app = express();
const port = 3009;  // API 接口将监听此端口

// 创建一个集合来存储不重复的 peers 信息
const peersSet = new Set();

// 定义函数来获取 peers 并添加到集合中
async function fetchPeers() {
    try {
        // 使用JSON-RPC调用Geth的admin_peers方法
        const response = await axios.post('http://127.0.0.1:8545', {
            jsonrpc: '2.0',
            method: 'admin_peers',
            params: [],
            id: 1
        });

        const peers = response.data.result;

        // 将每个 peer 的 ID 添加到 Set 中（防止重复）
        peers.forEach(peer => {
            if (peer.id) {
                peersSet.add(peer.id);  // 假设每个 peer 都有唯一的 id
            }
        });

        console.log(`Fetched ${peers.length} peers, current unique peers count: ${peersSet.size}`);

    } catch (error) {
        console.error('Failed to fetch peers:', error.message);
    }
}

// 每隔 1 秒获取一次 peers
setInterval(fetchPeers, 10000);

// 定义API接口来返回存储在 Set 中的所有 peers
app.get('/peers', (req, res) => {
    res.json({
        success: true,
        totalPeers: peersSet.size,
        peers: Array.from(peersSet),  // 将 Set 转换为数组进行返回
    });
});

// 启动API服务器
app.listen(port, () => {
    console.log(`API server is running on http://localhost:${port}`);
});
