# golang irc / twitter bot
If a message starts with "twit", this bot will make a tweet with the rest of that line.

If a message contains a tweet url, the bot will attempt to fetch its contents and message the channel with them.

# setup
You need twitter application auth set up. Once thats ready, copy config_example.yml to config.yml and populate the values for your twitter creds and irc server details.

# notes
This is obviously a tiny/rough bot and doesn't include a lot of functionality.