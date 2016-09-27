Native Testlets
================================
There are a number of testlets that are developed as part of the overall ToDD Project. This was done to provide immediate functionality for the vast majority of users that are looking for simple tests, that are useful to anyone. Things like ping, basic HTTP testing, and bandwidth testing are things that any user can take advantage of. These testlets are called "native testlets".

Native Testlets are maintained in their own separate repositories but are distributed alongside ToDD itself.

.. toctree::
   :maxdepth: 1

   ping.rst
   http.rst
   bandwidth.rst
   portknock.rst

They are also written in Go for a number of reasons:

* Each testlet can leverage some common code in the ToDD repository, to keep things simple. However, each testlet is provided as a separate binary (runs as a separate process to the todd-server or todd-agent)
* Testlets can be executed consistently across different platforms. The old model of using bash scripts meant the tests had to be run on a certain platform for which that testlet knew how to parse the output

The native testlets must be installed in a location that has been added to the PATH environment variable.  If you are building rom source, the included Makefile will kick off some scripts that perform "go get" commands for the native testlet repositories, and if your GOPATH is set up correctly, the binaries are placed in $GOPATH/bin. Of course, `$GOPATH/bin` must also be in your PATH variable, which is also a best practice for any Go project. In the future, additional installation methods should do this for you by placing all binaries in sensible locations like ``/usr/local/bin``.

If these testlets do not meet your needs, please check out the documentation for `User-Defined Testlets <../usertestlets.html>`_. One of the most important design requirements for ToDD was to allow for easy introduction of user-defined testing.
