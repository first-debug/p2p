# p2p - implementing a peer-to-peer network client on Golnag

## Specification

### Common information about each peer

1) Its ID in UUID format.
2) The port is open for the connections.

There will be more in the future:
- Available files.

### Exploring

#### UDP Multicasting

To use multicast UDP set at least `multicast` filed in the config.

All available params:
``` yaml
explorer:
  multicast:
    address: 235.5.5.11
    port: 8001
    interface-name: wlan0 # system name of interface (`ip a` on Unix or `ipconfig` on Windows)
```

> [!IMPORTANT]
> Peers must use the same multicast UDP addresses and ports to detect each other.

Read more [there](https://en.wikipedia.org/wiki/Multicast).

#### UDP Broadcasting

To use broadcast UDP set at least `broadcast` filed in the config.

All available params:
``` yaml
explorer:
  broadcast:
    address: 255.255.255.255
    port: 8001
```

### Communication

The current implementation uses [WebSocket](https://en.wikipedia.org/wiki/WebSocket) as a communication protocol and [Protobuf](https://protobuf.dev/) to describe unified data representation schemes in binary format.

All peers must listen http trafik on the selected port and have at least 2 endpoits:

1) `/ping` - return `200 OK` if peer ready to communicate else `503 Status Service Unavailable`.
2) `/ws` - requires specifying the `PeerID` in the header, which should identify the client.

#### Http handshake on `/ws` endpoint

1) Client send `GET` request with self ID in the `PeerID` in the header.
2) Server searches for the incoming `PeerID` in its own database of peers.
    * if found send a response with the necessary headers, which are defined by the WebSocket [specification](https://www.rfc-editor.org/rfc/rfc6455.html) and upgrade connection to WebScoket.
    * else return `403 Forbidden`.

#### WebScoket messaging

After upgrading the connection to WebSocker, peers must send and wait for messages that are described by the `Message` message in the [`proto/message.proto`](https://github.com/first-debug/p2p/blob/master/proto/message.proto) file.

