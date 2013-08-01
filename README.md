# queued

Simple HTTP-based queue server

## Getting Started

**Install:**

Ensure Go and LevelDB are installed and then run:

    $ go get
    $ make
    $ sudo make install

**Run:**

    $ queued [options]

## API

**Enqueue:**

    $ curl -X POST http://localhost:5353/:queue -d 'foo'

Append the POSTed data to the end of the specified queue (note that queues are created on-the-fly).  The `Location` header will point to the enqueued item and is of the form `http://localhost:5353/:queue/:id`.

**Dequeue:**

    $ curl -X POST http://localhost:5353/:queue/dequeue

Dequeue the item currently on the head of the queue.  Guaranteed not to return the same item twice unless a completion timeout is specified (see below).  The `Location` header will point to the dequeued item and is of the form `http://localhost:5353/:queue/:id`.  Queued message data is returned in the response body.

Dequeue optionally takes `wait` and/or `timeout` query string parameters:

* `wait=<sec>` - block for the specified number of seconds or until there is an item to be read
off the head of the queue

* `timeout=<sec>` - if the item is not completed (see endpoint below) within the specified number of seconds, the item will automatically be re-enqueued (when no timeout is specified the item is automatically completed when dequeued)

**Info:**

    $ curl -X GET http://localhost:5353/:queue/:id

Get info about the specified item (i.e. whether it is currently dequeued and waiting for completion).

**Complete:**

    $ curl -X DELETE http://localhost:5353/:queue/:id

Complete the specified item and destroy it (note that only items dequeued with a timeout can be completed).

## CLI Options

* **-auth=""** - HTTP basic auth password required for all requests
* **-db-path="./queued.db"** - the directory in which queue items will be persisted
* **-port=5353** - port on which to listen
* **-sync=true** - boolean indicating whether data should be synced to disk after every write (see LevelDB's `WriteOptions::sync`)