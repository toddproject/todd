ToDD Agent Orchestration (Comms)
================================

In ToDD, the model used to communicate between the server and the agents is simply referred to as "comms". Incidentally, this is also the name of the package in Go that contains the code relevant to this effort.

Within this package, there is a file "comms.go" which contains little more than an interface and a base struct that all comms plugins must follow. In order to be considered a comms plugin, an implementation must satisfy the interface described there. This interface describes functions like AdvertiseAgent(), ListenForTasks(), and more. Through this, all plugins must implement the same behavior, and in theory, any comms plugin can be used to facilitate server-to-agent communications.

The rest of this documentation will describe the behind-the-scenes behavior of the comms plugins currently implemented within ToDD.

RabbitMQ
--------

The RabbitMQ plugin is the first comms plugin to be implemented within ToDD. This plugin uses a fairly simple model of communicating with agents, and while scale was and is an important goal for the ToDD project, the RabbitMQ plugin was designed primarily for ease of use.

