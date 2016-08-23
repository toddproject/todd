Testlets
================================

.. toctree::
   :maxdepth: 2

   customtestlets.rst

Testing applications are called "testlets" in ToDD. This is a handy way of referring to "whatever is actually doing the work of testing". This concept keeps things very neatly separated - the testlets focus on performing tests, and ToDD focuses on ensuring that work is distributed as the user directs.

There are a number of testlets that have been developed as part of the ToDD project (referred to as "native testlets"):

* `http <https://github.com/toddproject/todd-nativetestlet-http>`_
* `bandwidth <https://github.com/toddproject/todd-nativetestlet-bandwidth>`_
* `ping <https://github.com/toddproject/todd-nativetestlet-ping>`_
* `portknock <https://github.com/toddproject/todd-nativetestlet-portknock>`_

They run as separate binaries, and are executed in the same way that custom testlets might be executed, if you were to provide one. If you install ToDD using the provided instructions, these are also installed on the system.

.. NOTE::

   If, however, you wish to build your own custom testlets, refer to `Custom Testlets <customtestlets.rst>`_; you'll find it's quite easy to build your own testlets and run them with ToDD. This extensibility was a core design principle of ToDD since the beginning of the project.

If you're not a developer, and/or you just want to USE these native testlets, you can install these binaries anywhere in your PATH. The quickstart instructions illustrate how to unzip the testlets into the right directory. (TODO make sure this is the case)

Developing Native Testlets for ToDD
-------------------------------------

Native Testlets are maintained in their own separate repositories but are distributed alongside ToDD itself. They are also written in Go for a number of reasons. First, it makes it easy for the testlets to honor the testlet format by leveraging some common code in the ToDD repository. However, the testlets are still their own binary. In addition, it allows ToDD to execute tests consistently across platforms (The old model of using bash scripts meant the tests had to be run on a certain platform for which that testlet knew how to parse the output)

The native testlets must be installed somewhere that your PATH environment variable knows about. The typical way to ensure this during development is to just use the Makefile, which kicks off some scripts that perform "go get" commands for the native testlet repositories, and if your GOPATH is set up correctly, the binaries are placed in $GOPATH/bin. Of course, $GOPATH/bin must also be in your PATH variable, which is also a best practice for any Go project.

Need to talk about the native tests you've built in, and turn the "testlets" doc into more of a "so you want to build your own, eh?"

Also need to figure out if you want to refer to both native and non-native as "testlets", or maybe reserve that for non-native

Need a design guide outlining some requirements for native testlets:

* Testlets must honor the "kill" channel passed to the RunTestlet function. If a "true" value is passed into that channel, the testlet must quit immediately.

* Need to put some specifics together regarding testlets that provide some kind of "server" functionality, kind of like what you've done for custom testlets

* How do args work in native testlets? It's a bit awkward to continue to use command-line style args in a native testlet but might be necessary to preserve consistency for the user.

* How to handle vendoring? If a testlet uses a library to run the tests, should the library get vendored with the testlet's repository, or within ToDD proper? Probably the former, but how is it used in that case? Just need to have a strategy here. (You probably vendored such libs in todd proper in order to develop them, so make sure you remove them once they're not needed)

* How are errors returned from the testlet logic? If a testlet returns an error, how is this handled?

* How does development work? Do you clone the testlet repo next to the todd repo, kind of like devstack?