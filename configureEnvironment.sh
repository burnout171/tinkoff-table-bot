#!/bin/bash

herokuProjectName=<NAME>

heroku config:set -a ${herokuProjectName} TELEGRAM_TOKEN=<TELEGRAM_TOKEN>
heroku config:set -a ${herokuProjectName} GOOGLE_CLIENT_ID=<GOOGLE_CLIENT_ID>
heroku config:set -a ${herokuProjectName} GOOGLE_PROJECT_ID=<GOOGLE_PROJECT_ID>
heroku config:set -a ${herokuProjectName} GOOGLE_AUTH_URI=<GOOGLE_AUTH_URI>
heroku config:set -a ${herokuProjectName} GOOGLE_TOKEN_URI=<GOOGLE_TOKEN_URI>
heroku config:set -a ${herokuProjectName} GOOGLE_CLIENT_SECRET=<GOOGLE_CLIENT_SECRET>
heroku config:set -a ${herokuProjectName} GOOGLE_REDIRECT_URIS=<GOOGLE_REDIRECT_URIS>
heroku config:set -a ${herokuProjectName} SHEET_ID=<SHEET_ID>
heroku config:set -a ${herokuProjectName} SHEET_ACCESS_TOKEN=<SHEET_ACCESS_TOKEN>
heroku config:set -a ${herokuProjectName} SHEET_TOKEN_TYPE=<SHEET_TOKEN_TYPE>
heroku config:set -a ${herokuProjectName} SHEET_REFRESH_TOKEN=<SHEET_REFRESH_TOKEN>
heroku config:set -a ${herokuProjectName} SHEET_TOKEN_EXPIRE_TIME=<SHEET_TOKEN_EXPIRE_TIME>
heroku config:set -a ${herokuProjectName} ENABLE_DEBUG=<ENABLE_DEBUG>