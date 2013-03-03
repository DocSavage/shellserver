shellserver
===========

Simple file and terminal proxy server for HTML-based preentation tools.  The shellserver
system comes equipped with support for reveal.js, impress.js, and Google I/O 2012 template
slides.  An example template file for each is available at the root directory..

## Approach

There are a lot of nice HTML-based presentation tools out there.  The shellserver system 
cleanly separates your tool-specific HTML presentation files from a git repo with 
a web server and git submodules for presentation tool libraries.  It should allow easy
hacking on the server to allow proxying commands (e.g., using an embedded HTML terminal)
to your shell.

Author your presentation HTML file for a particular tool, e.g., reveal.js.  Where you would
use relative paths that call "css", "lib", and "js" directory files, add a "reveal.js/" 
prefix.  Then run "shellserver --present=/dir/with/my_talk.html" and surf to
"localhost:6789/my_talk.html" and you should be seeing your reveal.js presentation.
(This assumes you are running shellserver from the working directory.  If not, you
can manually set the shellserver working directory using the --shellserver=/path/to/repo.)

## Quick Start

If you aren't a Go developer, this will suffice:

    % git clone https://github.com/DocSavage/shellserver.git
    % cd shellserver
    % git submodule init     # These two commands will pull all presentation libraries
    % git submodule update

If you aren't on 64-bit Mac, copy the appropriate executable for your platform:

    % rm shellserver
    % cp bin/0.1/linux_amd64/shellserver . 

Run the server.  By default this will use the git working directory as both the
presentation files directory and the shellserver directory.

    % ./shellserver

Now you are ready to try the included demos.

### See reveal.js presentation

Point your web browser to [localhost:6789/reveal.html](http://localhost:6789/reveal.html)
to see the basic reveal.js template.  (Only slightly modified to add 'reveal.js/' to
the relative paths in the html file.)

### See Google I/O 2012 slide template

Point your web browser to [localhost:6789/google-io.html](http://localhost:6789/google-io.html)
to see the basic Google I/O template.  (Only slightly modified to add 'google-io/' to
the relative paths in the html file.)

### See impress.js presentation

Point your web browser to [localhost:6789/impress.html](http://localhost:6789/impress.html)
to see the basic impress.js template.  (Only slightly modified to add 'impress.js/' to
the relative paths in the html file.)

## In-slide terminal support

shellserver supports an in-slide terminal that proxies commands (after hitting enter) to
the Go server, which then executes the command and returns the results.  Background
commands are supported using the "&" ending.  Also, the "cd /path/to/dir" command works
by changing the terminal's current working directory.

## Go Developers

There's only one small file, shellserver.go, and it's cross-compiled with 
[goxc](http://www.laher.net.nz/goxc/) but has only passed testing with 64-bit Mac.
64-bit Fedora 16 works fine if you do a "go build shellserver.go" at the Linux
prompt.  (Cross-compilation with goxc doesn't seem to work.)

## TODO

* <del>Add presentation terminal that proxies commands out to server, which actually runs and
returns the results.</del>
* Add Go in-presentation compilation and execution like [play.golang.org](http://play.golang.org).
