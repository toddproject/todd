Working with the ToDD Client
================================

ToDD comes with an easy-to-use CLI client. Use the ``--help`` flag to print some useful help output:

.. code-block:: text
   :emphasize-lines: 1

    mierdin@todd-1:~$ todd --help
    NAME:
       todd - A highly extensible framework for distributed testing on demand

    USAGE:
       todd [global options] command [command options] [arguments...]

    VERSION:
       v0.1.0

    COMMANDS:
       agents   Show ToDD agent information
       create   Create ToDD object (group, testrun, etc.)
       delete   Delete ToDD object
       groups   Show current agent-to-group mappings
       objects  Show information about installed group objects
       run      Execute an already uploaded testrun object
       help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       -H, --host "localhost"   ToDD server hostname
       -P, --port "8080"        ToDD server API port
       --help, -h           show help
       --version, -v        print the version


Agents
----------

Use the ``todd agents`` command to display information about the agents currently known to the ToDD server.

.. code-block:: text

    mierdin@todd-1:~$ todd agents --help
    NAME:
       temp-client agents - Show ToDD agent information

    USAGE:
       temp-client agents [arguments...]

You can run this command on it's own to display a summary of all agents, their facts, collectors, etc:

.. code-block:: text

    mierdin@todd-1:~$ todd agents
    UUID          EXPIRES ADDR        FACT SUMMARY        COLLECTOR SUMMARY
    4c1ef1fd94ce  23s     172.18.0.7  Addresses, Hostname get_addresses, get_hostname
    cba4e720efae  24s     172.18.0.8  Addresses, Hostname get_addresses, get_hostname
    555dacccb4ae  24s     172.18.0.9  Addresses, Hostname get_addresses, get_hostname
    79ffae90354e  24s     172.18.0.10 Hostname, Addresses get_addresses, get_hostname
    42b1341c22fe  24s     172.18.0.11 Addresses, Hostname get_addresses, get_hostname
    fdb4c3ddc8eb  25s     172.18.0.12 Addresses, Hostname get_hostname, get_addresses

Or, you could append an agent UUID to this command to see detailed information about that agent, such as the facts that it is reporting:

.. code-block:: text

    mierdin@todd-1:~$ todd agents 4c1ef1fd94ce 
    Agent UUID:  4c1ef1fd94ce91c9c589880c47fb5374bba91ecdeb852a9ac3bb4278507c0ba4
    Expires:  25s
    Collector Summary: get_addresses, get_hostname
    Facts:
    {
        "Addresses": [
            "127.0.0.1",
            "::1",
            "172.18.0.7",
            "fe80::42:acff:fe12:7"
        ],
        "Hostname": [
            "todd-agent-0"
        ]
    }

Create
----------

.. IMPORTANT::
   See the documentation on `ToDD objects <objects.html>`_ for an explanation on what they are and how they work.

Use the ``todd create`` command to upload an object to the ToDD server.

.. code-block:: text

    mierdin@todd-1:~$ todd agents --help
    NAME:
       temp-client agents - Show ToDD agent information

    USAGE:
       temp-client agents [arguments...]

Delete
----------

.. IMPORTANT::
   See the documentation on `ToDD objects <objects.html>`_ for an explanation on what they are and how they work.

Run "todd delete"

Groups
----------

Run "todd create"

Objects
----------

.. IMPORTANT::
   See the documentation on `ToDD objects <objects.html>`_ for an explanation on what they are and how they work.


Run "todd objects <type>"

Run
----------

Run "todd create"

Show optional arguments

