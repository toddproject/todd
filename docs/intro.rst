Introduction
================================

Welcome to the documentation for ToDD! ToDD is an extensible framework for providing natively distributed network testing on demand. ToDD is an acronym that stands for "Testing on Demand: Distributed!". 

Traditionally, the tooling used by network engineers to confirm continued network operation after any kind of change to the network is fairly limited. After a change, a network engineer may run "ping" or "traceroute" from their machine, or perhaps call some application owners to ensure that their apps are still working. Unfortunately, there is a significant difference in network activity between a 3AM change window and peak user activity during the day.

ToDD addresses gaps in today's testing software in three ways:

* Enables real-world traffic distribution for tests using simple grouping primitives
* Provides a open and extensible mechanism for defining tests
* Exposes testing data in an open way, easily allowing for 3rd party software to analyze and visualize

ToDD is a framework through which you can deploy simple test-oriented applications in a distributed manner. With ToDD, you distribute agents around your infrastructure in any place where you feel additional "testing power" is warranted. Then, these agents can be leveraged to mimic real-world network utilization by actually running those same applications at a large scale.

Here are some key features provided by ToDD:

- **Highly Extensible** - ToDD uses a generic interface called testlets for running applications, so that users can easily augment ToDD to support a new application.
- **Post-Test Analytics** - ToDD integrates with time-series databases, such as influxdb. With this, engineers can schedule ToDD test runs to occur periodically, and observe the testrun metrics changing over time.
- **Grouping** - ToDD performs testruns from groups of agents, instead of one specific agent. The user will provide a set of rules that place a given agent into a group (such as hostname, or ip subnet), and ToDD will instruct all agents in that group to perform the test. This means that the power of a test can be increased by simply spinning up additional agents in that group.
- **Diverse Target Types** - Test runs can be configured to target a list of "dumb" targets (targets that are not running a ToDD agent), or a ToDD group. This is useful for certain applications where you need to be able to set up both ends of a test (i.e. setting up a webserver and then testing against it with curl, or setting up an iperf client/server combo)