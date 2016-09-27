High-Level Design
================================

The High-Level Design of ToDD is fairly simple. ToDD is composed of three components:

* Server
* Agent
* Client

A general idea of how these components is depicted below:

.. image:: images/todd-hld.png

Some notes about this:

* All database integrations are at the server level - no agent communicates with database
* Agents **do not** communicate directly with server. This is done through some kind of message queue (i.e. RabbitMQ) using ToDD's "comms" abstraction.
* Server has a REST API built-in. No other software needed (see section "ToDD Server" for more)

The following sections elaborate on each component in greater detail.

ToDD Server
-----------

The ToDD Server has a few particular noteworthy attributes:

* Orchestrates test runs between groups of agents (not an endpoint for any testing)
* Manages agent registration and remediation
* Interacts with databases
* Manages group topology
* Provides HTTP API to the ToDD client and 3rd party software

ToDD Agent
-----------

The Agent is actually where the test is run. The specific tasks that a particular agent must perform are calculated by the server in order to facilitate the "big picture" of the overall system test, and pushed down to the agent via the Comms abstraction.

Here are some specific notes on the agents.

* Provides facts about operating environment back to server
* Receives and executes testrun instructions from server
* Variety of form factors (baremetal, VM, container, RasPi, network switch)

ToDD Client
-----------

The ToDD client is provided via the "todd" shell command. Several subcommands are available here, such as "todd agents" to print the list of currently registered agents.

* Manages installed ToDD objects (group and testrun definitions, etc)
* Queries state of ToDD infrastructure ("todd agents", "todd groups", etc.)
* Executes testruns ("todd run ...")
