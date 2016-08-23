ToDD Dependencies
================================

.. toctree::
   :maxdepth: 2

   comms.rst
   db.rst
   tsdb.rst

Internal Dependencies
---------------------

If you're not a developer concerned with the internals of ToDD, this is not something you need to worry about. Internal dependencies, such as Go libraries that ToDD uses to communicate with a message queue, or a database for example, are vendored in the `vendor` directory of the repository.

External Dependencies
---------------------

There are a number of external services that ToDD needs

- `Agent Communications <comms.html>`_
- `State Database <db.html>`_
- `Time-Series Database <tsdb.html>`_

Each of these dependencies can potentially be satisfied by multiple existing software projects, but please refer to the specific pages for each to see what specific integrations have been built into ToDD thus far.