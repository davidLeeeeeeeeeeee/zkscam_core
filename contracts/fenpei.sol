// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// 导入OpenZeppelin的合约库
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract ERC20DepositOnlyETHDistributor is ReentrancyGuard {
    using SafeERC20 for IERC20;

    // 状态变量
    IERC20 public immutable stakingToken; // 用户存入的ERC20代币
    uint256 public immutable distributionInterval = 24 hours; // 分配间隔时间

    uint256 public lastDistributionTime; // 上一次分配ETH的时间
    uint256 public rewardPerTokenStored; // 每代币累计的奖励
    uint256 public totalStaked; // 总存入的ERC20代币数量

    uint256 public lastBalance; // 上一次分配时的合约ETH余额

    mapping(address => uint256) public stakedBalances; // 每个用户的存入量
    mapping(address => uint256) public userRewardPerTokenPaid; // 用户已领取的每代币奖励
    mapping(address => uint256) public rewards; // 用户的累计奖励

    // 事件
    event Staked(address indexed user, uint256 amount);
    event RewardDistributed(uint256 amount);
    event RewardClaimed(address indexed user, uint256 reward);
    event ETHReceived(address indexed sender, uint256 amount);

    // 构造函数
    constructor(address _stakingToken) {
        stakingToken = IERC20(_stakingToken);
        lastDistributionTime = block.timestamp;
        lastBalance = address(this).balance;
    }

    // 修饰器：在执行函数前更新奖励
    modifier updateReward(address account) {
        _distributeRewards(); // 首先分配任何新增的ETH奖励
        if (account != address(0)) {
            rewards[account] = earned(account); // 更新用户的累计奖励
            userRewardPerTokenPaid[account] = rewardPerTokenStored; // 更新用户已领取的每代币奖励
        }
        _;
    }

    // 用户存入ERC20代币
    function stake(uint256 _amount) external nonReentrant updateReward(msg.sender) {
        require(_amount > 0, "Cannot stake 0");
        stakingToken.safeTransferFrom(msg.sender, address(this), _amount);
        stakedBalances[msg.sender] += _amount;
        totalStaked += _amount;
        emit Staked(msg.sender, _amount);
    }

    // 用户领取ETH奖励
    function claimRewards() external nonReentrant updateReward(msg.sender) {
        uint256 reward = rewards[msg.sender];
        require(reward > 0, "No rewards");
        rewards[msg.sender] = 0;
        (bool success, ) = msg.sender.call{value: reward}("");
        require(success, "ETH transfer failed");
        emit RewardClaimed(msg.sender, reward);
    }


    // 内部函数：分配奖励
    function _distributeRewards() internal {
        uint256 currentTime = block.timestamp;
        if (currentTime >= lastDistributionTime + distributionInterval) {
            uint256 periods = (currentTime - lastDistributionTime) / distributionInterval;
            uint256 currentBalance = address(this).balance;
            uint256 newRewards = currentBalance > lastBalance ? currentBalance - lastBalance : 0;

            if (newRewards > 0 && totalStaked > 0) {
                // 计算每代币的奖励，使用1e18作为精度
                rewardPerTokenStored += (newRewards * 1e18) / totalStaked;
                emit RewardDistributed(newRewards);
                lastBalance += newRewards;
                totalStaked = 0;
            }

            // 更新最后分配时间
            lastDistributionTime += periods * distributionInterval;
        }
    }

    // 计算当前的每代币奖励
    function rewardPerToken() public view returns (uint256) {
        return rewardPerTokenStored;
    }

    // 计算用户的累计奖励
    function earned(address account) public view returns (uint256) {
        return
            (stakedBalances[account] * (rewardPerToken() - userRewardPerTokenPaid[account])) /
            1e18 +
            rewards[account];
    }
    receive() external payable {
        emit ETHReceived(msg.sender, msg.value);
    }
}
