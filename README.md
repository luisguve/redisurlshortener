

### Important note 

The hobby tier for Heroku Redis does not have any persistence. This means that, if the instance needs to reboot or a failure occurs, the data on instance will be lost.

Therefore every 30 minutes of inactivity, the app will sleep and all the data will be lost. The next call to the API requesting to short a URL will return the short url id: "a" (the first entry on redis).