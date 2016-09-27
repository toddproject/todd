Dependencies
================================

.. toctree::
   :maxdepth: 2

   comms.rst
   db.rst
   tsdb.rst

Internal Dependencies
---------------------

Internal dependencies, such as Go libraries that ToDD uses to communicate with a message queue, or a database for example, are vendored in the ``vendor`` directory of the repository. There is no additional step to download such dependencies, in order to install ToDD from source and run it.

External Dependencies
---------------------

There are a number of external services that ToDD needs in order to run.

- `Agent Communications <comms.html>`_
- `State Database <db.html>`_
- `Time-Series Database <tsdb.html>`_

Please refer to the specific pages linked above for each to see what specific integrations have been built into ToDD in each area.
