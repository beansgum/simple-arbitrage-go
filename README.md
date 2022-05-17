simple-arbitrage-go
===================
This project is a Golang version of flashbot's simple-arbitrage. It does a simple cross market arbitrage so don’t expect to make profits from using the code without modifications.

Environment Variables
=====================
- **AVAX_WS_URL** - Avalanche web socket endpoint
- **MEV_PRIVATE** - Private key for the bot wallet(which is also contract executor).

Usage
======================
1. Generate a new bot wallet address and extract the private key into a raw 32-byte format.
2. Deploy the included BundleExecutor.sol to Ethereum, from a secured account, with the address of the newly created wallet as the constructor argument
3. Transfer WAVAX to the newly deployed BundleExecutor