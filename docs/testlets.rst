Testlets
================================

Testing applications are referred to "testlets" in ToDD. This is a handy way of referring to "whatever is actually doing the work". ToDD simply orchestrates this work.

There are a number of testlets built-in to the ToDD agent and are usable simply by installing the agent on a system:

* http
* bandwidth
* ping
* portknock

These have their own separate repositories and are distributed alongside ToDD proper. They are written in Go for a number of reasons. First, it makes it easy for the testlets to honor the testlet format by leveraging some common code in the ToDD repository. However, the testlets are still their own binary. In addition, it allows ToDD to execute tests consistently across platforms (The old model of using bash scripts meant the tests had to be run on a certain platform for which that testlet knew how to parse the output)

If you don't want to use any of the built-in testlets, you can, of course, build your own testlet (provided it follows the standard defined on this page) and refer to it by it's filename.

Check Mode
----------
Each testlet must support a "check mode". This is a way of running a testlet that allows the ToDD agent to know whether or not a test can be performed, without actually running the test.

For instance, when the ToDD agent runs the "ping" testlet in check mode, it would invoke it like this:

.. code-block:: text

    ./testletname check

That said, the ToDD Server will distribute testrun instructions to the agents in two phases:

However, please see "Custom Testlets", and you'll find it's quite easy to build your own testlets and run them with ToDD. This extensibility was a core design principle of ToDD since the beginning of the project.


Native Testlet Design Principles
--------------------------------


Need to talk about the native tests you've built in, and turn the "testlets" doc into more of a "so you want to build your own, eh?"

Also need to figure out if you want to refer to both native and non-native as "testlets", or maybe reserve that for non-native


Need a design guide outlining some requirements for native testlets:

* Testlets must honor the "kill" channel passed to the RunTestlet function. If a "true" value is passed into that channel, the testlet must quit immediately.

* Need to put some specifics together regarding testlets that provide some kind of "server" functionality, kind of like what you've done for custom testlets

* How do args work in native testlets? It's a bit awkward to continue to use command-line style args in a native testlet but might be necessary to preserve consistency for the user.

* How to handle vendoring? If a testlet uses a library to run the tests, should the library get vendored with the testlet's repository, or within ToDD proper? Probably the former, but how is it used in that case? Just need to have a strategy here. (You probably vendored such libs in todd proper in order to develop them, so make sure you remove them once they're not needed)

* How are errors returned from the testlet logic? If a testlet returns an error, how is this handled?

* How does development work? Do you clone the testlet repo next to the todd repo, kind of like devstack?