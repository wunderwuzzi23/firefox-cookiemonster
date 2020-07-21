# firefox-cookiemonster
Connect to Firefox debug port and issue a Javascript command to grab cookies.

For now I have focused on Windows, but it should work also with macOS - but I don't have a MacBook at the moment to test it.


## Technical things and protocol

This tool is doing things at the lowest possible level using a `TCP client` to connect to Firefox and send debug messages to eventually run Javascript commands using the `evaluateJSAsync` method and access `Services.cookies.cookies`. The `Services` object is only available when setting `devtools.chrome.enabled` to true in the user's settings - more about that in the pre-reqs.

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

## Pre-requisites

By default the (remote) debug port of firefox is not enabled. So the first step is to enable it, in particular depending on the scenario there are multiple Firefox configuration options to be aware of.

* devtools.debugger.remote-enabled
* devtools.debugger.prompt-connection
* devtools.chrome.enabled

If you don't expose the endpoint remotely, you only need to worry about the `devtools.chrome.enabled` setting.


### Windows Setup

TODO: This needs a bit more research for the minimum amount of steps needed.

```

$firstprofile = (gci $env:APPDATA\Mozilla\Firefox\Profiles\*.default-release -Directory | Select-Object -First 1).FullName
gci $env:APPDATA\Mozilla\Firefox\Profiles\*.default-release

write 'user_pref("devtools.debugger.remote-enabled", true);'  | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.debugger.prompt-connection", false);' | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.chrome.enabled", true);' | out-file $firstprofile\user.js -Append -Encoding ascii
```


## macOS Setup

// TODO

## Build

Get the code (main.go file) and build it:

```
go get github.com/wunderwuzzi23/firefox-cookiemonster
build -o ffcm main.go
```

If you code Go on Linux or WSL (like I do) you can cross-compile with:

```
$ env GOARCH=amd64 GOOS=windows go build -o ffcm.exe main.go
```

### Interesting behavior with cross compiled Go binaries

Windows Defender seems to be doing an some extra security scans for cross compiled binaries. I got a popup from Defender saying it might take up to 10 seconds for the binary to run because its being scanned... It still ran without issues though. When compiling natively on Windows there was no extra scan or popup.

** As always the reminder that pen testing requires authorization from proper stakeholders. Be nice, don't do crimes. **

