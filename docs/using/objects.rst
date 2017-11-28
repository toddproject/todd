ToDD Objects
================================

There are a number of `concepts <../concepts.html>`_ that power ToDD. If you haven't read that document, please take the time to review that now. In ToDD, the usage of these concepts is defined YAML files. These are imported into ToDD and are generally referred to as "objects". However, each concept is associated with its own object "type".

This document will outline the required syntax for these YAML definitions, as well as the commands needed to use these with ToDD.

All Objects
----------

There are a few components that must be included in every ToDD object definition. These are fields you'll notice regardless of the type of object being defined. They are:

- ``type`` - the type of object being defined (i.e. "group", or "testrun")
- ``label`` - human-readable name for this object so that you can refer to it later
- ``spec`` - the details of the object definition

The "meat" of any ToDD object definition will be found underneath ``spec``, as that is where the type-specific details will be found. For instance, a group definition's matching statements will all appear here.

.. NOTE::

    This page is intended to illustrate the syntax for ToDD objects, not how to work with the ToDD CLI to create, delete, or retrieve objects. For more information on that, please refer to the `CLI reference <cli.html>`_.

Group
----------
To review, a "group" is a collection of agents with similar attributes. The YAML files that define groups can use match statements to identify which agents fall into that group, based on the attributes collected about that node by the ToDD agent.

The core of a group definition is the list of "match" statements. These statements are evaluated one at a time, from first, to last when evaluating an agent. If a statement matches an agent attribute, the rest of the statements are skipped, and the agent becomes part of that group. If no items match, the agent is not part of that group, and other group objects are considered. If no match statements in any group definition matches an agent, the agent remains ungrouped, and therefore cannot be used for testing.

A match statement can allow you to group agents in one of two ways:

- Hostname
- IP Subnet

First, let's look at a group definition that groups three agents via hostname:

.. code-block:: text

    ---
    type: group
    label: group-datacenter
    spec:
        group: datacenter
        matches:
        - hostname: "todd-agent-1"
        - hostname: "todd-agent-2"
        - hostname: "todd-agent-3"

In the above example, only three agents will be placed into this group, as their hostname is being matched directly using one of the three match statements.

However, these ``hostname`` statements can also use regular expressions, to help simplify the group definition by allowing a statement to potentially match multiple agents. The below example is equivalent to the previous definition, but uses regular expression to simplify things:

.. code-block:: text

    ---
    type: group
    label: group-datacenter
    spec:
        group: datacenter
        matches:
        - hostname: "todd-agent-[1-3]"

.. NOTE::

    ToDD's grouping logic uses the `regexp package <https://golang.org/pkg/regexp/#Compile>`_ in Go's standard library to compile each parameter in a ``hostname`` statement, so you should look there for implementation details.

In addition to matching on hostname, group definitions can also match on IP subnet. See the below example for a group definition that includes all agents in the ``192.168.0.0/24`` subnet:

.. code-block:: text

    ---
    type: group
    label: group-datacenter
    spec:
        group: datacenter
        matches:
        - within_subnet: "192.168.0.0/24"


.. NOTE::

    ToDD's grouping logic uses the `net.ParseCIDR <https://golang.org/pkg/net/#ParseCIDR>`_ function to parse the provided subnet, and the
    `net.Contains <https://golang.org/pkg/net/#IPNet.Contains>`_ function to determine if an agent's IP address is within that network. Please refer to the documentation for those functions for details on how they work.

Testrun
----------
A ``testrun`` object defines the parameters for a test. Just like any other ToDD object, they must have a ``type`` field (set to "testrun" in this case), and a ``label`` field. The ``spec`` section contains a few fields that determine how the test will operate.

The first field is called ``targettype``. This determines if the test is being run against a list of "dumb" endpoints, or a group of ToDD agents. This will depend on the testlet being used. This needs to be set to "uncontrolled" if you're just testing against one or more endpoints that aren't running ToDD agents. If your target is another ToDD group, this needs to be set to "group".

The second field is called ``source``, and since all tests **originate** from a ToDD group (even though the destination may or may not be a group), we need to tell ToDD how to instruct the agents in this group to work. So, within the ``source`` configuration, we specify ``name``, which indicates the agent group the test should originate from, ``app``, which is the name of the testlet to use in this group, and ``args``, which is a string of additional command line parameters that may be required by the testlet.

Finally, we also need to provide ``target``. Since ``targettype`` was "uncontrolled", this is a list of IP addresses, or hostnames/FQDNs to test against.

Here's a working example of this kind of testrun with inline comments:

.. code-block:: text

    ---
    type: testrun
    label: test-ping-dns-hq
    spec:
        targettype: uncontrolled    # Is the test being run against
                                    # ToDD agents or "dumb" nodes?
        source:
            name: headquarters      # Which agent group is the "source" for this test? 
            app: ping               # What testlet should this group use for this test?
            args: "-c 10"           # Additional arguments to pass to testlet

        target:                     # Since targettype is "uncontrolled", this
                                    # is a list of IP addresses or FQDNs
        - 4.2.2.2
        - 8.8.8.8

Testruns can also be run against other ToDD groups. For instance, the ``iperf`` testlet requires both a client and a server component. So, in this case, we set ``targettype`` to "group", and instead of a list of IP addresses under ``target``, we instead provide the same three parameters that we did for ``source`` (though the actual values will probably be different between ``source`` and ``target``).

Again, here's a working example of this.

.. code-block:: text

    ---
    type: testrun
    label: test-dc-hq-bandwidth
    spec:
        targettype: group           # Is the test being run against
                                    # ToDD agents or "dumb" nodes?
        source:
            name: datacenter        # Which agent group is the "source" for this test? 
            app: iperf              # What testlet should be used for this test?
            args: "-c {{ target }}" # Additional arguments to pass to testlet
        target:
            name: headquarters      # Which agent group is the "target" for this test? 
            app: iperf              # What testlet should this group use for this test?
            args: "-s"              # Additional arguments to pass to testlet
