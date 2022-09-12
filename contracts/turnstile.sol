// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

contract Turnstile {
    uint64 private tokenID;

    // UpdateEvent is emitted when a user wants to add new smart contracts
    // to the same cst NFT.
    event Attach(address smartContractAddress, uint64 id);
    // RegisterEvent is emitted when a user wants to create a new CSR nft
    event Register(address smartContractAddress, address receiver, uint64 id);
    // RetroactiveRegisterEvent is emitted when a user wants to retroactively register a smart
    // contract that was previously deployed
    event RetroactiveRegisterEvent(
        address[] contracts,
        address deployer,
        uint64[][] nonces,
        bytes[] salt,
        uint64 id
    );

    constructor() {
        tokenID = 0;
    }

    // register the smart contract to an existing CSR nft
    function register(uint64 id) public {
        tokenID++;
        emit Attach(msg.sender, id);
    }

    // register and mint a new CSR nft that will be transferred to the to address entered
    function register(address to) public {
        tokenID++;
        emit Register(msg.sender, to, tokenID);
    }

    // retroactively register a set of smart contracts to a particular csr nft
    function retroactiveRegister(
        address[] memory contracts,
        address deployer,
        uint64[][] memory nonces,
        bytes[] memory salt,
        uint64 id
    ) public {
        emit RetroactiveRegisterEvent(
            contracts,
            deployer,
            nonces,
            salt,
            id
        );
    }
}
