Configuring ToDD
================================

ToDD uses configuration files (typically found in ``/etc/todd``) to control it's behavior. The server and the agent use their own individual config files, and sample files are shown below.

.. NOTE:: 
   Configuration files for both ``todd-server`` and ``todd-agent`` binaries load once at initial start. So, in order to apply any changes to the configuration, these processes need to be restarted.

Server Configuration
--------------------

The server configuration file (usually ``/etc/todd/server.cfg``) contains configurations for all services that the server will require - things like integrated databases, communications plugins, and internal calculations. Here is a sample server configuration file, with comments inline:

.. code-block:: text

    # ToDD's API
    [API]
    Host = 0.0.0.0
    Port = 8080

    # Serves assets like collectors, testlets, etc.
    [Assets] 
    IP = 0.0.0.0
    Port = 8090

    # Describes parameters for the "comms" system, which manages communications between
    # the server and the agents
    [Comms]                            
    User = guest                       # Username for comms
    Password = guest                   # Password for comms
    Host = localhost                   # Hostname or IP address for comms
    Port = 5672                        # Port for comms
    Plugin = rabbitmq                  # Comms plugin to use (i.e. "rabbitmq")

    # Parameters for database connectivity
    [DB]
    IP = 192.168.0.10                  # Hostname or IP address for database
    Port = 4001                        # Port for database
    Plugin = etcd                      # Database plugin to use (i.e. "etcd")

    # Parameters for time-series database connectivity
    [TSDB]
    IP = 192.168.0.10                  # Hostname or IP address for tsdb
    Port = 8086                        # Port for tsdb
    Plugin = influxdb                  # TSDB plugin to use (i.e. "influxdb")

    [Grouping]
    Interval = 10                      # Interval (in seconds) for the grouping calculation
                                       # to run on the server

    [Testing]
    Timeout = 30                       # This is the timer (in seconds) that a test will be
                                       # allowed to live

    # Describes parameters for local resources, such as network or filesystem resources
    [LocalResources]
    DefaultInterface = eth2            # Dictates what network interface is used for testing
                                       # purposes (i.e. informs the todd-server which IP
                                       # address can be used

    IPAddrOverride = 192.168.99.100    # Overrides DefaultInterface by providing a specific IP
                                       # address rather
    
    OptDir = /opt/todd/agent           # Operational directory for the agent. Houses things like
                                       # cache files, user-defined testlets, etc.

Agent Configuration
-------------------

The agent configuration (usually ``/etc/todd/agent.cfg``) is considerably simpler than the server configuration. Again, comments are provided below to help illustrate the various options:

.. code-block:: text

    # Describes parameters for the "comms" system, which manages communications between
    # the server and the agents
    [Comms]                            
    User = guest                       # Username for comms
    Password = guest                   # Password for comms
    Host = localhost                   # Hostname or IP address for comms
    Port = 5672                        # Port for comms
    Plugin = rabbitmq                  # Comms plugin to use (i.e. "rabbitmq")

    # Describes parameters for local resources, such as network or filesystem resources
    [LocalResources]
    DefaultInterface = eth2            # Dictates what network interface is used for testing
                                       # purposes (i.e. informs the todd-server which IP
                                       # address can be used

    IPAddrOverride = 192.168.99.100    # Overrides DefaultInterface by providing a specific IP
                                       # address rather
    
    OptDir = /opt/todd/agent           # Operational directory for the agent. Houses things like
                                       # cache files, user-defined testlets, etc.
