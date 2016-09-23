Using ToDD in a Container or Virtual Machine
================================

Docker
------
If you instead wish to run ToDD inside a Docker container, you can pull the current image from Dockerhub:

.. code-block:: text

    mierdin@todd-1:~$ docker pull mierdin/todd
    mierdin@todd-1:~$ docker run --rm mierdin/todd todd -h                        
    NAME:
       todd - A highly extensible framework for distributed testing on demand

    USAGE:
       todd [global options] command [command options] [arguments...]

    VERSION:
       v0.1.0
    ......

The binaries below are distributed inside this container and can be run as commands on top of the "docker run" command:

- ``todd`` - the CLI client for ToDD
- ``todd-server`` - the ToDD server binary
- ``todd-agent`` - the ToDD agent binary

A Dockerfile for running any ToDD component (server/agent/client) is provided in the repository if you wish to build the image yourself. This Dockerfile is what's used to automatically build the Docker image within Dockerhub.

Vagrant
-------
There is also a provided vagrantfile in the repo. This is not something you should use to actually run ToDD in production, but it is handy to get a quick server stood up, alongside all of the other dependencies like a database. This Vagrantfile is configured to use the provided Ansible playbook for provisioning, so in order to get a nice ToDD-ready virtual machine, one must only run the following from within the ToDD directory:

.. code-block:: text

    vagrant up
