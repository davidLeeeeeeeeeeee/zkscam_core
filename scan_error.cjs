const Web3 = require('web3');

// 连接到以太坊节点，例如本地 Geth 或 Infura
const web3 = new Web3('http://localhost:8545'); // 或使用 Infura: 'https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID'

const address = '0xFe2a7e374320Abe858c21310E533E169236e0F7E'; // 你要查询的地址
const startBlock = 0; // 起始块，可以设置为0或者你希望开始的块
const endBlock = 'latest'; // 结束块，可以设置为latest或者具体块高

// 定义函数来获取交易记录
async function getTransactionsByAccount(account, startBlockNumber, endBlockNumber) {
    if (endBlockNumber === undefined) {
        endBlockNumber = await web3.eth.getBlockNumber();
        console.log("Using endBlockNumber: " + endBlockNumber);
    }

    console.log(`Searching for transactions to/from account "${account}" from block ${startBlockNumber} to ${endBlockNumber}`);

    for (let i = startBlockNumber; i <= endBlockNumber; i++) {
        try {
            let block = await web3.eth.getBlock(i, true);

            if (block != null && block.transactions != null) {
                block.transactions.forEach((tx) => {
                    if (account === tx.from || account === tx.to) {
                        console.log(`\nBlock Number: ${block.number}`);
                        console.log(`Transaction Hash: ${tx.hash}`);
                        console.log(`From: ${tx.from}`);
                        console.log(`To: ${tx.to}`);
                        console.log(`Value: ${web3.utils.fromWei(tx.value, 'ether')} ETH`);
                        console.log(`Gas Used: ${tx.gas}`);
                        console.log(`Gas Price: ${web3.utils.fromWei(tx.gasPrice, 'gwei')} Gwei`);
                    }
                });
            }
        } catch (error) {
            console.error(`Error at block ${i}: ${error.message}`);
        }
    }
}

// 调用函数并输出结果
getTransactionsByAccount(address, startBlock, endBlock)
    .then(() => {
        console.log("Finished searching blocks.");
    })
    .catch((error) => {
        console.error("Error searching transactions:", error);
    });
