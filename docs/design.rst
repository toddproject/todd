ToDD - High-Level Design
================================

The High-Level Design of ToDD is fairly simple:

.. image:: images/todd-hld.png

Some notes about this:

* All database integrations are at the server level
* Agents communicate with server via "comms" abstraction (currently RMQ is supported)
* Server provides REST API to either the pre-packaged "todd" client tool, or to 3rd party software

ToDD Server
-----------

The ToDD Server has a few particular noteworth attributes:

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
