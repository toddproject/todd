Testlets
================================

.. toctree::
   :maxdepth: 1

   nativetestlets/nativetestlets.rst
   usertestlets.rst

Testing applications are called "testlets" in ToDD. This is a handy way of referring to "whatever is actually doing the work of testing". This concept keeps things very neatly separated - the testlets focus on performing tests, and ToDD focuses on ensuring that work is distributed as the user directs.

Testlets are executed directly by the ToDD agent. The agent will pass a standardized set of arguments to the testlet, such as the target of the test, as well as any other useful parameters. Once finished, the testlet will then send metrics back to the agent by printing a JSON object to stdout.

.. image:: ../images/testlet.png

This generic interface makes testing in ToDD very flexible.

There are a number of testlets (known as "native testlets") that have been developed as part of the ToDD project, because they represent very common use cases for many users:

* `http <nativetestlets/http.html>`_
* `bandwidth <nativetestlets/bandwidth.html>`_
* `ping <nativetestlets/ping.html>`_
* `portknock <nativetestlets/portknock.html>`_

They run as separate binaries, and are executed in the same way that custom testlets might be executed, if you were to provide one. If you install ToDD using the provided instructions, these are also installed on the system.

.. NOTE::

   If, however, you wish to build your own custom testlets, refer to `Custom Testlets <customtestlets.rst>`_; you'll find it's quite easy to build your own testlets and run them with ToDD. This extensibility was a core design principle of ToDD since the beginning of the project.

If you're not a developer, and/or you just want to USE these native testlets, you can install these binaries anywhere in your PATH. The included Makefile will do this for you (provided a proper Go setup), and future installation methods will also automate this process.
