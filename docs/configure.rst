Configuring ToDD
================================

ToDD uses configuration files (typically found in /etc/todd) to control it's behavior. The server and the agent use their own individual config files, and sample files are shown below.



First, the server configuration file (usually /etc/todd/server/cfg) contains configurations for all services that the server will require - things like integrated databases, communications plugins, and internal calculations. Here is a sample server configuration file, with comments inline:

Server Configuration
--------------------

.. code-block:: text

    [API]
    Host = 0.0.0.0
    Port = 8080

    # DEPRECATED SOON! (This was the original configuration for comms, but this
    # AMQP section will be condensed under the "Comms" section soon.)
    [AMQP]
    User = guest
    Password = guest
    Host = 192.168.0.10
    Port = 5672

    [Comms]
    Plugin = rabbitmq

    # ToDD's Assets API - allows agents to download assets like collectors, testlets, etc.
    [Assets] 
    IP = 0.0.0.0
    Port = 8090

    [DB]
    IP = 192.168.0.10
    Port = 4001
    Plugin = etcd

    [TSDB]
    IP = 192.168.0.10
    Port = 8086
    Plugin = influxdb

    [Grouping]
    Interval = 10   # Interval (in seconds) for the grouping calculation to run on the server

    [Testing]
    Timeout = 30   # This is the timer (in seconds) that a test will be allowed to live

    [LocalResources]
    DefaultInterface = eth0
    IPAddrOverride = 10.128.0.2 # Normally, the DefaultInterface configuration option is used to get IP address. This overrides that in the event that it doesn't work
    OptDir = /opt/todd/server

Agent Configuration
-------------------

The agent configuration is fairly simpler than the server configuration - largely because the agent doesn't integrate directly with the databases that the server does.

.. code-block:: text

    # DEPRECATED SOON! (This was the original configuration for comms, but this
    # AMQP section will be condensed under the "Comms" section soon.)
    [AMQP]
    User = guest
    Password = guest
    Host = 10.128.0.2
    Port = 5672

    [Comms]
    Plugin = rabbitmq

    [LocalResources]
    DefaultInterface = eth0
    # IPAddrOverride = 192.168.99.100  # Normally, the DefaultInterface configuration option is used to get IP address. This overrides that in the event that it doesn't work
    OptDir = /opt/todd/agent
