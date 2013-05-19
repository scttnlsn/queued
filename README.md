# kew

Simple HTTP-based queue service

## Getting Started

**Install:**

Ensure Go and LevelDB are installed and then run:

    $ go get
    $ make
    $ sudo make install

**Run:**

    $ kew [/path/to/kew.conf]

## API

**Enqueue:**

    $ curl -X POST http://localhost:5353/:queue -d 'foo'

Append the POSTed data to the end of the specified queue (note that queues are created on-the-fly).

**Dequeue:**

    $ curl -X GET http://localhost:5353/:queue/head

Get the item currently on the head of the queue.  Guaranteed not to return the same item twice unless a completion timeout is specified (see below).  Returns a JSON response of the form:

    { "id": 123, "value": "foo" }

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

## Config

* *port* - (default 5353) the port on which Kew will listen 
* *dbpath* - (default ./kew.db) the directory in which queue items will be persisted 
* *auth* - (default none) an HTTP basic auth password required for all requests 
* *sync* - (default true) boolean indicating whether data should be synced to disk after every write (see LevelDB's `WriteOptions::sync`)