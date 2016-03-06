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

You need a [Go compiler](https://golang.org) >= 1.6 and git and GNU make. That should be enough.

Clone the repository at [https://github.com/speedata/ltxdoc](https://github.com/speedata/ltxdoc).

After that run `make install`:

    git clone https://github.com/speedata/ltxdoc.git
    cd ltxdoc
    make install

this installs a binary in `bin/ltxdoc` which is enough to run the program. You can copy the binary anywhere you want, no extra files are needed.


Information
-----------

Contact: Patrick Gundlach, gundlach@speedata.de<br>
License: free / MIT license<br>
Sneak Preview: https://ltxref.org<br>
Status: pre-alpha<br>
Sourcecode: https://github.com/speedata/ltxdoc
