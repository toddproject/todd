Native Tests
================================

Need to talk about the native tests you've built in, and turn the "testlets" doc into more of a "so you want to build your own, eh?"

Also need to figure out if you want to refer to both native and non-native as "testlets", or maybe reserve that for non-native



Need a design guide outlining some requirements for native testlets:

* Testlets must honor the "kill" channel passed to the RunTestlet function. If a "true" value is passed into that channel, the testlet must quit immediately.

* Need to put some specifics together regarding testlets that provide some kind of "server" functionality, kind of like what you've done for custom testlets

* How do args work in native testlets? It's a bit awkward to continue to use command-line style args in a native testlet but might be necessary to preserve consistency for the user.

* How to handle vendoring? If a testlet uses a library to run the tests, should the library get vendored with the testlet's repository, or within ToDD proper? Probably the former, but how is it used in that case? Just need to have a strategy here. (You probably vendored such libs in todd proper in order to develop them, so make sure you remove them once they're not needed)

* How are errors returned from the testlet logic? If a testlet returns an error, how is this handled?