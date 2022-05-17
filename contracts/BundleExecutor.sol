//SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.13;

pragma experimental ABIEncoderV2;

interface IERC20 {
    event Approval(address indexed owner, address indexed spender, uint value);
    event Transfer(address indexed from, address indexed to, uint value);

    function name() external view returns (string memory);
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);
    function totalSupply() external view returns (uint);
    function balanceOf(address owner) external view returns (uint);
    function allowance(address owner, address spender) external view returns (uint);

    function approve(address spender, uint value) external returns (bool);
    function transfer(address to, uint value) external returns (bool);
    function transferFrom(address from, address to, uint value) external returns (bool);
}

interface IWETH is IERC20 {
    function deposit() external payable;
    function withdraw(uint) external;
}

interface IUniswapV2Pair {
    function token0() external view returns (address);
    function token1() external view returns (address);
    function getReserves() external view returns (uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast);
}

abstract contract UniswapV2Factory  {
    mapping(address => mapping(address => address)) public getPair;
    address[] public allPairs;
    function allPairsLength() external view virtual returns (uint);
}

// This contract simply calls multiple targets sequentially, ensuring WETH balance before and after

contract FlashBotsMultiCall {
    address private immutable owner;
    address private immutable executor;
    IWETH private constant WETH = IWETH(0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7); //WAVAX

    modifier onlyExecutor() {
        require(msg.sender == executor || msg.sender == owner);
        _;
    }

    modifier onlyOwner() {
        require(msg.sender == owner);
        _;
    }

    constructor(address _executor) payable {
        owner = msg.sender;
        executor = _executor;
        if (msg.value > 0) {
            WETH.deposit{value: msg.value}();
        }
    }

    receive() external payable {
    }

    function uniswapWeth(uint256 _wethAmountToFirstMarket, address[] memory _targets, bytes[] memory _payloads) external onlyExecutor payable {
        require (_targets.length == _payloads.length);

        uint256 _wethBalanceBefore = WETH.balanceOf(address(this));

        WETH.transfer(_targets[0], _wethAmountToFirstMarket); // transfer funds to first pair

        for (uint256 i = 0; i < _targets.length; i++) {
            (bool _success, bytes memory _response) = _targets[i].call(_payloads[i]); // call swap and send funds to next pair, next pair sends funds back
            require(_success); _response;
        }
 
        uint256 _wethBalanceAfter = WETH.balanceOf(address(this));
        require(_wethBalanceAfter > _wethBalanceBefore);
    }

    function call(address payable _to, uint256 _value, bytes calldata _data) external onlyExecutor payable returns (bytes memory) {
        require(_to != address(0));
        (bool _success, bytes memory _result) = _to.call{value: _value}(_data);
        require(_success);
        return _result;
    }

    function findPairs(UniswapV2Factory _uniswapFactory, address[][] memory pairs) external view returns (address[] memory) {
        address[] memory result = new address[](pairs.length);
    
        for (uint i = 0; i < pairs.length; i++) {
            address[] memory pair = pairs[i];
            address pairAddress = _uniswapFactory.getPair(pair[0], pair[1]);
            result[i] = pairAddress;
        }

        return result;
    }

    function findPairTokens(address[] memory pairAddresses) external view returns (address[][] memory) {
        address[][] memory result = new address[][](pairAddresses.length);

        for (uint i = 0; i < pairAddresses.length; i++) {
            IUniswapV2Pair _uniswapPair = IUniswapV2Pair(pairAddresses[i]);
            address[] memory pairTokens = new address[](2);
            pairTokens[0] = _uniswapPair.token0();
            pairTokens[1] = _uniswapPair.token1();

            result[i] = pairTokens;
        }

        return result;
    }

    function getReservesByPairs(IUniswapV2Pair[] calldata _pairs) external view returns (uint256[3][] memory) {
        uint256[3][] memory result = new uint256[3][](_pairs.length);
        for (uint i = 0; i < _pairs.length; i++) {
            (result[i][0], result[i][1], result[i][2]) = _pairs[i].getReserves();
        }
        return result;
    }

    function getPairsByIndexRange(UniswapV2Factory _uniswapFactory, uint256 _start, uint256 _stop) external view returns (address[3][] memory)  {
        uint256 _allPairsLength = _uniswapFactory.allPairsLength();
        if (_stop > _allPairsLength) {
            _stop = _allPairsLength;
        }
        require(_stop >= _start, "start cannot be higher than stop");
        uint256 _qty = _stop - _start;
        address[3][] memory result = new address[3][](_qty);
        for (uint i = 0; i < _qty; i++) {
            IUniswapV2Pair _uniswapPair = IUniswapV2Pair(_uniswapFactory.allPairs(_start + i));
            result[i][0] = _uniswapPair.token0();
            result[i][1] = _uniswapPair.token1();
            result[i][2] = address(_uniswapPair);
        }
        return result;
    }
}
