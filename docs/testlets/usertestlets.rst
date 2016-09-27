User-Defined Testlets
================================

One of the most important original design principles for ToDD was the ability for users to easily define their own testing. Indeed, this has become one of ToDD's biggest advantages over other testing solutions, both open-source and commercial.

The idea is to allow the user to use any testing application (provided it is available on the system on which the ToDD agent is running. All of the complicated stuff with respect to sending arguments to the underlying testing application as well as parsing the output, is performed inside the testlet.

.. image:: ../images/testlet.png

The testlet is actually run by the ToDD agent, so if there are 3 agents participating in a test, then 3 testlets are running. All logic that performs the test should be contained within the testlet. This is possible because of "the testlet standard", which is a standardized set of input and output that each testlet must support in order to be run by ToDD. This standard is documented in the sections below.

Referring to a Testlet
----------------------

When you want to run a certain testlet, you refer to it by name. There are a number of `testlets built-in to ToDD <nativetestlets/nativetestlets.html>`_ and are therefore reserved:

* http
* bandwidth
* ping
* portknock

Provided it has a unique name, and that it is executable (pre-compiled binary, Python script, bash script, etc.) then it can function as a testlet. Early testlets were actually just bash scripts that wrapped around existing applications like iperf or ping, and simply parsed their output.

.. NOTE::
    All native testlets maintain their own documentation. Please view the links at the top of <nativetestlets/nativetestlets.html>`_ for more information about these testlets, such as what arguments they require, and a sample of their output.


Check Mode
----------
Each testlet must support a "check mode". This is sort of a "pre-test" check to ensure the testlet can be run. When running in "check mode", a testlet should test it's own ability to run a "real" test, such as sending traffic to localhost to ensure it can use the network stack.

For instance, when the ToDD agent runs the "ping" testlet in check mode, it would invoke it like this:

.. code-block:: text

    ping check

That said, the ToDD Server will distribute testrun instructions to the agents in two phases:

* Install - run the referenced testlet in check mode, then record all of the parameters for the intended test in the agent's cache
* Execute - run the installed testrun instruction

Input
-----
Obviously, testing application vary greatly in terms of their input. Some testing applications use certain command-line arguments or flags, and others aren't even configured via the command-line.

The idea of a ToDD testlet is to standardize this input so that any testing application can be run identically. This was a very useful concept early in ToDD's life, as the very first testlets were simple bash scripts that wrapped existing applications like ``ping`` and ``iperf``, and passed around the required arguments to make them conform to the standard we'll discuss now.

All testlets must follow the following standard input:

.. code-block:: text

    ./testletname < target > < args >

* "target" - this is always the first parameter. ``todd-agent`` will spin up N instances of a testlet, where N equals the number of targets that a given agent has been instructed to test against. So, the testlet must accept the target's IP address or hostname as it's first argument.
* "args" - any arguments required by the testlet. These could be arguments required by the testlet itself, or they could be arguments required by an underlying application that the testlet is wrapping. This depends on the testlet implementation.

Output
------
The output for every testlet is a single-level JSON object, which contains key-value pairs representing the metrics gathered for that testlet.

Since the ToDD agent is responsible for executing a testlet, it will watch ``stdout``, which is where ``todd-agent`` will output this JSON object. This makes testlets a very flexible method of performing tests; since it only needs to output these metrics as JSON to stdout, the testlet can be written in any language, as long as they support the input and output standardized in this document.

A sample JSON object that the "ping" testlet will provide is shown below:

.. code-block:: text

    {
        "avg_latency_ms": "27.007",
        "packet_loss_percentage": "0"
    }

.. NOTE::
    The ToDD agent does not have an opinion on the values contained in the keys or values for this JSON object, or how many k/v pairs there are - only that it is valid JSON, and is a single level (no nested objects, lists, etc).

This specific output covers the metrics for a single testlet run, which means that this is relevant to only a single target, run by a single ToDD agent. The ToDD agent will receive this output once for each target in the testrun, and submit this entire dataset up to the ToDD server via the ``comms`` system when finished.

The ToDD Server will also aggregate each agent's report to a single metric document for the entire testrun, so that it's easy to see the metrics for each source-to-target relationship for a testrun.
