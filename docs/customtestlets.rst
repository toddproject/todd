Custom Testlets
================================

ToDD was originally built with no testlets built-in to the agent. All tests were performed using external executable files (i.e. scripts, binaries) that accept a standard set of input, run a test application, and provide a standard set of output containing metrics from that test. Though ToDD has since adopted a number of testlets to be built-in to the agent for simplicity, this functionality still remains, so you can extend ToDD to run whatever types of tests you wish.

This allows the user to use any testing application (provided it is available on the system on which the ToDD agent is running, and specify which agents run this application. All of the complicated stuff with respect to sending arguments to the underlying testing application as well as parsing the output, is performed inside the testlet.

.. image:: images/testlet.png

The testlet is actually run by the ToDD agent, so if there are 3 agents participating in a test, then 3 testlets are running. All logic that performs the test should be contained within the 

Testrun Definition
------------------

When you want to run a certain testlet, you refer to it by name. There are a number of testlets built-in to ToDD and are therefore reserved:

* http
* bandwidth
* ping

You can, of course, build your own testlet (provided it follows the standard defined on this page) and refer to it by it's filename.

Check Mode
----------
Each testlet must support a "check mode". This is a way of running a testlet that allows the ToDD agent to know whether or not a test can be performed, without actually running the test.

For instance, when the ToDD agent runs the "ping" testlet in check mode, it would invoke it like this:

.. code-block:: text

    ./testletname check

That said, the ToDD Server will distribute testrun instructions to the agents in two phases:

* Install - run the referenced testlet in check mode, and record all params in local agent cache
* Execute - run the installed testrun instruction

Input
-----
There is little to no similarity between various testing applications with respect to the parameters required by those applications. However, in order to simplify things for the ToDD agent, the testlet - due to it's place as a "wrapper" for a testing application - standardizes this input so the ToDD agent can invoke any testlet in a consistent manner

.. code-block:: text

    ./testletname < target > < args >

The ToDD agent will execute the testlet as it exists on the system, and will pass a few arguments to it (meaning the testlet must support and honor these arguments):

* "target" - this is always the first parameter - represents the IP address or FQDN of the target for this test instance.
* "args" - any arguments required by the underlying application. These should be passed to that application via the testlet

Output
------
The output for every testlet is a single-level JSON object, which contains key-value pairs for the metrics gathered for that testlet.

Since the ToDD agent is responsible for executing a testlet, it is also watching stdout for the testlet to provide this JSON object. This is one of the things that make testlets a very flexible method of performing tests - since it only needs to output these metrics as JSON to stdout, the testlet can be written in any language, as long as they support the arguments described in the "Input" section.

A sample JSON object that the "ping" testlet will provide is shown below:

.. code-block:: text

    {
        "avg_latency_ms": "27.007",
        "packet_loss_percentage": "0"
    }

This specific output covers the metrics for a single testlet run, which means that this is relevant to only a single target, run by a single ToDD agent. The ToDD agent will receive this output once for each target in the testrun, and submit this up to the ToDD server for collection.

.. NOTE::
   The ToDD Server will also aggregate each agent's report to a single metric document for the entire testrun, so that it's easy to see the metrics for each source-to-target relationship for a testrun.

The ToDD agent does not have an opinion on the values contained in the keys or values for this JSON object, or how many k/v pairs there are - only that it is valid JSON, and is a single level (no nested objects, lists, etc).