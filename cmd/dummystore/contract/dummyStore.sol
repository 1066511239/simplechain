pragma solidity 0.5.11;

contract Store {

    mapping (bytes32 => bytes32) public items;

    function set(bytes32 key, bytes32 value) public {
        items[key] = value;
    }
}