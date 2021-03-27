httpipe
=======
http <-> stdio bridge inspired by <https://github.com/openfaas/faas>.

![meme chiba dog kiss http stdio](./meme.jpg)

Highlights
----------
- each function gets a subdirectory in `fn` directory
- on GET request, it displays a `doc` file
- on POST request, it runs `handle` executable

Example
-------
The simplest use case is with strings:

    $ cat ./fn/b64e/handle
    #!/bin/sh -e
	base64

	$ curl -d "hello world" 0.0.0.0:3000/b64e
	aGVsbG8gd29ybGQ=

	$ curl -d "aGVsbG8gd29ybGQ=" 0.0.0.0:3000/b64d
	hello world%

But piping also works with files:

    $ cat ./fn/bw/handle
    #!/bin/sh -e
    convert - -grayscale Rec709Luminance fd:1

    $ curl --data-binary @./meme.jpg 0.0.0.0:3000/bw > meme_bw.jpg
