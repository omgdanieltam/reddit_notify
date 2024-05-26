# Reddit notify
An app that periodically checks a subreddit for a keyword and will send out an SMTP message if found

# Config
Configuration is done via a `config.ini` file within the same location of the application. Check the example on how to properly set up the `config.ini` file. You can also pass in a command line argument to specify the location of the config file:
```
python main.py --config ~/config.ini
```

## App
The app section defines running configuration to parse through.

### Subreddit
The subreddit of your choosing:
```
subreddit = hardwareswap
```

The configuration does allow for multiple subreddits to be parsed through by adding a comma between the subreddits:
```
subreddit = hardwareswap, mechmarket
```

### Interval
The interval in minutes on how ofter to request data from Reddit. This example will request data from Reddit every 5 minutes:
```
interval = 5
```

### Keyword
The keywords to be matched:
```
keyword = holy pandas
```

The configuration does allow for multiple keywords to be parsed through by adding a comma between the keywords:
```
keyword = holy pandas, boba u4t
```

## SMTP
The SMTP section defines the configuration for sending to the SMTP server.

### SMTP Server
The location of the SMTP server:
```
smtp_server = localhost
```

### SMTP Port
The port used for SMTP (at this current time it does not support TLS):
```
smtp_port = 25
```

### SMTP Username
The username used to authenticate to the SMTP server:
```
smtp_username = username
```

### SMTP Password
The password used to authenticate to the SMTP server:
```
smtp_password = password
```

### SMTP To
The e-mail address you want to send the SMTP message to:
```
smtp_to = example_to@example.com
```

### SMTP From
The e-mail address you want to send the SMTP message from:
```
smtp_from = example_from@example.com
```