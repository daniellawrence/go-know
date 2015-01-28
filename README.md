go-know
---------

A quick hack to gather up information about a host to then push into a central location.


Building
----------

    $ sudo apt-get install libmagic-dev
    $ go build gather.go
	$ go build cater.go
	$ ls categer gather


Overview
----------

run `gather` on all the nodes that you want to gather up files in `/etc`.
It will push all the files from /etc into a redis server.

run `cater` to print file contents from redis server.


Examples
---------

Gather `etc` on all nodes

    node1$ ./gather
    node2$ ./gather
    node3$ ./gather

Check node1's /etc/passwd to see if it has a root user

    laptop$ ./cater node1:/etc/passwd |grep -c root
	1

Check to see how many root users in across all the nodes

    laptop$ ./cater *:/etc/passwd |grep -c root
	3
