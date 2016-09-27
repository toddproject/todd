ping
================================

[`Visit Github Repository <https://github.com/toddproject/todd-nativetestlet-ping>`_]

TODO update docs with message about taking out sysctl commands and instead going with capabilities modification. Need to create a doc that talks about testlet compatibility broadly (what about bsd? Windows?)

Linux Sockets
-------------

# If on Linux, enable ping testlet functionality (DEPRECATED in favor of granting socket capabilities on testlets)
    # sysctl -w net.ipv4.ping_group_range="0 0" || echo "Unable to set kernel parameters to allow ping. Some testlets may not work."

capabilities