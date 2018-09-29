# t

A dead simple (tea) timer for the CLI. Works on Linux, Mac and Win. Opens things to notify.

## Install

Run `go get github.com/dbriemann/t`. If you don't like the `t` name or it conflicts with some other program just rename it in the Go bin folder.

Currently this is only tested on Linux. Please tell me if it works or doesn't work on Windows and Mac.

## Usage

Just running `t` will print you the (very simple) help and a table listing all your timers. The table will be empty in the beginning. Let's add some timers.

Run the following commands one after the other:

```
t sencha 1m20s ~/.config/t/pics/teapot.jpg
t pizzo 10m "https://www.google.com/search?q=pizza&tbm=isch"
t oops 1s "https://google.com"
```

This will result in:
```
Saved timers:
+----+--------+-----------+---------------------------------+------+
| ID |  NAME  | COUNTDOWN |               TARGET            | USED |
+----+--------+-----------+---------------------------------+------+
|  1 | sencha | 1m20s     | /home/dlb/.con..ics/teapot.jpg  |    0 |
|  2 | pizzo  | 10m0s     | /home/dlb/.con..pics/pizza.jpg  |    0 |
|  3 | oops   | 1s        | https://google.com              |    0 |
+----+--------+-----------+---------------------------------+------+

```

Note that you have to specify your own files (pics, sounds..), `t` doesn't include any. You can also specify links. All paths and links must be absolute. And remember to put your links into quotes so your shell doesn't do crazy stuff.

Now we made some mistakes. Let's correct them. Renaming pizzo to pizza is easy. Just do:

```
t pizzo = pizza
```

And now let's delete that timer that nobody really wanted:

```
t oops del
```

### Spontaneous Timing

Sometimes you just need to time something right now and don't want to save a timer first. Here you go:

```
t 6m23s
```

runs a timer for 6 minutes and 23 seconds and then opens the daily bing photo in your web browser.

## What else can it do?

Nothing, sorry :)

## Future

This will be a toolbox. List of possible features:

- timers (done)
- notes
- calendar
- weather
