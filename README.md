# Demonstration Application "FunnyCounter"


This is a demonstration application for Multiverse, based on the counter example application of tendermint.
The ultimate aim of this application is to demonstrate the possibility to update a live blockchain with new behavior from Multiverse (functionality not yet implemented at the moment).

This application counts the number of transaction that happened on the blockchain, and rewards the node who registered the transaction by a fixed amount.

It consist of two executable :
* `abci_counter` the NodeApp the tendermint node should connect to
* `demo_monitor` a small CLI UI to use this FunnyCounter


## Installation

```bash
go get github.com/MultiverseHQ/demo_app/abci_counter
go get github.com/MultiverseHQ/demo_app/demo_monitor
```

## Use the monitor

The monitor just require the address of one of the tendermint node of the blockchain that you want to use.

```bash
demo_monitor -a <multiverse_node_address> -p <port_of_the_node>
```

You can press `q` to exit the monitor, or `t`to trigger a valid transaction on one of the node. A summary of the status of all the node is updated every 500ms.

Recorded demosntration :
https://asciinema.org/a/xauMvIvm9UWY6YelpvQskcllS
