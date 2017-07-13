# NXIRCD

This readme is very brief and is being worked on if you have any questions hit us up in the issues tab or 
on 
```
Server: irc.centralchat.net 
Channel: #nxircd
```

## Downloading

Download the correct release from the release page


## Configuring

Create a config.json in the folder where your binary is


### Example Config
```
{
  "name":    "irc.nxircd.org",
  "network": "nxircd",
  
  "loglevel": "DEBUG",
  "listen": [
    { "host": "127.0.0.1:6666", "type" : "ircd" },
    { "host": "127.0.0.1:6667", "type" : "ircd" },
    { "host": "127.0.0.1:9001", "type" : "ws" },
    { 
       "host": ":8080",         
       "type":  "web",   
       "options": {
         "auth": "enabled"
       }
    }
  ],  
  "ircops": [
    {
      "user": "developer",
      "pass": "password",
      "hosts": [
        "*.example.com",
        "localhost"
      ]
    }
  ]
}
```