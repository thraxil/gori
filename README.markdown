# Gori

Very simple personal wiki. Built in Go, using Riak for a backend.

Install (assuming a basic Go environment with GOPATH, etc.):

    make install_deps
    make
    sudo make install

make a config file with contents like:

    riak_host = "127.0.0.1:8087" # IP:port for your riak's protocol buffers interface
    port = "8888" # port to run on
    media_dir = "/path/to/media/directory/" where the css/bootstrap stuff lives

then run:

    $ gori -config=/path/to/config.conf

(`make install` should've installed into `/usr/local/bin/` so it
should be in your path).

Then pull up http://localhost:8888/ in your browser and go.
