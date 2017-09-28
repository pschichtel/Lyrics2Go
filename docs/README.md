Documentation
====================

Links
-----
* [Filter](Filters.md)
* [Validator](Validations.md)
* [Example configuration](../providers/example.yml)

Regular expressions
-------------------
This application relies heavily on Go's regular expressions for its functionality.
Go uses Google's [re2](https://github.com/google/re2) linear-time regular expression engine.
This imposes a few limitations, most notably: No support for look-around and backtracking. 
Tutorials and tools to write these expressions are all over the internet, for example [RegExr](http://www.regexr.com)
which is a greate tool to write and test regular expressions. A nice automata-like visualization is provided by [Debuggex](https://www.debuggex.com).
Remember though: These sites support backtracking and look-around which are not supported here.

A syntax overview is provided in [re2's source repository](https://github.com/google/re2/blob/master/doc/syntax.txt).

There are a few things to watch out for:
* Flags (like case-sensitivity) are specified like this: ```(?is)``` (`i` for case-insensitive, `s` for single-line behavior), ther flags can be found in the syntax overview
* backslashes (```\\```) must be escaped when using double quotes ("), as it is a meta character in the YAML format: ```\s``` -> ```\\\s```
* named capturing groups are defined like this: ```(?P<lyrics>.*?)``` with "lyrics" being the name of the group
