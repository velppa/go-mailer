# mailer

Simple transactional mailer middleware. Listens for POST HTTP requests using
webserver, sends emails using provider API upon request.

To send an email you need to send POST request to `/send` endpoint with data
being form-encoded (header `application/x-www-form-urlencoded`).

### Features

* Multiple transactional mail providers
    * Mandrill
    * Mailgun
    * SpartPost
    * TODO: SendGrid

* [MJML] templates support (if `mjml` app is available)

* Simplest API possible

* Async send

### Authentication

Authentication is done via tokens, configured in configuration file. Token should be sent in HTTP Header:

    Authorization: <secret-token>

### Data

Data is provided as JSON document of the following format:


    {
        "subject": "Hello message",
        "text": "Hello, world!",
        "html": "<h1>Hello, world!</h1>",
        "mjml": "<mjml>Some mjml markup</mjml>",
        "from": {
            "name": "pavel",
            "address": "schmooser@gmail.com"
        },
        "to": [
            {
                "name": "pavel",
                "address": "schmooser+to@gmail.com"
            }
        ],
        "cc": [
            {
                "name": "pavel",
                "address": "schmooser+cc@gmail.com"
            }
        ],
        "bcc": [
            {
                "name": "pavel",
                "address": "schmooser+bcc@gmail.com"
            }
        ]
    }

*Note*, that if `html` parameter is provided, then `mjml` parameter is ignored.
If both `text` and `html` (or `mjml`) parameters are provided then resulting
email will have both html and text parts (what part will be visible by user
depends on user mail client software).


### Return codes

Application returns result in JSON encoded format.

Structure of response:

    {"Status": "<status>", "Message": "<message>"}

* `Status` - can be either `ok`, or `error`. If `error` then corresponding
  non-200 HTTP return code will be returned
* `Message` - string, explaining status message


### Requirements

1. Go runtime
2. [MJML] when using mjml templates

### Example

1. Configure `config.toml` file and set up transactional mail provider. See
   `config-example.toml` for example.

2. Run it via docker:

    docker run -d --name mailer -p 1233:1233 -v /path/to/config/toml:/config.toml velppa/go-mailer

3. Send test email

    curl -X POST -H "Authorization: token" -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
        "subject": "Hello from Mailer app",
        "text": "Hello, world!",
        "html": "",
        "from": {
            "name": "Mailer App",
            "address": "some@one.com"
        },
        "to": [
            {
                "name": "recipient",
                "address": "hellomailer1234@mailinator.com"
            }
        ]
    }' "http://your.domain.com:1233/send"


[MJML]: http://mjml.io
