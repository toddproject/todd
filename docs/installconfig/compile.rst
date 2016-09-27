Compile from Source
================================

First, make sure the following software is installed and correctly configured for your platform:

- Go (at least 1.6)
- Ensure ``$GOPATH\bin`` is also in your ``$PATH``
- Git

.. NOTE:: 

   If you are installing ToDD on a Raspberry Pi, there are specialized install instructions `here <installrpi.html>`_. 

The best way to install ToDD onto a system is with the provided Makefile. In this section, we'll retrieve the ToDD source, compile into the three ToDD binaries, and install these binaries onto the system.

First, ``go get`` the ToDD source. As mentioned at the beginning of this document, this assumes a system where Go has been properly set up:

.. code-block:: text

    go get -d github.com/toddproject/todd

.. NOTE:: 

    At this point, you may get an error along the lines of "no buildable GO source files in...". Ignore this error; you should still be able to install ToDD.

Navigate to the directory where Go would have downloaded ToDD. As an example:

.. code-block:: text

    cd $GOPATH/src/github.com/toddproject/todd

Finally, compile and install the binaries:

.. code-block:: text

    make
    sudo make install
