LaTeX documentation system
==========================


This is work-in-progress and more a prototype than anything useful.

The aim is to have a reference manual for LaTeX accessible via HTTP and command line.

This documentation will have a precise description of all LaTeX kernel commands, environments and the most basic packages and document classes.

Usage
-----

Run `ltxdoc -http :6555` on the command line and point your browser to [http://localhost:6555](http://localhost:6555) to start the built-in webserver.

Run `ltxdoc <commandname>`  such as `ltxdoc section` to get plain text information on that command (not done yet).

How to build
------------

You need a [Go compiler](https://golang.org) and git and perhaps GNU make. That should be enough.

Set the environment variables `$GOPATH` to `$PWD` and `$GOBIN` to `$PWD/bin` and add the bin directory to the `$PATH` variable if necessary:

    export GOPATH=$PWD
    export GOBIN=$PWD/bin
    export PATH=$PATH:$GOBIN

Clone the repository at [https://github.com/speedata/ltxdoc](https://github.com/speedata/ltxdoc) and run either `make getdependencies` or the commands listed in the `Makefile` in that section (`go get -u ...`)

After that run `make install`:

    git clone https://github.com/speedata/ltxdoc.git
    make getdependencies
    make install

this installs a binary in `bin/ltxdoc` which is enough to run the program. You can copy the binary anywhere you want, no extra files are needed.


Information
-----------

Contact: Patrick Gundlach, gundlach@speedata.de<br>
License: free / MIT license<br>
Status: pre-alpha<br>
Sourcecode: https://github.com/speedata/ltxdoc

