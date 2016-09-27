ping
================================

[`Visit Github Repository <https://github.com/toddproject/todd-nativetestlet-ping>`_]

The ``ping`` testlet provides basic ICMP echo tests. It reports on things like latency, and packet loss.

Linux Sockets
-------------

On Linux, the ability to leverage ICMP sockets in software usually requires special permissions. The easiest answer is to run such software as root, or with equivalent permissions.

However, this carries it's own complexity - so there are two alternatives to running the ``ping`` testlet:

* A `2011 commit <http://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/commit/?id=c319b4d76b9e583a5d88d6bf190e079c4e43213d>`_ to the Linux kernel introduced a new ICMP socket type. Using this socket does not require root, but only if the system is configured to do so by configuring ``net.ipv4.ping_group_range`` with ``sysctl``. This turned out to be a problem when running ToDD in a Docker container, as in this case, the kernel is shared, and would require host configuration changes. Not exactly an ideal solution, especially in cloud deployments.
* The ``setcap`` command can be used to provide "special" permissions to an executable binary called "capabilities". One capability, "cap_net_raw", allows applications to use raw sockets without root. This is not a system-wide setting, but rather is granted to a specific application. The ``scripts/set-testlet-capabilities.sh`` script, which is invoked when running ``make install`` to build ToDD from source, grants the ``ping`` testlet this capability, as well as any other native testlet that might need it.

The second option was chosen, as it was a simpler and more secure option, and one that worked on a variety of platforms.

The ``ping`` testlet will first try to use ICMP raw sockets, and then will fall back to using UDP. This allows it to work on both OSX (Darwin) and Linux.
