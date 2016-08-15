Testlets
================================

Testing applications are referred to "testlets" in ToDD. This is a handy way of referring to "whatever is actually doing the work". ToDD simply orchestrates this work.

There are a number of testlets built-in to the ToDD agent and are usable simply by installing the agent on a system:

* http
* bandwidth
* ping

However, please see "Custom Testlets", and you'll find it's quite easy to build your own testlets and run them with ToDD. This extensibility was a core design principle of ToDD since the beginning of the project.


Native Testlet Design Principles
--------------------------------

Need a design guide outlining some requirements for native testlets:

* Testlets must honor the "kill" channel passed to the RunTestlet function. If a "true" value is passed into that channel, the testlet must quit immediately.

* Need to put some specifics together regarding testlets that provide some kind of "server" functionality, kind of like what you've done for custom testlets

* How do args work in native testlets? It's a bit awkward to continue to use command-line style args in a native testlet but might be necessary to preserve consistency for the user.

* How to handle vendoring? If a testlet uses a library to run the tests, should the library get vendored with the testlet's repository, or within ToDD proper? Probably the former, but how is it used in that case? Just need to have a strategy here. (You probably vendored such libs in todd proper in order to develop them, so make sure you remove them once they're not needed)

* How are errors returned from the testlet logic? If a testlet returns an error, how is this handled?

* How does development work? Do you clone the testlet repo next to the todd repo, kind of like devstack?