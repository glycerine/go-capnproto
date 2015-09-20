Version 1.0 vs 2.0
-------

Update 2015 Sept 20:

Big news! Version 2.0 of the go-bindings, authored by Ross Light, is now released and newly available! It features capnproto RPC and capabilities support. See https://github.com/zombiezen/go-capnproto2 for the v2 code and docs.

This repository (https://github.com/glycerine/go-capnproto) is now being called version 1.0 of the go bindings for capnproto. It does not have RPC (it was created before the RPC protocol was defined). Version 1 has schema generating tools such as https://github.com/glycerine/bambam. Personally I have many projects that use v1 with mangos for network transport; https://github.com/gdamore/mangos. Here is an example of using them together: https://github.com/glycerine/goq. Nonetheless, for new projects, especially once v2 has been hardened and tested, v2 should be preferred.

Version 1 will be maintained for applications that currently use it.  However new users, new features and new code contributions should be directed to the version 2 code base to take advantage of the RPC and capabilities.


License
-------

MIT - see LICENSE file

Documentation
-------------
In godoc see http://godoc.org/github.com/glycerine/go-capnproto


News
----

5 April 2014: James McKaskill, the author of go-capnproto (https://github.com/jmckaskill/go-capnproto), 
has been super busy of late, so I agreed to take over as maintainer. This branch 
(https://github.com/glycerine/go-capnproto) includes my recent work to fix bugs in the
creation (originating) of structs for Go, and an implementation of the packing/unpacking capnp specification.
Thanks to Albert Strasheim (https://github.com/alberts/go-capnproto) of CloudFlare for a great set of packing tests. - Jason

Getting started
---------------

New! Visit the sibling project to this one, [bambam](https://github.com/glycerine/bambam), to automagically generate a capnproto schema from the struct definitions in your go source files. Bambam makes it easy to get starting with go-capnproto.

pre-requisite: Due to the use of the customtype annotation feature, you will need a relatively recent capnproto installation.  At or after 1 July 2014 (at or after b2d752beac5436bada2712f1a23185b78063e6fa) is known to work.

~~~
# first: be sure you have your GOPATH env variable setup.
$ go get -u -t github.com/glycerine/go-capnproto
$ cd $GOPATH/src/github.com/glycerine/go-capnproto
$ make # will install capnpc-go and compile the test schema aircraftlib/aircraft.capnp, which is used in the tests.
$ diff ./capnpc-go/capnpc-go `which capnpc-go` # you should verify that you are using the capnpc-go binary you just built. There should be no diff. Adjust your PATH if necessary to include the binary capnpc-go that you just built/installed from ./capnpc-go/capnpc-go.
$ go test -v  # confirm all tests are green
~~~

What is Cap'n Proto?
--------------------

The best cerealization...

http://kentonv.github.io/capnproto/


