ToDD
====
[![Build Status](https://travis-ci.org/toddproject/todd.svg?branch=master)](https://travis-ci.org/toddproject/todd)
[![Documentation Status](https://readthedocs.org/projects/todd/badge/?version=latest)](http://todd.readthedocs.org/en/latest/?badge=latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/toddproject/todd)](https://goreportcard.com/report/github.com/toddproject/todd)

ToDD stands for "Testing on Demand: Distributed!". ToDD is an extensible framework for providing natively distributed testing on demand.

> ToDD should be considered "early alpha" status at this point. Current efforts are focused primarily at increasing test coverage and greater reliability, in order to move ToDD into a more stable, beta status.

# Overview

Traditionally, the tooling used by network engineers to confirm continued network operation after any kind of change to the network is fairly limited. After a change, a network engineer may run "ping" or "traceroute" from their machine, or perhaps call some application owners to ensure that their apps are still working. Unfortunately, there is a very big difference in network activity between a 3AM change window and peak user activity during the day.

ToDD addresses gaps in today's testing software in three ways:

- Enables real-world traffic distribution for tests using very simple grouping primitives
- Provides a totally open and extensible mechanism for defining tests
- Exposes testing data in a totally open way, easily allowing for 3rd party software to analyze and visualize

ToDD is a framework through which you can deploy simple test-oriented applications in a distributed manner. With ToDD, you distribute agents around your infrastructure in any place where you feel additional "testing power" is warranted. Then, these agents can be leveraged to mimic real-world network utilization by actually running those same applications at a large scale.

Here are some key features provided by ToDD:

- **Highly Extensible** - ToDD uses an extremely generic interface (called testlets) for running applications, so that users can very easily augment ToDD to support a new application.
- **Post-Test Analytics** - ToDD integrates with time-series databases, such as influxdb. With this, engineers can schedule ToDD test runs to occur periodically, and observe the testrun metrics changing over time.
- **Grouping** - ToDD performs testruns from groups of agents, instead of one specific agent. The user will provide a set of rules that place a given agent into a group (such as hostname, or ip subnet), and ToDD will instruct all agents in that group to perform the test. This means that the power of a test can be increased by simply spinning up additional agents in that group.
- **Diverse Target Types** - Test runs can be configured to target a list of "dumb" targets (targets that are not running a ToDD agent), or a ToDD group. This is useful for certain applications where you need to be able to set up both ends of a test (i.e. setting up a webserver and then testing against it with curl, or setting up an iperf client/server combo)

# Resources

Documentation for ToDD is available [here](http://todd.readthedocs.org/en/latest/). Note that in the time immediately following the release, these docs will be constantly updated, so don't fret if these pages are a bit empty for the next few weeks.

The [ToDD mailing list](https://groups.google.com/forum/#!forum/todd-dev) is also an excellent place to get some support from the community, or to have a discussion about a feature or problem with ToDD.

# Getting Started

The best way to get ToDD running is by leveraging the Vagrantfile provided in this repository. This Vagrantfile comes with an Ansible playbook, so the simplest way to get this environment kicked off is:

    git clone git@github.com/toddproject/todd.git
    cd todd/
    vagrant up

Wait for the Ansible playbook to finish, and when it's done, you will have a VM with the Go environment set up for you, and ready for you to perform an install.

To install, SSH into the VM and use the provided Makefile:

    vagrant ssh
    cd ~/go/src/github.com/toddproject/todd
    sudo -E GOPATH=/home/vagrant/go PATH=$PATH:/home/vagrant/go make install
    todd --help

That last command should show the ToDD help output - if you see this, ToDD has been installed!

There is also a handy script for starting a few containers that are useful for running ToDD. Not only is ToDD available as a docker container (downloadable from Docker Hub) but the script will also spin up some infrastructure that ToDD will use to communicate with agents, or store test data:

    scripts/start-containers.sh

This script will kill any already-running containers, but if this is the first time running this script, it's a handy tool to get all of these services started quickly. The Ansible playbook accompanying this vagrantfile will set up Docker on the VM, so you should be able to run this script right away.

You can, of course, set up your own machine, and just use the provided makefile. If you do this, you must ensure the following is done:

- Your Go installation is available to the user you're running, and your GOPATH is set correctly. Basically, you have a working Go configuration.
- You have added $GOPATH/bin to your PATH
- You have internet access (we need to "go get" a few tools during the build)

Assuming the above criteria have been met, you should be ready to run "make install". Once this has finished, you should be able to run the "todd-server", "todd-agent", and "todd" binaries from your shell.

# Docker

A Docker container for ToDD [is available from Docker Hub](https://hub.docker.com/r/toddproject/todd/):

    docker pull toddproject/todd

# Contributing

If you want to contribute some code to ToDD, please review [CONTRIBUTING.md](https://github.com/toddproject/todd/blob/master/CONTRIBUTING.md) first.

# WARNING

With great power comes great responsibility. Do not use ToDD for any purposes other than to test your own infrastructure. ToDD was not created for nefarious reasons, and you alone are responsible for how you use ToDD.
