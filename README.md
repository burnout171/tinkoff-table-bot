## Tinkoff Table Bot
The bot simplify [tinkoff table](https://journal.tinkoff.ru/spreadsheet/) management generally allow you to add you daily excepnses using Telegram messenger. Currently it also has a possibility to show your daily balance, monthly balance and monthly accumulation.

### How to use
I develop the bot mainly for myself. If you want to use it too you need to
1. Fork this repositroy.
2. Play with [google sheet API](https://developers.google.com/sheets/api/quickstart/go) to get your credentials.
3. Register your bot with [BotFather](https://core.telegram.org/bots#3-how-do-i-create-a-bot).
4. Create a project for it on [Heroku](https://www.heroku.com/) and link it with your fork.
5. Prepare Heroku environment with `configureEnvironment.sh` script and run the bot.
6. Enjoy!