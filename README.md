# firefox-cookiemonster
Connect to Firefox debug port and issue a Javascript command to grab cookies.

For now I have focused on Windows, but it should work also with macOS - but I don't have a MacBook at the moment to test it.

There is likely a much better/easier way to implemented this, especially since recent Firefox versions expose a `Network.getAllCookies` API - but I was not able to figure out how to invoke that yet. So this is doing things at the lowest possible level using a TCP client sending the debugging messages of Firefox to connect and send Javascript debug command to access `Services.cookies.cookies`.

Its written in Golang using concurrent sender/receiver routines, because I thought it's cool to play around with that.
The code is a bit confusing due to having to send 5 different "kind of" messages to the server to setup and get all that is needed, and the client uses a single data structure (called `wireMessage`) to represent all possible JSON requests/responses - so it can get a bit messy trying to understand the code. Yay. :)

**Inspired by Cookie Crimes** 

What inspired me to research and build this? Go check out [Cookie Crimes](https://github.com/defaultnamehere/cookie_crimes) for Chrome by "Alex".

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

## As always the reminder that pen testing requires authorization from proper stakeholders. Be nice, don't do crimes.

