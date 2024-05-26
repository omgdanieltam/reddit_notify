import argparse
import configparser
import sys
import time
import requests
import json
import datetime
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart

# Arguments parse
parser = argparse.ArgumentParser(description='Reddit notify on keywords')
parser.add_argument("--config", type=str, help="name of config file", 
                    default='config.ini.example')
args = parser.parse_args()

# Config parse
config_parser = configparser.ConfigParser()

# HTTP headers
http_headers = {
    'User-Agent': 'Python_Reddit_Notif_Dtam/1.0'
}

# Etc
nl = '\n'

# Config error message
def config_error(msg):
    print(f'Error parsing config file: {msg}')
    sys.exit(1)

# Read config file
def get_config(filename):
    config = {}

    try:
        config_parser.read(filename)

        # If unable to read config
        if len(config_parser.sections()) == 0 :
            raise configparser.Error
        
    except configparser.Error as e:
        print(f'Error reading config file {e}')
        sys.exit(1)

    # Parse config file
    if config_parser.has_option('app', 'subreddit') and \
            len(config_parser.get('app', 'subreddit').strip()) > 0:
        config['subreddit'] = config_parser.get('app', 'subreddit').strip()
    else:
        config_error("Missing 'subreddit'")

    if config_parser.has_option('app', 'interval') and \
            len(config_parser.get('app', 'interval').strip()) > 0:
        config['interval'] = config_parser.get('app', 'interval').strip()
    else:
        config_error("Missing 'interval'")

    if config_parser.has_option('app', 'keyword') and \
            len(config_parser.get('app', 'keyword').strip()) > 0:
        config['keyword'] = config_parser.get('app', 'keyword').strip()
    else:
        config_error("Missing 'keyword'")

    # SMTP options
    if config_parser.has_option('smtp', 'smtp_server') and \
            len(config_parser.get('smtp', 'smtp_server').strip()) > 0:
        config['smtp_server'] = config_parser.get('smtp', 'smtp_server').strip()
    else:
        config_error("Missing 'smtp_server'")

    if config_parser.has_option('smtp', 'smtp_port') and \
            len(config_parser.get('smtp', 'smtp_port').strip()) > 0:
        config['smtp_port'] = config_parser.get('smtp', 'smtp_port').strip()
    else:
        config_error("Missing 'smtp_port'")

    if config_parser.has_option('smtp', 'smtp_username') and \
            len(config_parser.get('smtp', 'smtp_username').strip()) > 0:
        config['smtp_username'] = config_parser.get('smtp', 'smtp_username').strip()
    else:
        config_error("Missing 'smtp_username'")

    if config_parser.has_option('smtp', 'smtp_password') and \
            len(config_parser.get('smtp', 'smtp_password').strip()) > 0:
        config['smtp_password'] = config_parser.get('smtp', 'smtp_password').strip()
    else:
        config_error("Missing 'smtp_password'")

    if config_parser.has_option('smtp', 'smtp_to') and \
            len(config_parser.get('smtp', 'smtp_to').strip()) > 0:
        config['smtp_to'] = config_parser.get('smtp', 'smtp_to').strip()
    else:
        config_error("Missing 'smtp_to'")

    if config_parser.has_option('smtp', 'smtp_from') and \
            len(config_parser.get('smtp', 'smtp_from').strip()) > 0:
        config['smtp_from'] = config_parser.get('smtp', 'smtp_from').strip()
    else:
        config_error("Missing 'smtp_from'")

    return config

# Constant loop to check subreddit
def check_reddit(config):
    subreddits = [subreddit.strip() for subreddit in config.get('subreddit').split(',')]
    keywords = set([keyword.strip().lower() for keyword in config.get('keyword').split(',')])

    while(True):
        for subreddit in subreddits:
            resp = requests.get(f'https://www.reddit.com/r/{subreddit}/new.json', headers=http_headers)

            if resp.status_code == 200:

                # Body
                full_body = json.loads(resp.text)

                # Number of items
                items = int(full_body.get('data').get('dist'))
                
                # Loops through items
                for i in range(0, items):
                    # Gather data
                    title = full_body.get('data').get('children')[i].get('data').get('title')
                    text = full_body.get('data').get('children')[i].get('data').get('selftext')

                    # Loop through keywords
                    for keyword in keywords:

                        if keyword in title.lower() or keyword in text.lower():
                            # Gather additional data
                            url = full_body.get('data').get('children')[i].get('data').get('permalink')
                            timestamp = full_body.get('data').get('children')[i].get('data').get('created_utc')

                            # Send alert
                            send_alert(config, title, text, url, timestamp, keyword)
        
        time.sleep(config.get('interval'))

# Send alert out
def send_alert(config, title, text, url, timestamp, keyword):
    # Setup
    smtp_from = config.get('smtp_from')
    smtp_to = config.get('smtp_to')

    # Setup message
    message = MIMEMultipart()
    message["From"] = smtp_from
    message["To"] = smtp_to
    message["Subject"] = f'Reddit Notify: Found Match ({title})'
    time_format = datetime.datetime.fromtimestamp(timestamp)
    body = f'Keyword: {keyword}{nl}{nl}{nl}\
        {title}{nl}{nl}{nl}\
        {text}{nl}{nl}{nl}\
        URL: {url}{nl}\
        Time: {time_format.strftime("%Y-%m-%d %H:%M:%S")}'
    
    message.attach(MIMEText(body, "plain"))

    # Send message
    try:
        with smtplib.SMTP(config.get('smtp_server'),
                        config.get('smtp_port')) as server:
            server.login(config.get('smtp_username'),
                        config.get('smtp_password'))
            text = message.as_string()

            server.sendmail(smtp_from, smtp_to, text)
    except Exception as e:
        print(f'SMTP Send Error: {e}')

if __name__ == '__main__':

    config = get_config(args.config)

    print(f'Current config file: {config}')

    check_reddit(config)