GRPC file streaming library
===========================

GRPC imposes a hard limit on request size of 4MB. This quickly becomes problematic when large payloads, typically files
need to be send between 2 GRPC services. To workaround this limitation, it is advised to use GRPC streaming. However,
the logic of splitting the files into small chunks and streaming it to the other services is lef to the user. 

This library handles this logic of for both GRPC `ServerStream` and `ClientStream` objects in a simple and safe way,
making sure the integrity of the file is preserved and failing as early as possible when an error occurs.

### How does it work?
To better explain how this works, it's important to differentiate between the sender and the receiver. The sender is
the one wanting to send a large file to the receiver.
The sender pushes messages to the GRPC stream while the receiver reads them.

Note that the concept of sender VS receiver is independent of the concept of client VS server GRPC has for streaming.
In case of a file upload, the client is the sender and the server the receiver. For a file download, however, it's the
opposite as the one pushing date through the stream is the server.

The library works the following way:
- The sender has the responsibility of splitting the initial payload into fixed-size chunks
- The sender pushes each chunk to the stream. Each chunk is sent alongside a hash of the _previous_ chunk, which was
computed when the previous chunk was processed. For the first chunk, the hash is empty
- The receiver pulls each chunk from the stream
- Each time the receiver recevies a chunk, it computes its hash, which will be compared to the hash received with the
next chunk!
- If the hash differs, it means a chunk was lost in between, and the receivers immediately cancels the transfer.

Here is a sequence diagram summarizing the above steps:
![sequence_diagram.png](img/sequence_diagram.png)

Illustration of a successful transfer:
![successful_transfer.png](img/successful_transfer.png)

Illustration of a failed transfer because the receiver was able to identify that one chunk is missing
![failed_transfer.png](img/failed_transfer.png)

