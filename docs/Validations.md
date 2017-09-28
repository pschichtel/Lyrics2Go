Validators
==========

Validators are meant to verify the loaded lyrics.
An example where this would be necessary:

A website that doesn't return an error 404 when lyrics were not found,
but instead show a page with the exact same format, but a "not found"-message
instead of lyrics.

The result of validators can be inverted by prefixing their name with "not ".

Examples:
* [contains, lyrics]
* [not contains, not found]

contains
--------
This validator checks whether the content contains a given string (first argument).

contains_ci
--------
This validator checks whether the content contains a given string (first argument) while ignore the character casing.


matches
-------
This validator checks whether the given regex (first argument) matches something in the content.
