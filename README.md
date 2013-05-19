# kew

Simple HTTP-based queue service

## Getting Started

Install:

    $ make
    $ sudo make install

Run:

    $ kew [/path/to/kew.conf]

## API

    # Enqueue item
    $ curl -X POST http://localhost:5353/queue -d 'foo'

    # Dequeue item
    $ curl -X GET http://localhost:5353/queue/head

    # Get item info
    $ curl -X GET http://localhost:5353/queue/:id

    # Complete item
    $ curl -X DELETE http://localhost:5353/queue/:id