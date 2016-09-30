Using ToDD in a Container or Virtual Machine
================================

Docker
------
If you instead wish to run ToDD inside a Docker container, you can pull the current image from Dockerhub:

.. code-block:: text

    mierdin@todd-1:~$ docker pull toddproject/todd
    mierdin@todd-1:~$ docker run --rm toddproject/todd todd -h                        
    NAME:
       todd - A highly extensible framework for distributed testing on demand

    USAGE:
       todd [global options] command [command options] [arguments...]

    VERSION:
       v0.1.0
    ......

All three ToDD binaries (as well as native testlets) are distributed inside this container and can be run as commands on top of the "docker run" command:

- ``todd`` - the CLI client for ToDD
- ``todd-server`` - the ToDD server binary
- ``todd-agent`` - the ToDD agent binary

A Dockerfile is provided in the repository if you wish to build the image yourself. The Docker image repository is configured to automatically build the image from this Dockerfile whenever changes are pushed to the `master` branch in Github, so you always know you're pulling down the latest and greatest.

Vagrant
-------
There is also a provided vagrantfile in the repo. This is not something you should use to actually run ToDD in production, but it is handy to get a quick server stood up, alongside all of the other dependencies like a database. This Vagrantfile is configured to use the provided Ansible playbook for provisioning, so in order to get a nice ToDD-ready virtual machine, one must only run the following from within the ToDD directory:

.. code-block:: text

    vagrant up
