# GoAkt Bank Cluster Example

A minimal example demonstrating **clustering** and **cross-node messaging** with [GoAkt v4](https://github.com/Tochemey/goakt). It shows how to:

- Start an actor system in **cluster mode** with remoting enabled
- Register actor kinds for cluster-wide placement
- Resolve actors by name across nodes using location transparency
- Send messages between nodes (`Ask` / `Tell`)