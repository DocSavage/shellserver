shellserver
===========

Simple presentation server for HTML-based tools like reveal.js and Google I/O slide templates.  Packages html terminal, web server, and shell proxy.

## Approach

There are a lot of nice HTML-based presentation tools out there.  This is a simple tool 
that lets you separate your tool-specific HTML presentation files and a git repo with 
a web server and git submodules for each presentation tool library.

Author your presentation HTML file for a particular tool, e.g., reveal.js.  Where you would
use relative paths that call "css", "lib", and "js" directory files, add a "reveal.js/" 
prefix.  Then run "shellserver --present=/dir/with/presentation.html" and surf to
"localhost:6789/presentation.html" and you should be seeing your reveal.js presentation.
(This assumes you are running shellserver from the working directory.  If not, you
can manually set the shellserver working directory using the --shellserver=/path/to/repo.)

## Quick Start

If you aren't a Go developer, this will suffice:

    % git clone https://github.com/DocSavage/shellserver.git
    % cd shellserver
    % cp bin/darwin_amd64/shellserver .  (or copy the appropriate executable for your platform)
    % ./shellserver

### See reveal.js presentation

Point your web browser to [localhost:6789/revealjs.html](http://localhost:6789/revealjs.html)
to see the basic reveal.js template.  (Only slightly modified to add 'reveal.js/' to
the relative paths in the html file.)

### See Google I/O 2012 slide template

Point your web browser to [localhost:6789/google-io.html](http://localhost:6789/google-io.html)
to see the basic reveal.js template.  (Only slightly modified to add 'google-io/' to
the relative paths in the html file.)

## Go Developers

There's only one small file, shellserver.go, and it's cross-compiled with 
[goxc](http://www.laher.net.nz/goxc/) of which I've tested it on one system:
64-bit Mac :)  64-bit Linux and Windows 7 coming up.

## TODO

* Add presentation terminal that proxies commands out to server, which actually runs and
returns the results.
* Add Go in-presentation compilation and execution like [play.golang.org](http://play.golang.org).