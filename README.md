# Firefox - Cookie Monster (ffcm)

Connect to Firefox debug port and issue Javascript commands, useful for grabbing cookies.

For now I have focused on Windows, but it should work with macOS (even more useful there actually) - but I don't have a MacBook at the moment to test it.


## Technical Things and Protocol

This tool is doing things at the TCP level using `net.Dial` to get a TCP client (`Conn`) to Firefox. 

It then sends various config and setup debug messages as JSON serialized objects to eventually run Javascript commands using the `evaluateJSAsync` method and access `Services.cookies.cookies`. 

There is likely a much better/easier way to implemented this, as Firefox recently (since 78) added a `Network.getAllCookies` Debug API, and I was not yet able to figure out how to invoke that. 

The Mozilla documentation for the `Remote Debug Protocol` is located here: https://docs.firefox-dev.tools/backend/protocol.html

This tool is written in Golang using concurrent sender/receiver routines, because I thought it's cool to play around with that.
The code is a bit confusing due to having to send 5 different "kind of" messages to the server to setup and get all that is needed, and the client uses a single data structure (called `wireMessage`) to represent all possible JSON requests/responses - so it can get a bit messy trying to understand the code. Yay. :)

## Inspired by Cookie Crimes

What inspired me to research and build this? Go check out [Cookie Crimes](https://github.com/defaultnamehere/cookie_crimes) for Chrome by @mangopdf.

## Background and more Info about Browser Remote Debugging

There is more background info about the tool and browser remote debugging on my blog at: 

* [Remote Debugging with Firefox](https://embracethered.com/blog/posts/2020/cookies-on-firefox/)
* [Post-Exploitation: Abusing Chrome's debugging feature to observe and control browsing sessions remotely](https://embracethered.com/blog/posts/2020/chrome-spy-remote-control/)
* [Cookie Crimes and the new Microsoft Edge Browser](https://embracethered.com/blog/posts/2020/cookie-crimes-on-mirosoft-edge/)

Let's look at the setup first.

## Basic Usage

```
.\ffcm.exe 
```

The result is cookies in the form of `name:value:domain`.

### Want to run some other code int he debug console?

You an update the Javascript command being sent to the server by changing the `defaultCommand` constant in the source code.


### Command Line Options

* **-server**: the name of the debug server, by default localhost
* **-port**: the port of the debug server, by default set to 9222
* **-log**: flag that will enable more logging for debug purposes, by default not specified


## Pre-requisites

By default the (remote) debug port of firefox is not enabled. So the first step is to enable it, in particular depending on the scenario there are multiple Firefox configuration options to be aware of.

* devtools.debugger.remote-enabled
* devtools.debugger.prompt-connection
* devtools.chrome.enabled

If you don't expose the endpoint remotely, you only need to worry about the `devtools.chrome.enabled` setting.


### Windows Setup

By default with Firefox (unless Chrome) remote debugging is disabled. So a couple of settings have to be updated, and Firefox needs a restart for them to be picked up.

Below are a few lines of PowerShell which create a `user.js` which typically seems to get merged into the `pref.js` file. If it does not work via the `user.js` file, you can try to update the `pref.js` file directly - but for me the `user.js` file has worked well.


First you can retrieve the Firefox profile via:
```
$firstprofile = (gci $env:APPDATA\Mozilla\Firefox\Profiles\*.default-rel* -Directory | Select-Object -First 1).FullName
```

And add the following lines to the uesr.js file (by default this file does not exist):

```
write 'user_pref("devtools.chrome.enabled", true);' | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.debugger.remote-enabled", true);'  | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.debugger.prompt-connection", false);' | out-file $firstprofile\user.js -Append -Encoding ascii
```

That's it, next time Firefox starts the settings will be applied.


### Connecting

Here are two commands that might come in handy when trying this, first is to terminate all instances of Firefox:
```
Get-Process -Name firefox | Stop-Proces
```

And launching Firefox with the `-start-debugger-server` option:

```
Start-Process 'C:\Program Files\Mozilla Firefox\firefox.exe' -ArgumentList "-start-debugger-server 9222 -headless"
```

After that you can launch ffcm.exe.


### macOS Setup

// TODO


## Build

Very simple, get the code (`main.go` file) and build it.

### Get the Code

For instance download via
```
go get github.com/wunderwuzzi23/firefox-cookiemonster
```

or 

```
git clone https://github.com/wunderwuzzi23/firefox-cookiemonster
```

Now you have the code, and are ready to build it.

### Build Command

Build with:

```
build -o ffcm.exe main.go
```

#### Cross Compile

If you code Go on Linux or WSL (like I do) you can cross-compile with:

```
$ env GOARCH=amd64 GOOS=windows go build -o ffcm.exe main.go
```

### Interesting behavior with cross compiled Go binaries!

Windows Defender seems to be doing **some extra security scans for cross compiled binaries**. I got a popup from Defender saying it might take up to 10 seconds for the binary to run because its being scanned... It still ran without issues though. When compiling natively on Windows there was no extra scan or popup.


## Final Remarks

**As always the reminder that pen testing requires authorization from proper stakeholders. Be nice, don't do crimes.**

